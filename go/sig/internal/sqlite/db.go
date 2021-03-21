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

func (d *DB) Close() {
	d.db.Close()
}

//InsertNewMSToken inserts a new MSToken into ms_token table
func (e *executor) InsertNewMSToken(ctx context.Context, entry []byte) (sql.Result, error) {
	res, err := e.db.ExecContext(ctx, InsertNewMSToken, entry)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//InsertNewPushedPrefix inserts a prefix into pushed_prefixes table
func (e *executor) InsertNewPushedPrefix(ctx context.Context, prefix string) (sql.Result, error) {
	res, err := e.db.ExecContext(ctx, InsertNewPushedPrefix, prefix)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//GetPushedPrefixes returns a list of prefixes from pushed_prefixes table
func (e *executor) GetPushedPrefixes(ctx context.Context) ([]string, error) {
	e.RLock()
	defer e.RUnlock()
	rows, err := e.db.QueryContext(ctx, PushedPrefixes)
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
