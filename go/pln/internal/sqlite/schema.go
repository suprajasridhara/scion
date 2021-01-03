package sqlite

const (
	/*Schema PLN database contains the following tables:
	- pln_entries: these are entries of the PCNs that the PLN has discovered through gossip
	*/
	Schema = `
	CREATE TABLE IF NOT EXISTS pln_entries(
		id INTEGER PRIMARY KEY,
		pcnID DATA NOT NULL,
		ia INTEGER NOT NULL,
		raw BLOB
		);
	`
)
