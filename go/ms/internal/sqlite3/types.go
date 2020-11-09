package sqlite3

import "database/sql"

// FullMapRow sql table struct that to store into mysql
type FullMapRow struct {
	Id        int            `TbField:"id"`
	IP        sql.NullString `TbField:"ip"`
	IA        sql.NullString `TbField:"ia"`
	Timestamp int            `TbField:"timestamp"`
}

type NewEntry struct {
	entry *[]byte `TbField:"entry"`
}
