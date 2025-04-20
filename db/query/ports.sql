-- name: GetPortById :one
SELECT * FROM ports WHERE id = $1;

-- name: CreatePort :one
INSERT INTO ports (player_id, name, x, y)
VALUES ($1, $2, $3, $4)
RETURNING *; 