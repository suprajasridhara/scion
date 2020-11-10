package sqlite

import "database/sql"

//FullMapRow sql table struct that to store into and fetch from full_map table
type FullMapRow struct {
	Id        int            `TbField:"id"`
	IP        sql.NullString `TbField:"ip"`
	IA        sql.NullString `TbField:"ia"`
	Timestamp int            `TbField:"timestamp"`
}

//NewEntry sql table struct that to store into and fetch from new_entries table
type NewEntry struct {
	entry *[]byte `TbField:"entry"`
}
