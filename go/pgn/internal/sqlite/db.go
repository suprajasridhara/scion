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

package sqlite

import (
	"context"
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/infra/modules/db"
	"github.com/scionproto/scion/go/lib/serrors"
)

var Db *DB

type DB struct {
	db *sql.DB
	*executor
}

type executor struct {
	sync.RWMutex
	db db.Sqler
}

// New returns a new SQLite backend opening a database at the given path. If
// no database exists a new database is be created. If the schema version of the
// stored database is different from the one in schema.go, an error is returned.
func New(path string, schemaVersion int) error {
	db, err := db.NewSqlite(path, Schema, schemaVersion)
	if err != nil {
		return err
	}
	Db = NewFromDB(db)
	return nil
}

// NewFromDB returns a new backend from the given database.
func NewFromDB(db *sql.DB) *DB {
	return &DB{
		db: db,
		executor: &executor{
			db: db,
		},
	}
}

//Close closes the database connection
func (d *DB) Close() {
	d.db.Close()
}

//InsertEntry inserts a new row into pgn_entries
func (e *executor) InsertEntry(ctx context.Context, entry []byte,
	commitID string, srcIA string, timestamp uint64, entryType string,
	signedBlob []byte) (sql.Result, error) {

	res, err := e.db.ExecContext(ctx, InsertNewEntry, entry, commitID, srcIA,
		timestamp, entryType, signedBlob)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//InsertEmptyObject inserts a new row into pgn_entries
func (e *executor) InsertEmptyObject(ctx context.Context, emptyObject []byte,
	isd string, timestamp uint64) (sql.Result, error) {

	res, err := e.db.ExecContext(ctx, InsertEmptyObject, isd, timestamp, emptyObject)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//UpdateEntry updates a row in pgn_entries based on srcIA and entryType
func (e *executor) UpdateEntry(ctx context.Context, entry []byte,
	commitID string, srcIA string, timestamp uint64, entryType string,
	signedBlob []byte) (sql.Result, error) {

	res, err := e.db.ExecContext(ctx, UpdateEntry, entry, commitID,
		timestamp, signedBlob, srcIA, entryType)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//GetEntriesByTypeAndSrcIA queries pgn_entries by entryType and srcIA. entryType
//and srcIA are matched by pattern so can contain wildcards
func (e *executor) GetEntriesByTypeAndSrcIA(ctx context.Context,
	entryType string, srcIA string) ([]PGNEntry, error) {

	e.RLock()
	defer e.RUnlock()
	rows, err := e.db.QueryContext(ctx, EntriesByTypeAndSrcIA, entryType, srcIA)
	if err != nil {
		return nil, serrors.Wrap(db.ErrReadFailed, err)
	}
	defer rows.Close()
	got := []PGNEntry{}
	for rows.Next() {
		var r PGNEntry
		err = rows.Scan(&r.ID, &r.Entry, &r.CommitID, &r.SrcIA, &r.Timestamp,
			&r.EntryType, &r.SignedBlob)
		if err != nil {
			return nil, serrors.Wrap(db.ErrDataInvalid, err)
		}
		got = append(got, r)
	}
	return got, nil
}

//GetEmptyObjects queries empty_objects
func (e *executor) GetEmptyObjects(ctx context.Context) ([]common.RawBytes, error) {

	e.RLock()
	defer e.RUnlock()
	rows, err := e.db.QueryContext(ctx, EmptyObjects)
	if err != nil {
		return nil, serrors.Wrap(db.ErrReadFailed, err)
	}
	defer rows.Close()
	got := []common.RawBytes{}
	for rows.Next() {
		var r common.RawBytes
		err = rows.Scan(&r)
		if err != nil {
			return nil, serrors.Wrap(db.ErrDataInvalid, err)
		}
		got = append(got, r)
	}
	return got, nil
}

//GetEntrySRCIAs queries pgn_entries for all SRCIAs
func (e *executor) GetEntrySRCIAs(ctx context.Context) ([]string, error) {

	e.RLock()
	defer e.RUnlock()
	rows, err := e.db.QueryContext(ctx, EntrySRCIAs)
	if err != nil {
		return nil, serrors.Wrap(db.ErrReadFailed, err)
	}
	defer rows.Close()
	got := []string{}
	for rows.Next() {
		var r string
		err = rows.Scan(&r)
		if err != nil {
			return nil, serrors.Wrap(db.ErrDataInvalid, err)
		}
		got = append(got, r)
	}
	return got, nil
}
