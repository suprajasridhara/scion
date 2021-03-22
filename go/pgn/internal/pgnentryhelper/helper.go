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

	"github.com/scionproto/scion/go/lib/ctrl/pgn_mgmt"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/pgn/internal/pgncrypto"
	"github.com/scionproto/scion/go/pgn/internal/sqlite"
)

//TODO (supraja): move this and make it configurable
const pgn_entry_valid_time = 1000000 * time.Minute

var PGNID string

func IsValidPGNEntry(pgnEntry *pgn_mgmt.AddPGNEntryRequest) (bool, error) {
	//validate the PGN entry
	timestamp := time.Unix(int64(pgnEntry.Timestamp), 0)
	if !timestamp.Add(pgn_entry_valid_time).After(time.Now()) {
		return false, serrors.New("Invalid or expired timestamp in PGNEntry")
	}

	if pgnEntry.PGNId != PGNID {
		return false, serrors.New("Invalid PGNID in PGNEntry")
	}
	return true, nil
}

func PersistEntry(entry *pgn_mgmt.AddPGNEntryRequest, e pgncrypto.PGNEngine) error {
	allEntries, err := sqlite.Db.GetAllEntries(context.Background())
	if err != nil {
		log.Error("error reading list from db", err)
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
			entry.CommitID, entry.SrcIA, entry.Timestamp, entry.EntryType)
		if err != nil {
			log.Error("Error updating entry ", "Err: ", err)
			return err
		}
	} else if insert {
		log.Info("Inserting  PGNEntry in DB. SrcIA: " + entry.SrcIA +
			" EntryType: " + entry.EntryType)
		_, err = sqlite.Db.InsertEntry(context.Background(), entry.Entry,
			entry.CommitID, entry.SrcIA, entry.Timestamp, entry.EntryType)
		if err != nil {
			log.Error("Error inserting entry ", "Err: ", err)
			return err
		}
	}
	return nil
}
