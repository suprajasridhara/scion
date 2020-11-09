package sqlite

//PlnListEntry sql table structure to read from pln_entries
type PlnListEntry struct {
	Id    int    `TbField:"id"`
	PcnId string `TbField:"pcnId"`
	IA    int64  `TbField:"ia"`
	Raw   []byte `TbField:"raw"`
}
