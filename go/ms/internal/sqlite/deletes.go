package sqlite

const (
	//DelFullMapEntry deletes rows from the full_map table by id
	DelFullMapEntry = `
	DELETE FROM full_map WHERE id = ?
	`
)
