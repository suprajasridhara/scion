package sqlite3

const (
	/*Schema for MS:
	- full_map - stores the processed mapping entries that the MS
		pulls from the Publishing Infrastructured
	- new_entries - stores new mappings that MS recives from SIGs
		to be pushed to the Publishing Infrastructure
	- pcn_reps - stores response tokens from the publishing
		infrastructure.
	*/
	Schema = `
	CREATE TABLE IF NOT EXISTS full_map(
		id INTEGER NOT NULL,
		ip DATA NOT NULL,
		ia DATA NOT NULL,
		timestamp INTEGER NOT NULL,
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
