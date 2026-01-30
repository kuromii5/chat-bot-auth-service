package user

const (
	createAuthUserQuery = `
        INSERT INTO auth.users (email, password_hash)
        VALUES ($1, $2)
        RETURNING id, created_at, token_version
    `

	createProfileQuery = `
        INSERT INTO core.profiles (user_id, username, role)
        VALUES ($1, $2, $3)
        RETURNING username, role
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
