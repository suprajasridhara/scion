package sqlite

import "database/sql"

// NodeListEntry : sql table struct that to read from sqlite
type NodeListEntry struct {
	Id       int            `TbField:"id"`
	MsList   *[]byte        `TbField:"msList"`
	CommitId sql.NullString `TbField:"commitId"`
}
