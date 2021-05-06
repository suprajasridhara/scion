// Copyright 2021 ETH Zurich
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pgncomm

import (
	"context"
	"database/sql"
	"math/rand"
	"os/exec"
	"strings"
	"time"
	"unsafe"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/ctrl"
	"github.com/scionproto/scion/go/lib/ctrl/ms_mgmt"
	"github.com/scionproto/scion/go/lib/ctrl/pgn_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/ms/internal/mscrypto"
	"github.com/scionproto/scion/go/ms/internal/msmsgr"
	"github.com/scionproto/scion/go/ms/internal/pgnentryhelper"
	"github.com/scionproto/scion/go/ms/internal/sqlite"
	"github.com/scionproto/scion/go/ms/internal/validator"
	"github.com/scionproto/scion/go/ms/plncomm"
	"github.com/scionproto/scion/go/pkg/trust"
	"github.com/scionproto/scion/go/proto"
)

const (
	MS_LIST = "MS_LIST"
)

/*SendSignedList The MS sends its signed mapping list to PGNs periodically. The time interval can
be specified in the config file during service startups. This time interval should not be more
than the global value for validity of MS lists, otherwise the lists in the PGNs will be stale
and invalid.

To push the list the MS performs the following operations:
- get the AS entries in the new_entries table
- fetch the PGN list from the configured PLN
- pick a random PGN from the list
- form the ms_mgmt.SignedMSList payload and pgn_mgmt.AddPGNEntryRequest
- send the AddPGNEntryRequest to the PGN that was picked using the messenger instance stored in
	msmsgr
*/

func SendSignedList(ctx context.Context, interval time.Duration) {
	logger := log.FromCtx(ctx)
	err := pushSignedPrefix(ctx)
	if err != nil {
		logger.Error("Error in pushSignedPrefix. Retry in next time interval", "Error: ", err)
	}
	pushTicker := time.NewTicker(interval * time.Minute)
	for {
		select {
		case <-pushTicker.C:
			err := pushSignedPrefix(ctx)
			if err != nil {
				logger.Error("Error in pushSignedPrefix. Retry in next time interval",
					"Error: ", err)
			}
		}
	}
}

func pushSignedPrefix(ctx context.Context) error {
	logger := log.FromCtx(ctx)

	//signed ASMapEntries in the form of SignedPld
	asEntries, err := sqlite.Db.GetNewEntries(context.Background())
	if err != nil {
		logger.Error("Could not get entries from DB", "Err: ", err)
		return err
	}

	signer, err := registerSigner()
	if err != nil {
		logger.Error("Error getting signer", "Err: ", err)
		return err
	}

	entries := []ms_mgmt.SignedASEntry{}
	for _, asEntry := range asEntries {
		entry := ms_mgmt.NewSignedASEntry(asEntry.Blob, asEntry.Sign)
		entries = append(entries, *entry)
	}

	timestamp := time.Now()

	pgn := getRandomPGN(context.Background())
	address := &snet.SVCAddr{IA: pgn.PGNIA, SVC: addr.SvcPGN}
	signedList := ms_mgmt.NewSignedMSList(uint64(timestamp.Unix()), entries, msmsgr.IA.String())
	msMgmtPld, err := ms_mgmt.NewPld(1, signedList)
	if err != nil {
		logger.Error("Error gorming msMgmtPld ", "Err: ", err)
		return err
	}
	pld, _ := ctrl.NewPld(msMgmtPld, &ctrl.Data{ReqId: rand.Uint64()})
	signedEntry, err := pld.SignedPld(context.Background(), signer)
	if err != nil {
		logger.Error("Error creating SignedPld", "Err: ", err)
		return err
	}
	entry, err := proto.PackRoot(signedEntry)

	if err != nil {
		logger.Error("Error packing signedEntry", "Err: ", err)
		return err
	}
	//For now push full lists, with empty commitID. This will be changed in the next iteration to
	//only push updates
	timestampPGN := time.Now().Add(msmsgr.MSListValidTime)

	req := pgn_mgmt.NewAddPGNEntryRequest(entry, MS_LIST, "", pgn.PGNId,
		uint64(timestampPGN.Unix()), msmsgr.IA.String())

	pgn_pld, err := pgn_mgmt.NewPld(1, req)
	if err != nil {
		logger.Error("Error forming pgn_mgmt payload", "Err: ", err)
		return err
	}

	reply, err := msmsgr.Msgr.SendPGNMessage(ctx, pgn_pld, address,
		rand.Uint64(), infra.AddPGNEntryRequest)

	if err != nil {
		logger.Error("Error getting reply from PGN", "Err: ", err)
		return err
	}

	//Validate PGN signature
	e := mscrypto.MSEngine{Msgr: msmsgr.Msgr, IA: msmsgr.IA}
	verifier := trust.Verifier{BoundIA: address.IA, Engine: e}
	err = verifier.Verify(ctx, reply.Blob, reply.Sign)

	if err != nil {
		logger.Error("Error verifying sign for PGN rep", "Err: ", err)
		return err
	}

	//persist reply
	packed, err := proto.PackRoot(reply)
	if err != nil {
		logger.Error("Error packing reply", "Err: ", err)
		return err
	}
	_, err = sqlite.Db.InsertPGNRep(context.Background(), packed)

	if err != nil {
		logger.Error("Error persisting PGN rep", "Err: ", err)
		return err
	}

	return nil

}

