-- name: GetAllFactions :many
SELECT * FROM factions ORDER BY id;

-- name: GetFactionByID :one
SELECT * FROM factions WHERE id = $1;

-- name: GetFactionByName :one
SELECT * FROM factions WHERE name = $1;

-- name: UpdatePlayerFaction :exec
UPDATE players 
SET faction = $1
WHERE id = $2;

-- name: GetPlayerWithFaction :one
SELECT 
    p.id,
    p.email,
    p.display_name,
    p.faction,
    p.created_at,
    p.user_id,
    f.name as faction_name
FROM players p
JOIN factions f ON p.faction = f.id
WHERE p.id = $1;