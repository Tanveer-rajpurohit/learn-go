-- name: CreateUserWithPassword :one
INSERT INTO users (name, email, password)
VALUES ($1, $2, $3)
RETURNING id, name, email, role, avatar_sd, avatar_hd, avatar_raw, created_at;

-- name: GetUserByEmail :one
SELECT id, name, email, password, role FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT id, name, email, role, avatar_sd, avatar_hd, avatar_raw, created_at 
FROM users 
WHERE id = $1;

-- name: UpdateUserAvatar :one
UPDATE users
SET 
  avatar_sd = COALESCE(sqlc.narg('avatar_sd'), avatar_sd),
  avatar_hd = COALESCE(sqlc.narg('avatar_hd'), avatar_hd),
  avatar_raw = COALESCE(sqlc.narg('avatar_raw'), avatar_raw)
WHERE id = $1
RETURNING id, name, email, role, avatar_sd, avatar_hd, avatar_raw, created_at;

-- name: UpdateUser :one
UPDATE users
SET 
  name = COALESCE(sqlc.narg('name'), name),
  email = COALESCE(sqlc.narg('email'), email)
WHERE id = $1
RETURNING id, name, email, role, avatar_sd, avatar_hd, avatar_raw, created_at;
