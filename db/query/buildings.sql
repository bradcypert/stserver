-- name: GetBuilding :one
SELECT * FROM buildings WHERE id = $1;

-- name: GetBuildingByPortAndType :one
SELECT * FROM buildings WHERE port_id = $1 AND type = $2;

-- name: CreateBuilding :one
INSERT INTO buildings (port_id, type)
VALUES ($1, $2)
RETURNING *; 

-- name: GetBuildingsByPort :many
SELECT * FROM buildings WHERE port_id = $1;

-- name: UpdateBuilding :one
UPDATE buildings
SET level = $2
WHERE id = $1
RETURNING *;
