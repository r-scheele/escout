-- name: CreateUser :one
INSERT INTO users (username, password, email, created_at) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT users.*, total.total_count
FROM users
CROSS JOIN (SELECT COUNT(*) AS total_count FROM users) AS total
LIMIT $1 OFFSET $2;



-- name: UpdateUser :one
UPDATE users SET username = $2, password = $3, email = $4 WHERE id = $1 RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: GetUser :one
SELECT * FROM users WHERE username = $1 OR email = $2 LIMIT 1;
