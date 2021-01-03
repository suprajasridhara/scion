package sqlite

//PLNListEntry sql table structure to read from pln_entries
type PLNListEntry struct {
	ID    int    `TbField:"id"`
	PcnID string `TbField:"pcnID"`
	IA    int64  `TbField:"ia"`
	Raw   []byte `TbField:"raw"`
}
