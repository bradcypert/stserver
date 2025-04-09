-- name: GetPlayerByID :one
SELECT * FROM players WHERE id = $1;

-- name: CreatePlayer :one
INSERT INTO players (email, display_name, faction)
VALUES ($1, $2, $3)
RETURNING *; 
