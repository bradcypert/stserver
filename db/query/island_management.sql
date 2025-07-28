-- Building Types Queries
-- name: GetAllBuildingTypes :many
SELECT * FROM building_types ORDER BY category, type_name;

-- name: GetBuildingTypeByName :one
SELECT * FROM building_types WHERE type_name = $1;

-- name: GetBuildingTypesByCategory :many
SELECT * FROM building_types WHERE category = $1 ORDER BY type_name;

-- Building Production Queries
-- name: GetProductionRatesForBuilding :many
SELECT bp.resource_type, bp.production_rate 
FROM building_production bp
JOIN building_types bt ON bp.building_type_id = bt.id
WHERE bt.type_name = $1 AND bp.level = $2;

-- name: GetAllProductionForBuildingType :many
SELECT bp.level, bp.resource_type, bp.production_rate
FROM building_production bp
JOIN building_types bt ON bp.building_type_id = bt.id
WHERE bt.type_name = $1
ORDER BY bp.level, bp.resource_type;

-- Island/Port Management Queries
-- name: GetPortWithResources :one
SELECT 
    p.id as port_id,
    p.name as port_name,
    p.player_id,
    p.x,
    p.y,
    p.created_at as port_created_at,
    r.wood,
    r.iron,
    r.rum,
    r.sugar,
    r.tobacco,
    r.cotton,
    r.coffee,
    r.grain,
    r.gold,
    r.silver,
    r.updated_at as resources_updated_at
FROM ports p
LEFT JOIN resources r ON p.id = r.port_id
WHERE p.id = $1;

-- name: GetPortBuildings :many
SELECT 
    b.id,
    b.port_id,
    b.type,
    b.level,
    b.under_construction,
    b.construction_complete_at,
    b.last_production_at,
    b.created_at,
    bt.display_name,
    bt.description,
    bt.category,
    bt.max_level
FROM buildings b
JOIN building_types bt ON b.type = bt.type_name
WHERE b.port_id = $1
ORDER BY bt.category, b.type;

-- name: CreateBuildingConstruction :one
INSERT INTO buildings (port_id, type, under_construction, construction_complete_at)
VALUES ($1, $2, TRUE, $3)
RETURNING *;

-- name: CompleteBuildingConstruction :exec
UPDATE buildings 
SET under_construction = FALSE, 
    construction_complete_at = NULL,
    last_production_at = NOW()
WHERE id = $1;

-- name: GetBuildingsUnderConstruction :many
SELECT b.*, bt.display_name, bt.base_build_time
FROM buildings b
JOIN building_types bt ON b.type = bt.type_name
WHERE b.under_construction = TRUE 
AND b.construction_complete_at <= NOW();

-- name: UpgradeBuilding :exec
UPDATE buildings 
SET level = level + 1,
    under_construction = TRUE,
    construction_complete_at = $2
WHERE buildings.id = $1 AND buildings.level < (
    SELECT bt.max_level FROM building_types bt WHERE bt.type_name = (
        SELECT b.type FROM buildings b WHERE b.id = $1
    )
);

-- name: GetBuildingsReadyForProduction :many
SELECT 
    b.id,
    b.port_id,
    b.type,
    b.level,
    b.last_production_at
FROM buildings b
WHERE b.under_construction = FALSE 
AND (b.last_production_at IS NULL OR b.last_production_at < $1);

-- name: UpdateBuildingLastProduction :exec
UPDATE buildings 
SET last_production_at = NOW()
WHERE id = $1;

-- Resource Management Queries
-- name: InitializePortResources :exec
INSERT INTO resources (port_id) 
VALUES ($1)
ON CONFLICT (port_id) DO NOTHING;

-- name: AddResourcesToPort :exec
UPDATE resources 
SET 
    wood = wood + $2,
    iron = iron + $3,
    rum = rum + $4,
    sugar = sugar + $5,
    tobacco = tobacco + $6,
    cotton = cotton + $7,
    coffee = coffee + $8,
    grain = grain + $9,
    gold = gold + $10,
    silver = silver + $11,
    updated_at = NOW()
WHERE port_id = $1;

-- name: ConsumeResourcesFromPort :exec
UPDATE resources 
SET 
    wood = GREATEST(0, wood - $2),
    iron = GREATEST(0, iron - $3),
    rum = GREATEST(0, rum - $4),
    sugar = GREATEST(0, sugar - $5),
    tobacco = GREATEST(0, tobacco - $6),
    cotton = GREATEST(0, cotton - $7),
    coffee = GREATEST(0, coffee - $8),
    grain = GREATEST(0, grain - $9),
    gold = GREATEST(0, gold - $10),
    silver = GREATEST(0, silver - $11),
    updated_at = NOW()
WHERE port_id = $1;

-- name: CheckResourceAvailability :one
SELECT 
    (wood >= $2) as has_wood,
    (iron >= $3) as has_iron,
    (gold >= $4) as has_gold
FROM resources 
WHERE port_id = $1;