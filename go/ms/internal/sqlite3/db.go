package sqlite3

import (
	"context"
	"database/sql"
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

// func InsertDummyData() {
// 	stmt, err := db.Prepare(InsertDummyData1)
// 	if err != nil {
// 		log.Error(err.Error())
// 	}
// 	stmt.Exec()

// 	stmt, err = db.Prepare(InsertDummyData2)
// 	if err != nil {
// 		log.Error(err.Error())
// 	}
// 	stmt.Exec()
// }

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
		err = rows.Scan(&r.Id, &r.IP, &r.IA)
		if err != nil {
			return nil, serrors.Wrap(db.ErrDataInvalid, err)
		}
		got = append(got, r)
	}
	return got, nil
}

func (e *executor) InsertNewEntry(ctx context.Context, entry []byte) (sql.Result, error) {

	//TODO (supraja): handle transaction correctly here

	// strings.Replace(InsertNewEntry, "$1", entry, -1)
	res, err := e.db.ExecContext(ctx, InsertNewEntry, entry)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (e *executor) GetNewEntryById(ctx context.Context, id int) (*ctrl.SignedPld, error) {
	e.RLock()
	defer e.RUnlock()
	rows, err := e.db.QueryContext(ctx, NewEntryById, id)
	if err != nil {
		return &ctrl.SignedPld{}, serrors.Wrap(db.ErrReadFailed, err)
	}
	defer rows.Close()
	got := &ctrl.SignedPld{}
	for rows.Next() {
		var r NewEntry
		err = rows.Scan(&r.entry)
		if err != nil {
			return &ctrl.SignedPld{}, serrors.Wrap(db.ErrDataInvalid, err)
		}
		proto.ParseFromRaw(got, *r.entry)
	}

	return got, nil
}
