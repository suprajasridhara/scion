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

//InsertNewNodeListEntry inserts a new row into node_list_entries
func (e *executor) InsertNewNodeListEntry(ctx context.Context, entry []byte,
	commitID string, msIA string, timestamp uint64) (sql.Result, error) {
	res, err := e.db.ExecContext(ctx, InsertNewEntry, entry, commitID, msIA, timestamp)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//UpdateNodeListEntry updates a row in node_list_entries based on msIA
func (e *executor) UpdateNodeListEntry(ctx context.Context, entry []byte,
	commitID string, msIA string, timestamp uint64) (sql.Result, error) {
	res, err := e.db.ExecContext(ctx, UpdateNodeListEntry, entry, commitID, timestamp, msIA)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//GetFullNodeList fetches all rows in node_list_entries
func (e *executor) GetFullNodeList(ctx context.Context) ([]NodeListEntry, error) {
	e.RLock()
	defer e.RUnlock()
	rows, err := e.db.QueryContext(ctx, FullNodeList)
	if err != nil {
		return nil, serrors.Wrap(db.ErrReadFailed, err)
	}
	defer rows.Close()
	got := []NodeListEntry{}
	for rows.Next() {
		var r NodeListEntry
		err = rows.Scan(&r.ID, &r.MsList, &r.CommitID, &r.MSIA, &r.Timestamp)
		if err != nil {
			return nil, serrors.Wrap(db.ErrDataInvalid, err)
		}
		got = append(got, r)
	}
	return got, nil
}
