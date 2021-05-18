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

package pgnentryhelper

import (
	"context"
	"math/rand"
	"strconv"
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
	"github.com/scionproto/scion/go/pgn/internal/pgncrypto"
	"github.com/scionproto/scion/go/pgn/internal/pgnmsgr"
	"github.com/scionproto/scion/go/pgn/internal/sqlite"
	"github.com/scionproto/scion/go/pkg/trust"
	"github.com/scionproto/scion/go/proto"
)

var PGNID string

/*ValidatePGNEntry validates the PGN entry along with its signatures.
To validate signatures:
It first validates the IA signature on signedPld based on the srcIA in pgnEntry,
it then validates the signature from the same IA on pgnEntry.entry.
This ensures that the timestamp in pgnEntry and the entry in that entry
were both created by the same IA.
After signature validation is complete, it checks that the entry is not stale
i.e. has a valid timestamp.
And then if the checkPGNID flag is set, ensures that the current PGNID and the
pgnId in the pgnEntry are equal.
Params:
- pgnEntry: parsed pgnEntry to validate
- signedPld: ctrl.SignedPld with the pgnEntry. Mostly received from another
			PGN or a other services
- checkPGNID: if set, the function checks if pgnEntry.PGNId is equal to the current PGNID
**/
func ValidatePGNEntry(pgnEntry *pgn_mgmt.AddPGNEntryRequest,
	signedPld *ctrl.SignedPld, checkPGNID bool) error {

	if err := validatePGNEntrySignatures(pgnEntry, signedPld); err != nil {
		return err
	}
	//validate the PGN entry
	timestamp := time.Unix(int64(pgnEntry.Timestamp), 0) //the entry is valid till this time
	if !timestamp.After(time.Now()) {                    //check if the entry has expired
		return serrors.New("Invalid or expired timestamp in PGNEntry")
	}

	//validate PGNId only if checkPGNID is true
	if checkPGNID && pgnEntry.PGNId != PGNID {
		return serrors.New("Invalid PGNID in PGNEntry")
	}
	return nil
}

func validatePGNEntrySignatures(pgnEntry *pgn_mgmt.AddPGNEntryRequest,
	signedPld *ctrl.SignedPld) error {

	ia, err := addr.IAFromString(pgnEntry.SrcIA)
	if err != nil {
		return err
	}
	if err := VerifyASSignature(context.Background(), signedPld, ia); err != nil {
		return err
	}

	signedEntry := &ctrl.SignedPld{}
	proto.ParseFromRaw(signedEntry, pgnEntry.Entry)

	if err := VerifyASSignature(context.Background(), signedEntry, ia); err != nil {
		return err
	}
	return nil
}

func VerifyASSignature(ctx context.Context, message *ctrl.SignedPld, IA addr.IA) error {
	//Verify AS signature
	e := pgncrypto.PGNEngine{Msgr: pgnmsgr.Msgr, IA: pgnmsgr.IA}
	verifier := trust.Verifier{BoundIA: IA, Engine: e}
	return verifier.Verify(ctx, message.Blob, message.Sign)
}

func PersistEntry(entry *pgn_mgmt.AddPGNEntryRequest, e pgncrypto.PGNEngine,
	signedBlob []byte) error {

	allEntries, err := sqlite.Db.GetEntriesByTypeAndSrcIA(context.Background(), "%", "%")
	if err != nil {
		log.Error("error reading list from db", "err", err)
		return err
	}
	insert := true
	update := false
	//check if entry exists with the same commit ID, replace only if not
	for _, dbEntry := range allEntries {
		if dbEntry.SrcIA.String == entry.SrcIA &&
			dbEntry.EntryType.String == entry.EntryType &&
			uint64(dbEntry.Timestamp) > entry.Timestamp {
			//if srcIa and entryType are the same and the timestamp in the db is
			//newer then there is no need to update it
			insert = false
			break
		} else if dbEntry.SrcIA.String == entry.SrcIA &&
			dbEntry.EntryType.String == entry.EntryType {
			update = true
			break
		}
	}
	if update {
		log.Info("Updating PGNEntry in DB. SrcIA: " + entry.SrcIA +
			" EntryType: " + entry.EntryType)
		_, err = sqlite.Db.UpdateEntry(context.Background(), entry.Entry,
			entry.CommitID, entry.SrcIA, entry.Timestamp, entry.EntryType, signedBlob)
		if err != nil {
			log.Error("Error updating entry ", "Err: ", err)
			return err
		}
	} else if insert {
		log.Info("Inserting  PGNEntry in DB. SrcIA: " + entry.SrcIA +
			" EntryType: " + entry.EntryType)
		_, err = sqlite.Db.InsertEntry(context.Background(), entry.Entry,
			entry.CommitID, entry.SrcIA, entry.Timestamp, entry.EntryType, signedBlob)
		if err != nil {
			log.Error("Error inserting entry ", "Err: ", err)
			return err
		}
	}
	return nil
}

