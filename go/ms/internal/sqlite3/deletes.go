package sqlite3

const (
	DelFullMapEntry = `
	DELETE FROM full_map WHERE id = ?
	`
)
