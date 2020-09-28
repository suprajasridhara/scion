package sqlite

const (
	Schema = `
	CREATE TABLE IF NOT EXISTS pln_entries(
		id INTEGER PRIMARY KEY,
		pcnId DATA NOT NULL,
		ia INTEGER NOT NULL,
		raw BLOB
		);
	`
)