func GetEmptyObjects(isds map[uint64]bool, signer *trust.Signer) []common.RawBytes {

	emptyObjects, err := sqlite.Db.GetEmptyObjects(context.Background())
	if err != nil {
		log.Error("Error getting empty objects", "Error: ", err)
	}
	pgnRange := strings.Split(pgnmsgr.ISDRange, "-")
	start, _ := strconv.Atoi(pgnRange[0])
	end, _ := strconv.Atoi(pgnRange[1])

	for i := uint64(start); i <= uint64(end); i++ {
		if !isds[i] {
			//no list for isd i. Create empty token
			empty := pgn_mgmt.NewEmptyObject(strconv.Itoa(int(i)), uint64(time.Now().Unix()))
			e, _ := pgn_mgmt.NewPld(1, empty)
			pld, _ := ctrl.NewPld(e, &ctrl.Data{ReqId: rand.Uint64()})
			spld, _ := pld.SignedPld(context.Background(), signer)
			b, _ := proto.PackRoot(spld)
			emptyObjects = append(emptyObjects, b)
		}
	}
	return emptyObjects
}

func GetISDsInEntries(dbEntries []sqlite.PGNEntry) map[uint64]bool {
	set := make(map[uint64]bool)
	for _, entry := range dbEntries {
		ia, _ := addr.IAFromString(entry.SrcIA.String)
		set[uint64(ia.I)] = true
	}
	return set
}

func CreateObjMS() (*pgn_mgmt.AddPGNEntryRequest, *ctrl.SignedPld) {
	ctx := context.Background()
	signer, err := registerSigner()
	if err != nil {
		log.Error("Error getting signer ", "Err: ", err)
		return nil, nil
	}
	ts := uint64(time.Now().UnixNano())
	asEntry := ms_mgmt.NewASMapEntry([]string{"10.71.57.0/26"}, pgnmsgr.IA.String(), ts, "add_as_entry")
	mspld, err := ms_mgmt.NewPld(1, asEntry)
	pld, _ := ctrl.NewPld(mspld, &ctrl.Data{ReqId: rand.Uint64()})
	signedPld, err := pld.SignedPld(ctx, signer)
	asEntryBytes, err := proto.PackRoot(signedPld)
	log.Info("size of asEntry ", "size ", asEntryBytes.Len())
	var asEntries []*ctrl.SignedPld
	noOfASEntries := 1000
	log.Info("No of AAAS entries ", "entriess ", noOfASEntries)
	for i := 0; i < noOfASEntries; i++ {
		asEntries = append(asEntries, signedPld)
	}
	entries := []ms_mgmt.SignedASEntry{}
	for _, asEntry := range asEntries {
		entry := ms_mgmt.NewSignedASEntry(asEntry.Blob, asEntry.Sign)
		//e, _ := proto.PackRoot(entry)
		//log.Info("sizeof signed AS entry ", "size ", e.Len())
		entries = append(entries, *entry)
	}
	//create signed MS list
	timestamp := time.Now()

	signedList := ms_mgmt.NewSignedMSList(uint64(timestamp.Unix()), entries, pgnmsgr.IA.String())
	msMgmtPld, err := ms_mgmt.NewPld(1, signedList)
	if err != nil {
		log.Error("Error gorming msMgmtPld ", "Err: ", err)
		return nil, nil
	}
	pld, _ = ctrl.NewPld(msMgmtPld, &ctrl.Data{ReqId: rand.Uint64()})
	log.Info("sizeof pld ", "size ", unsafe.Sizeof(pld))
	signedEntry, err := pld.SignedPld(context.Background(), signer)
	if err != nil {
		log.Error("Error creating SignedPld", "Err: ", err)
		return nil, nil
	}
	entry, err := proto.PackRoot(signedEntry)

	if err != nil {
		log.Error("Error packing signedEntry", "Err: ", err)
		return nil, nil
	}
	//For now push full lists, with empty commitID. This will be changed in the next iteration to
	//only push updates
	timestampPGN := time.Now().Add(1000 * time.Hour)

	req := pgn_mgmt.NewAddPGNEntryRequest(entry, "MS_LIST", "", PGNID,
		uint64(timestampPGN.Unix()), pgnmsgr.IA.String())

	pgnPld, err := pgn_mgmt.NewPld(1, req)
	if err != nil {
		log.Error("Error forming pgn_mgmt payload", "Err: ", err)
		return nil, nil
	}
	pld, _ = ctrl.NewPld(pgnPld, &ctrl.Data{ReqId: rand.Uint64()})
	signedPlnPLd, _ := pld.SignedPld(context.Background(), signer)
	return req, signedPlnPLd
}

func registerSigner() (*trust.Signer, error) {
	pgncrypt := &pgncrypto.PGNSigner{}
	pgncrypt.Init(context.Background(), pgnmsgr.Msgr, pgnmsgr.IA, pgncrypto.CfgDir)
	signer, err := pgncrypt.SignerGen.Generate(context.Background())
	if err != nil {
		return nil, serrors.WrapStr("Unable to create signer to AddASMap", err)
	}
	pgnmsgr.Msgr.UpdateSigner(signer, []infra.MessageType{infra.AddPGNEntryRequest,
		infra.PGNEntryRequest})
	return &signer, nil
}
