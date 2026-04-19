-- name: GetUserByEmail :one
SELECT id, name, email, password, role FROM users WHERE email = $1;

-- name: CreateUserWithPassword :one
INSERT INTO users (name, email, password, role)
VALUES ($1, $2, $3, 'user')
RETURNING id, name, email, role;

-- name: SaveRefreshToken :one
INSERT INTO refresh_tokens (user_id, token, expires_at)
VALUES ($1, $2, $3)
ON CONFLICT (user_id) DO UPDATE
SET token = EXCLUDED.token,
	expires_at = EXCLUDED.expires_at
RETURNING id, user_id, token, expires_at;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens WHERE token = $1;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens WHERE token = $1;