package token

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
