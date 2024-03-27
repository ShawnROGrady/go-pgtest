-- name: CreateUser :one
INSERT INTO users (
	email,
	name
)
VALUES (
	$1,
	$2
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE id=$1;
