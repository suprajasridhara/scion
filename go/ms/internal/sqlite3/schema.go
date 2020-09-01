package sqlite3

const (
	Schema = `
	CREATE TABLE IF NOT EXISTS full_map(
		id INTEGER NOT NULL,
		ip DATA NOT NULL,
		ia DATA NOT NULL,
		PRIMARY KEY (id)
	);
	`
)
