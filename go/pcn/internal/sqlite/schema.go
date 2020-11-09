package sqlite

const (
	/*Schema PCN uses the following tables:
	- node_list_entries to store the mapping lists it recives from mapping
	servics and other PCNs through gossip*/
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
