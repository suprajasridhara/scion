package sqlite

const (
	//InsNewEntry inserts a row into the new_entries table
	InsNewEntry = `
	INSERT INTO new_entries(entry) VALUES (?)
	`

	//InsFullMapEntry inserts a row into the full_map table
	InsFullMapEntry = `
	INSERT INTO full_map(ip, ia, timestamp) VALUES (?,?,?)
	`

	//InsPCNRep inserts a row into the pcn_reps table
	InsPCNRep = `
	INSERT INTO pcn_reps(fullRep) VALUES (?)
	`
)
