package sqlite

const (
	/*Schema PLN database contains the following tables:
	- pln_entries: these are entries of the pcns that the pln has discovered through gossip
	*/
	Schema = `
	CREATE TABLE IF NOT EXISTS pln_entries(
		id INTEGER PRIMARY KEY,
		pcnId DATA NOT NULL,
		ia INTEGER NOT NULL,
		raw BLOB
		);
	`
)
