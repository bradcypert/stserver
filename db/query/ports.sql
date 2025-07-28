-- name: GetPortById :one
SELECT * FROM ports WHERE id = $1;

-- name: GetPortByPlayerId :one
SELECT * FROM ports WHERE player_id = $1;

-- name: CreatePort :one
INSERT INTO ports (player_id, name, x, y)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CreatePlayerIsland :one
INSERT INTO ports (player_id, name, x, y, island_type, starting_resources_initialized)
VALUES ($1, $2, $3, $4, 'tropical', TRUE)
RETURNING *; 