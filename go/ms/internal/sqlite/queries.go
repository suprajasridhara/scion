package sqlite

const (
	//FullMap queries for all rows in the full_map table
	FullMap = `
	SELECT * FROM full_map
	`

	//FullMapEntryByIP queries for all rows in the full_map table that match the ip
	FullMapEntryByIP = `
	SELECT * from full_map where ip = ?
	`

	//NewEntryById queries for all rows in the new_entries table that match the id
	NewEntryById = `
	SELECT entry FROM new_entries where id = ?
	`

	//NewEntries queries for entries in the new_entries table
	NewEntries = `
	SELECT entry FROM new_entries
	`
)
