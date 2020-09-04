package sqlite

const (
	InsertNewMSToken = `
	INSERT INTO ms_token(token) VALUES (?)
	`
)
