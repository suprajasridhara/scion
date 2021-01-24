package sqlite

const (
	//PushedPrefixes queries all prefixes in pushed_prefixes
	PushedPrefixes = `
	SELECT prefix FROM pushed_prefixes
	`
)
