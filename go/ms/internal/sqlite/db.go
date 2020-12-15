// Copyright 2020 ETH Zurich
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
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	"github.com/scionproto/scion/go/lib/ctrl"
	"github.com/scionproto/scion/go/lib/infra/modules/db"
	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/proto"
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

func (d *DB) Close() {
	d.db.Close()
}

//GetFullMap reads the full_map entries from the database and parses it
//to return a slice of FullMapRow objects
func (e *executor) GetFullMap(ctx context.Context) ([]FullMapRow, error) {
	e.RLock()
	defer e.RUnlock()
	rows, err := e.db.QueryContext(ctx, FullMap)
	if err != nil {
		return nil, serrors.Wrap(db.ErrReadFailed, err)
	}
	defer rows.Close()
	got := []FullMapRow{}
	for rows.Next() {
		var r FullMapRow
		err = rows.Scan(&r.ID, &r.IP, &r.IA, &r.Timestamp)
		if err != nil {
			return nil, serrors.Wrap(db.ErrDataInvalid, err)
		}
		got = append(got, r)
	}
	return got, nil
}

//GetFullMapEntryByIP queries the full_map table by IP
func (e *executor) GetFullMapEntryByIP(ctx context.Context, ip string) ([]FullMapRow, error) {
	e.RLock()
	defer e.RUnlock()
	rows, err := e.db.QueryContext(ctx, FullMapEntryByIP, ip)
	if err != nil {
		return nil, serrors.Wrap(db.ErrReadFailed, err)
	}
	defer rows.Close()
	got := []FullMapRow{}
	for rows.Next() {
		var r FullMapRow
		err = rows.Scan(&r.ID, &r.IP, &r.IA, &r.Timestamp)
		if err != nil {
			return nil, serrors.Wrap(db.ErrDataInvalid, err)
		}
		got = append(got, r)
	}
	return got, nil
}

//InsertFullMapEntry inserts a row into the full_map table
func (e *executor) InsertFullMapEntry(ctx context.Context, fmRow FullMapRow) (sql.Result, error) {
	//TODO (supraja): handle transaction correctly here
	res, err := e.db.ExecContext(ctx, InsFullMapEntry, fmRow.IP, fmRow.IA, fmRow.Timestamp)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//DeleteFullMapEntryByID deletes a row in the full_map table by id
func (e *executor) DeleteFullMapEntryByID(ctx context.Context, id int) (sql.Result, error) {

	//TODO (supraja): handle transaction correctly here
	res, err := e.db.ExecContext(ctx, DelFullMapEntry, id)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//InsertNewEntry inserts a row into the new_entries table
func (e *executor) InsertNewEntry(ctx context.Context, entry []byte) (sql.Result, error) {

	//TODO (supraja): handle transaction correctly here
	res, err := e.db.ExecContext(ctx, InsNewEntry, entry)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//InsertPCNRep inserts a row into the pcn_reps table
func (e *executor) InsertPCNRep(ctx context.Context, entry []byte) (sql.Result, error) {

	//TODO (supraja): handle transaction correctly here
	res, err := e.db.ExecContext(ctx, InsPCNRep, entry)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//GetNewEntryByID queries the new_entries table by id to return the signedPld stored in it
func (e *executor) GetNewEntryByID(ctx context.Context, id int) (*ctrl.SignedPld, error) {
	e.RLock()
	defer e.RUnlock()
	rows, err := e.db.QueryContext(ctx, NewEntryByID, id)
	if err != nil {
		return nil, serrors.Wrap(db.ErrReadFailed, err)
	}
	defer rows.Close()

	rawResult, err := getRawResultFromRows(rows)
	if err != nil {
		return nil, err
	}
	if len(rawResult) != 1 {
		return nil, serrors.Wrap(db.ErrDataInvalid, err)
	}

	//can use index 0, we expect only one entry to be returned as id is the primary key
	got := &ctrl.SignedPld{}
	proto.ParseFromRaw(got, rawResult[0])
	return got, nil
}

//GetNewEntries returns all rows of the new_entries table
func (e *executor) GetNewEntries(ctx context.Context) ([]*ctrl.SignedPld, error) {
	e.RLock()
	defer e.RUnlock()
	rows, err := e.db.QueryContext(ctx, NewEntries)
	if err != nil {
		return nil, serrors.Wrap(db.ErrReadFailed, err)
	}
	defer rows.Close()

	rawResult, err := getRawResultFromRows(rows)

	if err != nil {
		return nil, err
	}

	l := []*ctrl.SignedPld{}

	for _, rawResult := range rawResult {
		got := &ctrl.SignedPld{}
		proto.ParseFromRaw(got, rawResult)
		l = append(l, got)
	}
	return l, nil
}

func getRawResultFromRows(rows *sql.Rows) ([][]byte, error) {
	cols, err := rows.Columns()
	if err != nil {
		fmt.Println("Failed to get columns", err)
		return nil, err
	}

	rawResult := make([][]byte, len(cols))
	dest := make([]interface{}, len(cols)) // A temporary interface{} slice

	for i := range rawResult {
		dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	}

	for rows.Next() {
		err = rows.Scan(dest...)
		if err != nil {
			fmt.Println("Failed to scan row", err)
			return nil, err
		}
	}
	return rawResult, err
}
