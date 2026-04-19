-- name: CreateUser :one
INSERT INTO users (name, email)
VALUES ($1, $2)
RETURNING id, name, email;

-- name: GetUsers :many
SELECT id, name, email FROM users;

-- name: GetUser :one
SELECT id, name, email FROM users WHERE id = $1;

-- name: UpdateUser :one
UPDATE users
SET name = $2, email = $3
WHERE id = $1
RETURNING id, name, email;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;