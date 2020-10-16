package sqlite

const (
	UpdateNodeListEntry = `
	Update node_list_entries SET 
	msList = ?,
	commitId = ?,
	timestamp = ?
	WHERE msIA = ?
	`
)
