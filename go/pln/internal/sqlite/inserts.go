package sqlite

const (
	InsertPLNEntry = `
	INSERT INTO pln_entries(pcnId, ia) VALUES(?,?)
	`
)
