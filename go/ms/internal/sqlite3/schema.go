package sqlite3

const (
	Schema = `
	CREATE TABLE IF NOT EXISTS full_map(
		id INTEGER NOT NULL,
		ip INTEGER NOT NULL,
		ia DATA NOT NULL,
		PRIMARY KEY (id)
	);
	`
)
