package sqlite

const (
	Schema = `

	CREATE TABLE IF NOT EXISTS node_list_entries(
		id INTEGER PRIMARY KEY,
		msList BLOB NOT NULL,
		commitId DATA NOT NULL,
		msIA INTEGER DATA NOT NULL,
		timestamp INTEGER NOT NULL
	);	

	`
)
