package sqlite

const (
	InsertPLNEntry = `
	INSERT INTO pln_entries(ia) VALUES(?)
	`
)
