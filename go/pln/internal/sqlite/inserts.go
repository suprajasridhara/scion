package sqlite

const (
	//InsertPLNEntry is the query to insert a new row into pln_entries
	InsertPLNEntry = `
	INSERT INTO pln_entries(pcnID, ia, raw) VALUES(?,?,?)
	`
)
