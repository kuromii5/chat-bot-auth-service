package postgres

// refresh token queries
const (
	createRefreshTokenQuery = `
		INSERT INTO auth.refresh_tokens (user_id, token_hash, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	revokeRefreshTokenQuery = `
		UPDATE auth.refresh_tokens 
		SET is_revoked = true, revoked_at = NOW() 
		WHERE token_hash = $1 AND is_revoked = false
	`

	getRefreshTokenQuery = `
        SELECT id, user_id, token_hash, user_agent, ip_address, is_revoked, expires_at, created_at, revoked_at
        FROM auth.refresh_tokens
        WHERE token_hash = $1
    `

	revokeAllTokensQuery = `
        UPDATE auth.refresh_tokens
        SET is_revoked = true, revoked_at = NOW()
        WHERE user_id = $1 AND is_revoked = false
    `
)

// user queries
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
		SELECT id, email, password_hash, token_version, created_at, updated_at
		FROM auth.users
		WHERE email = $1
	`

	findByUsernameQuery = `
		SELECT id, email, password_hash, token_version, created_at, updated_at
		FROM auth.users
		WHERE username = $1
	`
)
