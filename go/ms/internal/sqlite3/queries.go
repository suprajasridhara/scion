package sqlite3

const (
	FullMap = `
	SELECT * FROM full_map
	`
	NewEntryById = `
	SELECT entry FROM new_entries where id = ?
	`
	NewEntries = `
	SELECT entry FROM new_entries
	`
)
