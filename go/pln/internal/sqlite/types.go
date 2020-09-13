package sqlite

// FullMap : sql table struct that to store into mysql
type PlnListEntry struct {
	Id    int    `TbField:"id"`
	PcnId string `TbField:"pcnId"`
	IA    int64  `TbField:"ia"`
}
