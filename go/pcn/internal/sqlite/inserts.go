package sqlite

const (
	InsertNewEntry = `
	INSERT INTO node_list_entries(msList, commitId, msIA) VALUES (?,?,?)
	`
)
