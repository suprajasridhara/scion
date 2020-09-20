package sqlite3

const (
	Schema = `
	CREATE TABLE IF NOT EXISTS full_map(
		id INTEGER NOT NULL,
		ip DATA NOT NULL,
		ia DATA NOT NULL,
		PRIMARY KEY (id)
	);

	CREATE TABLE IF NOT EXISTS new_entries(
		id INTEGER PRIMARY KEY,
		entry BLOB
	);

	CREATE TABLE IF NOT EXISTS pcn_reps(
		id INTEGER PRIMARY KEY,
		fullRep BLOB
	)
	`
)
