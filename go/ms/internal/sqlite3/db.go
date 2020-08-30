package sqlite3

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// var Db *database

var Db *sql.DB

func NewDb(path string) {
	Db, _ = sql.Open("sqlite3", path)
	print("here")
}

func Init() {
	stmt, _ := Db.Prepare(Schema)
	stmt.Exec()
}