func PullAllPGNEntries(ctx context.Context, interval time.Duration) {
	PullPGNEntryByQuery(ctx, MS_LIST, "")
	pushTicker := time.NewTicker(interval)
	for {
		select {
		case <-pushTicker.C:
			PullPGNEntryByQuery(ctx, MS_LIST, "") //"" is considered wildcard.
		}
	}
}

func PullPGNEntryByQuery(ctx context.Context, entryType string, srcIA string) (int64, error) {
	// pgnEntryRequest := pgn_mgmt.NewPGNEntryRequest(entryType, srcIA)
	// pgn := getRandomPGN(context.Background())
	pgn := plncomm.PGN{PGNId: "pgn1-ff00:0:110", PGNIA: msmsgr.IA}
	signer, err := registerSigner()
	if err != nil {
		log.Error("Error registering signer", "Err: ", err)
		return 0, err
	}
	// pgn_pld, err := pgn_mgmt.NewPld(1, pgnEntryRequest)
	// if err != nil {
	// 	log.Error("Error forming pgn_mgmt payload", "Err: ", err)
	// 	return err
	// }
	// address := &snet.SVCAddr{IA: pgn.PGNIA, SVC: addr.SvcPGN}

	// reply, err := msmsgr.Msgr.SendPGNMessage(ctx, pgn_pld, address,
	// 	rand.Uint64(), infra.PGNEntryRequest)

	// if err != nil {
	// 	log.Error("Error getting reply from PGN", "Err: ", err)
	// 	return err
	// }

	//Validate PGN signature
	// e := mscrypto.MSEngine{Msgr: msmsgr.Msgr, IA: msmsgr.IA}
	// verifier := trust.Verifier{BoundIA: address.IA, Engine: e}
	// err = verifier.Verify(ctx, reply.Blob, reply.Sign)

	// if err != nil {
	// 	log.Error("Error verifying sign for PGN rep", "Err: ", err)
	// 	return err
	// }

	// pld := &ctrl.Pld{}
	// if err := proto.ParseFromRaw(pld, reply.Blob); err != nil {
	// 	log.Error("Error parsing PGN reply", "Err: ", err)
	// 	return err
	// }

	//Performance test -------

	//create AS MAP entry
	ts := uint64(time.Now().UnixNano())
	asEntry := ms_mgmt.NewASMapEntry([]string{"10.71.57.0/26"}, msmsgr.IA.String(), ts, "add_as_entry")
	mspld, err := ms_mgmt.NewPld(1, asEntry)
	pld, _ := ctrl.NewPld(mspld, &ctrl.Data{ReqId: rand.Uint64()})
	signedPld, err := pld.SignedPld(ctx, signer)
	asEntryBytes, err := proto.PackRoot(signedPld)
	log.Info("size of asEntry ", "size ", asEntryBytes.Len())
	var asEntries []*ctrl.SignedPld
	for i := 0; i < msmsgr.NoOfASEntries; i++ {
		asEntries = append(asEntries, signedPld)
	}
	entries := []ms_mgmt.SignedASEntry{}
	for _, asEntry := range asEntries {
		entry := ms_mgmt.NewSignedASEntry(asEntry.Blob, asEntry.Sign)
		entries = append(entries, *entry)
	}

	//create signed MS list
	timestamp := time.Now()

	signedList := ms_mgmt.NewSignedMSList(uint64(timestamp.Unix()), entries, msmsgr.IA.String())
	msMgmtPld, err := ms_mgmt.NewPld(1, signedList)
	if err != nil {
		log.Error("Error gorming msMgmtPld ", "Err: ", err)
		return 0, err
	}
	pld, _ = ctrl.NewPld(msMgmtPld, &ctrl.Data{ReqId: rand.Uint64()})
	log.Info("sizeof pld ", "size ", unsafe.Sizeof(pld))
	signedEntry, err := pld.SignedPld(context.Background(), signer)
	if err != nil {
		log.Error("Error creating SignedPld", "Err: ", err)
		return 0, err
	}
	entry, err := proto.PackRoot(signedEntry)

	if err != nil {
		log.Error("Error packing signedEntry", "Err: ", err)
		return 0, err
	}

	//Create PGNEntry request
	//For now push full lists, with empty commitID. This will be changed in the next iteration to
	//only push updates
	timestampPGN := time.Now().Add(msmsgr.MSListValidTime)

	req := pgn_mgmt.NewAddPGNEntryRequest(entry, MS_LIST, "", pgn.PGNId,
		uint64(timestampPGN.Unix()), msmsgr.IA.String())
	pgnPld, _ := pgn_mgmt.NewPld(1, req)
	pld, _ = ctrl.NewPld(pgnPld, &ctrl.Data{ReqId: rand.Uint64()})
	signedPld, _ = pld.SignedPld(context.Background(), signer)
	l, _ := proto.PackRoot(signedPld)
	log.Info("size of l ", "size ", l.Len())

	// for _, l := range pld.Pgn.PGNList.L {
	//validate each entry here

	start := time.Now()
	r, err := validateAndGetPGNEntry(l)
	if err != nil {
		log.Error("Error validating PGNEntry ", "Error: ", err)
		//continue
	}
	//parse persist entry here
	if err := processEntry(r.Entry, r.Timestamp); err != nil {
		log.Error("Error processing entry ", "Error: ", err)
		//continue
	}
	//}

	duration := time.Since(start)
	log.Info("Time elapsed ReverseMappinglist", "duration ", duration.String())

	return duration.Milliseconds(), nil
}

