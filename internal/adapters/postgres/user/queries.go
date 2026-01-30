package user

const (
	createUserQuery = `
	INSERT INTO core.users (email, password_hash, username, role)
		VALUES ($1, $2, $3, $4)
	RETURNING id, email, password_hash, username, role, token_version, created_at, updated_at
	`

	findByEmailQuery = `
	SELECT id, email, password_hash, username, role, is_verified, token_version, created_at, updated_at
	FROM core.users
	WHERE email = $1
	`

	findByUsernameQuery = `
	SELECT id, email, password_hash, username, role, is_verified, token_version, created_at, updated_at
	FROM core.users
	WHERE username = $1
	`
)
