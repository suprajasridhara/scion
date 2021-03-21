package sqlite

const (
	//InsertNewMSToken query to insert values into ms_token
	InsertNewMSToken = `
	INSERT INTO ms_token(token) VALUES (?)
	`

	//InsertNewPushedPrefix query to insert values into pushed_prefixes
	InsertNewPushedPrefix = `
	INSERT INTO pushed_prefixes(prefix) VALUES (?)
	`
)
