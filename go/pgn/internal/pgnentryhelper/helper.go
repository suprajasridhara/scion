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
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/ctrl"
	"github.com/scionproto/scion/go/lib/ctrl/pgn_mgmt"
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
