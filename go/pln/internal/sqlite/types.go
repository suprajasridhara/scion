package sqlite

// FullMap : sql table struct that to store into mysql
type PlnListEntry struct {
	Id int   `TbField:"id"`
	I  int64 `TbField:"i"`
	A  int64 `TbField:"a"`
}
