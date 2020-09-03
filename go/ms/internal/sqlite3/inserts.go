package sqlite3

const (
	InsertDummyData1 = `
	INSERT INTO full_map VALUES(
		1,"10.17.23.0/24","1-ff00:0:110"
	)
	`
	InsertDummyData2 = `
	INSERT INTO full_map VALUES(
		2,"10.17.32.0/24","2-ff00:0:110"
	)
	`

	InsertNewEntry = `
	INSERT INTO new_entries(entry) VALUES (?)
	`
)
