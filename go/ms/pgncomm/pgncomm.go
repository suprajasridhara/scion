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
	"math/rand"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl/ms_mgmt"
	"github.com/scionproto/scion/go/lib/ctrl/pgn_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/ms/internal/mscrypto"
	"github.com/scionproto/scion/go/ms/internal/msmsgr"
	"github.com/scionproto/scion/go/ms/internal/sqlite"
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

	_, err = registerSigner(infra.AddPGNEntryRequest)
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
	entry, err := proto.PackRoot(signedList)

	if err != nil {
		logger.Error("Error packing signedList", "Err: ", err)
		return err
	}
	//For now push full lists, with empty commitID. This will be changed in the next iteration to
	//only push updates
	req := pgn_mgmt.NewAddPGNEntryRequest(entry, MS_LIST, "", pgn.PGNId,
		uint64(timestamp.Unix()), msmsgr.IA.String())

	pld, err := pgn_mgmt.NewPld(1, req)
	if err != nil {
		logger.Error("Error forming ms_mgmt payload", "Err: ", err)
		return err
	}

	reply, err := msmsgr.Msgr.SendPGNMessage(ctx, pld, address,
		rand.Uint64(), infra.AddPGNEntryRequest)

	if err != nil {
		logger.Error("error getting reply from PCN", "Err: ", err)
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

func registerSigner(msgType infra.MessageType) (*trust.Signer, error) {
	mscrypt := &mscrypto.MSSigner{}
	mscrypt.Init(context.Background(), msmsgr.Msgr, msmsgr.IA, mscrypto.CfgDir)
	signer, err := mscrypt.SignerGen.Generate(context.Background())
	if err != nil {
		return nil, serrors.WrapStr("Unable to create signer to AddASMap", err)
	}
	msmsgr.Msgr.UpdateSigner(signer, []infra.MessageType{msgType})
	return &signer, nil
}

func getRandomPGN(ctx context.Context) plncomm.PGN {
	logger := log.FromCtx(ctx)
	pgns, err := plncomm.GetPLNList(context.Background())
	if err != nil {
		logger.Error("Error getting PGNs", "Err: ", err)
	}
	//pick a random pgn to send signed list to
	randomIndex := rand.Intn(len(pgns))
	return pgns[randomIndex]
}
