package sqlite3

const (
	FullMap = `
	SELECT * FROM full_map
	`

	FullMapEntryByIP = `
	SELECT * from full_map where ip = ?
	`
	NewEntryById = `
	SELECT entry FROM new_entries where id = ?
	`
	NewEntries = `
	SELECT entry FROM new_entries
	`
)
