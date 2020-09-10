package sqlite

const (
	Schema = `
	CREATE TABLE IF NOT EXISTS ms_token(
		id INTEGER PRIMARY KEY,
		token BLOB
	);
	CREATE TABLE IF NOT EXISTS pushed_prefixes(
		id INTEGER PRIMARY KEY,
		prefix DATA NOT NULL
	);
	`
)