func validateAndGetPGNEntry(l common.RawBytes) (*pgn_mgmt.AddPGNEntryRequest, error) {
	signedPGNEntry := &ctrl.SignedPld{}
	err := proto.ParseFromRaw(signedPGNEntry, l)
	if err != nil {
		log.Error("Error getting signedPGNEntry", "Error: ", err)
		return nil, err
	}
	pld := &ctrl.Pld{}
	if err := proto.ParseFromRaw(pld, signedPGNEntry.Blob); err != nil {
		return nil, err
	}
	r := pld.Pgn.AddPGNEntryRequest
	if r.EntryType != MS_LIST {
		log.Error("Received entry is not an MS_LIST ")
		return nil, serrors.New("Received entry is not an MS_LIST ")
	}
	if err := pgnentryhelper.ValidatePGNEntry(r, signedPGNEntry); err != nil {
		log.Error("Error validating PGN entry ", "Error: ", err)
		return nil, err
	}
	return r, nil
}

func processEntry(entry []byte, timestamp uint64) error {
	signedPld := &ctrl.SignedPld{}
	err := proto.ParseFromRaw(signedPld, entry)
	if err != nil {
		log.Error("Error parsing entry ", "Error: ", err)
		return err
	}
	pld := &ctrl.Pld{}
	if err := proto.ParseFromRaw(pld, signedPld.Blob); err != nil {
		log.Error("Error parsing entry ", "Error: ", err)
		return err
	}
	// var wg sync.WaitGroup
	// wg.Add(len(pld.Ms.SignedMSList.ASEntries))
	c := make(chan int, msmsgr.WorkerPoolSize)
	for i, asEntry := range pld.Ms.SignedMSList.ASEntries {
		c <- i
		go func(asEntry ms_mgmt.SignedASEntry, ch chan int) {
			defer log.HandlePanic()
			sigPld := &ctrl.Pld{}
			if asEntry.Blob == nil {
				log.Info("Empty AS entry ")
			}
			if err := proto.ParseFromRaw(sigPld, asEntry.Blob); err != nil {
				log.Error("Error parsing sigPld ", "Error: ", err)
			}

			if sigPld.Ms.AsActionReq.Action == "add_as_entry" {
				persistASEntry(*sigPld.Ms.AsActionReq, timestamp)
			} else if sigPld.Ms.AsActionReq.Action == "del_as_entry" {
				deleteASEntry(*sigPld.Ms.AsActionReq)
			}
			index := <-ch

			log.Info("Done ", "index ", index)

		}(asEntry, c)
	}
	if len(c) > 0 {
		log.Info("Waiting for all workers to finish ", "len ", len(c))
	}
	for len(c) > 0 {
		log.Info("Waiting ")
	}
	// wg.Wait()

	return nil
}

