-- name: CreateUser :one
INSERT INTO users (stytch_user_id, first_name, last_name, email, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateUser :one
UPDATE users 
SET first_name = $2, last_name = $3, email = $4, status = $5, updated_at = CURRENT_TIMESTAMP
WHERE stytch_user_id = $1
RETURNING *;

-- name: GetUserByStytchID :one
SELECT * FROM users WHERE stytch_user_id = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE stytch_user_id = $1; 