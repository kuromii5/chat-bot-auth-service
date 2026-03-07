package postgres

// refresh token queries
const (
	createRefreshTokenQuery = `
		INSERT INTO auth.refresh_tokens (user_id, token_hash, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	revokeRefreshTokenQuery = `
		UPDATE auth.refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`

	getRefreshTokenQuery = `
        SELECT
			rt.id
			, rt.user_id
			, rt.token_hash
			, rt.user_agent
			, rt.ip_address
			, rt.expires_at
			, rt.created_at
			, rt.revoked_at
			, u.role
        FROM auth.refresh_tokens rt
		JOIN auth.users u ON rt.user_id = u.id
        WHERE token_hash = $1
    `

	revokeAllTokensQuery = `
        UPDATE auth.refresh_tokens
        SET revoked_at = NOW()
        WHERE user_id = $1 AND revoked_at IS NULL
    `
)

// user queries
const (
	createAuthUserQuery = `
        INSERT INTO auth.users (email, password_hash, role, email_notifications_enabled)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, token_version, role
    `

	updatePreferencesQuery = `
        UPDATE auth.users
        SET email_notifications_enabled = $1
        WHERE id = $2
    `

	createProfileQuery = `
        INSERT INTO core.profiles (user_id, username)
        VALUES ($1, $2)
        RETURNING user_id
    `

	getUserByEmailQuery = `
		SELECT id, email, password_hash, token_version, created_at, updated_at, role
		FROM auth.users
		WHERE email = $1
	`
)