func deleteASEntry(asEntry ms_mgmt.ASMapEntry) {
	for _, ip := range asEntry.Ip {
		_, err := sqlite.Db.DeleteFullMapEntryByIPAndIA(context.Background(), ip, asEntry.Ia)
		if err != nil {
			log.Error("Error deleteing from row", "ip", ip, " ia ", asEntry.Ia, "Error: ", err)
		}
	}
}

func persistASEntry(asEntry ms_mgmt.ASMapEntry, timestamp uint64) {
	for _, ip := range asEntry.Ip {
		//validate RPKI signatures before presisting
		ia, _ := addr.IAFromString(asEntry.Ia)
		if err := validateRPKI(ia, ip); err != nil {
			log.Error("RPKI validation failed IA ", "ia ", ia.String(), "IP ", ip, "err ", err)
			continue
		}
		rows, err := sqlite.Db.GetFullMapEntryByIP(context.Background(), ip)
		if err != nil {
			log.Error("Error getting full map entries by ip", "Error: ", err)
			continue
		}
		newestRow := sqlite.FullMapRow{IA: sql.NullString{String: asEntry.Ia, Valid: true},
			IP: sql.NullString{String: ip, Valid: true}, Created: int(asEntry.Timestamp),
			ValidUntil: int(timestamp)}
		if len(rows) > 0 {
			//keep the newest entry for this IP
			for _, row := range rows {
				if row.Created > newestRow.Created {
					newestRow = row
				}
			}
		}
		_, err = sqlite.Db.DeleteFullMapEntryByIP(context.Background(), ip)
		if err != nil {
			log.Error("Error deleting full map entries by IP", "Error: ", err)
		}
		_, err = sqlite.Db.InsertFullMapEntry(context.Background(), newestRow)
		if err != nil {
			log.Error("Error inserting full map row", "Error: ", err)
		}
	}
}

func validateRPKI(ia addr.IA, ip string) error {

	cmd := exec.Command("/bin/sh", validator.Path, "1234", ip)

	op, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(err.Error())
		return err
	}

	if strings.TrimSpace(string(op)) != validator.EntryValid {
		//log.Error("Not valid mapping")
		//return serrors.New("Not a valid mapping")
		return nil
	}
	return nil
}

func registerSigner() (*trust.Signer, error) {
	mscrypt := &mscrypto.MSSigner{}
	mscrypt.Init(context.Background(), msmsgr.Msgr, msmsgr.IA, mscrypto.CfgDir)
	signer, err := mscrypt.SignerGen.Generate(context.Background())
	if err != nil {
		return nil, serrors.WrapStr("Unable to create signer to AddASMap", err)
	}
	msmsgr.Msgr.UpdateSigner(signer, []infra.MessageType{infra.AddPGNEntryRequest,
		infra.PGNEntryRequest})
	return &signer, nil
}

func getRandomPGN(ctx context.Context) plncomm.PGN {
	logger := log.FromCtx(ctx)
	pgns, err := plncomm.GetPLNList(context.Background())
	if err != nil {
		logger.Error("Error getting PGNs", "Err: ", err)
	}
	//pick a random pgn to send signed list to
	if len(pgns) != 0 {
		randomIndex := rand.Intn(len(pgns))
		return pgns[randomIndex]
	}
	return plncomm.PGN{}
}
