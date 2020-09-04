package sqlite

const (
	Schema = `
	CREATE TABLE IF NOT EXISTS ms_token(
		id INTEGER PRIMARY KEY,
		token BLOB
	);
	`
)
