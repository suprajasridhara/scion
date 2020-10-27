package sqlite3

const (
	InsNewEntry = `
	INSERT INTO new_entries(entry) VALUES (?)
	`

	InsFullMapEntry = `
	INSERT INTO full_map(ip, ia, timestamp) VALUES (?,?,?)
	`
	InsPCNRep = `
	INSERT INTO pcn_reps(fullRep) VALUES (?)
	`
)
