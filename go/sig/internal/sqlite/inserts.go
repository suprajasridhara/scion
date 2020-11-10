package sqlite

const (
	InsertNewMSToken = `
	INSERT INTO ms_token(token) VALUES (?)
	`

	InsertNewPushedPrefix = `
	INSERT INTO pushed_prefixes(prefix) VALUES (?)
	`
)
