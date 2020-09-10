package sqlite

const (
	Schema = `
	CREATE TABLE IF NOT EXISTS pln_entries(
		id INTEGER PRIMARY KEY,
		a INTEGER NOT NULL,
		i INTEGER NOT NULL
	);
	`
)
