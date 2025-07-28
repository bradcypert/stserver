-- +goose Up
-- +goose StatementBegin

-- Create building types configuration table
CREATE TABLE building_types (
    id SERIAL PRIMARY KEY,
    type_name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    description TEXT,
    category TEXT NOT NULL, -- 'resource', 'trade', 'military', 'infrastructure'
    max_level INTEGER NOT NULL DEFAULT 5,
    base_cost_wood INTEGER NOT NULL DEFAULT 0,
    base_cost_iron INTEGER NOT NULL DEFAULT 0,
    base_cost_gold INTEGER NOT NULL DEFAULT 0,
    base_build_time INTEGER NOT NULL DEFAULT 300, -- seconds
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert predefined building types
INSERT INTO building_types (type_name, display_name, description, category, max_level, base_cost_wood, base_cost_iron, base_cost_gold, base_build_time) VALUES
('lumberyard', 'Lumberyard', 'Produces wood resources over time', 'resource', 5, 50, 0, 10, 300),
('mine', 'Mine', 'Produces iron and precious metals over time', 'resource', 5, 30, 10, 25, 450),
('plantation', 'Plantation', 'Produces sugar, tobacco, cotton, and coffee', 'resource', 5, 40, 0, 15, 360),
('farm', 'Farm', 'Produces grain and rum over time', 'resource', 5, 25, 0, 5, 240),
('trade_center', 'Trade Center', 'Enables trading with other players and NPCs', 'trade', 3, 100, 50, 100, 600),
('shipyard', 'Shipyard', 'Enables ship construction and repairs', 'military', 5, 150, 100, 200, 900),
('warehouse', 'Warehouse', 'Increases resource storage capacity', 'infrastructure', 10, 80, 20, 30, 300),
('dock', 'Dock', 'Required for ship docking and trade', 'infrastructure', 3, 60, 30, 20, 360),
('tavern', 'Tavern', 'Increases crew recruitment and morale', 'infrastructure', 5, 40, 10, 50, 300),
('fort', 'Fort', 'Provides island defense against attacks', 'military', 5, 100, 200, 150, 1200);

-- Create building production rates table
CREATE TABLE building_production (
    id SERIAL PRIMARY KEY,
    building_type_id INTEGER NOT NULL REFERENCES building_types(id),
    level INTEGER NOT NULL,
    resource_type TEXT NOT NULL, -- matches column names in resources table
    production_rate INTEGER NOT NULL DEFAULT 0, -- per tick
    UNIQUE(building_type_id, level, resource_type)
);

-- Insert production rates for each building type and level
-- Lumberyard production (wood)
INSERT INTO building_production (building_type_id, level, resource_type, production_rate) VALUES
((SELECT id FROM building_types WHERE type_name = 'lumberyard'), 1, 'wood', 5),
((SELECT id FROM building_types WHERE type_name = 'lumberyard'), 2, 'wood', 8),
((SELECT id FROM building_types WHERE type_name = 'lumberyard'), 3, 'wood', 12),
((SELECT id FROM building_types WHERE type_name = 'lumberyard'), 4, 'wood', 18),
((SELECT id FROM building_types WHERE type_name = 'lumberyard'), 5, 'wood', 25);

-- Mine production (iron, gold, silver)
INSERT INTO building_production (building_type_id, level, resource_type, production_rate) VALUES
((SELECT id FROM building_types WHERE type_name = 'mine'), 1, 'iron', 3),
((SELECT id FROM building_types WHERE type_name = 'mine'), 1, 'gold', 1),
((SELECT id FROM building_types WHERE type_name = 'mine'), 2, 'iron', 5),
((SELECT id FROM building_types WHERE type_name = 'mine'), 2, 'gold', 2),
((SELECT id FROM building_types WHERE type_name = 'mine'), 3, 'iron', 8),
((SELECT id FROM building_types WHERE type_name = 'mine'), 3, 'gold', 3),
((SELECT id FROM building_types WHERE type_name = 'mine'), 3, 'silver', 1),
((SELECT id FROM building_types WHERE type_name = 'mine'), 4, 'iron', 12),
((SELECT id FROM building_types WHERE type_name = 'mine'), 4, 'gold', 5),
((SELECT id FROM building_types WHERE type_name = 'mine'), 4, 'silver', 2),
((SELECT id FROM building_types WHERE type_name = 'mine'), 5, 'iron', 18),
((SELECT id FROM building_types WHERE type_name = 'mine'), 5, 'gold', 8),
((SELECT id FROM building_types WHERE type_name = 'mine'), 5, 'silver', 4);

-- Plantation production (sugar, tobacco, cotton, coffee)
INSERT INTO building_production (building_type_id, level, resource_type, production_rate) VALUES
((SELECT id FROM building_types WHERE type_name = 'plantation'), 1, 'sugar', 2),
((SELECT id FROM building_types WHERE type_name = 'plantation'), 1, 'tobacco', 2),
((SELECT id FROM building_types WHERE type_name = 'plantation'), 2, 'sugar', 4),
((SELECT id FROM building_types WHERE type_name = 'plantation'), 2, 'tobacco', 3),
((SELECT id FROM building_types WHERE type_name = 'plantation'), 2, 'cotton', 2),
((SELECT id FROM building_types WHERE type_name = 'plantation'), 3, 'sugar', 6),
((SELECT id FROM building_types WHERE type_name = 'plantation'), 3, 'tobacco', 5),
((SELECT id FROM building_types WHERE type_name = 'plantation'), 3, 'cotton', 4),
((SELECT id FROM building_types WHERE type_name = 'plantation'), 3, 'coffee', 2),
((SELECT id FROM building_types WHERE type_name = 'plantation'), 4, 'sugar', 10),
((SELECT id FROM building_types WHERE type_name = 'plantation'), 4, 'tobacco', 8),
((SELECT id FROM building_types WHERE type_name = 'plantation'), 4, 'cotton', 6),
((SELECT id FROM building_types WHERE type_name = 'plantation'), 4, 'coffee', 4),
((SELECT id FROM building_types WHERE type_name = 'plantation'), 5, 'sugar', 15),
((SELECT id FROM building_types WHERE type_name = 'plantation'), 5, 'tobacco', 12),
((SELECT id FROM building_types WHERE type_name = 'plantation'), 5, 'cotton', 10),
((SELECT id FROM building_types WHERE type_name = 'plantation'), 5, 'coffee', 8);

-- Farm production (grain, rum)
INSERT INTO building_production (building_type_id, level, resource_type, production_rate) VALUES
((SELECT id FROM building_types WHERE type_name = 'farm'), 1, 'grain', 4),
((SELECT id FROM building_types WHERE type_name = 'farm'), 1, 'rum', 1),
((SELECT id FROM building_types WHERE type_name = 'farm'), 2, 'grain', 7),
((SELECT id FROM building_types WHERE type_name = 'farm'), 2, 'rum', 2),
((SELECT id FROM building_types WHERE type_name = 'farm'), 3, 'grain', 12),
((SELECT id FROM building_types WHERE type_name = 'farm'), 3, 'rum', 4),
((SELECT id FROM building_types WHERE type_name = 'farm'), 4, 'grain', 18),
((SELECT id FROM building_types WHERE type_name = 'farm'), 4, 'rum', 6),
((SELECT id FROM building_types WHERE type_name = 'farm'), 5, 'grain', 25),
((SELECT id FROM building_types WHERE type_name = 'farm'), 5, 'rum', 10);

-- Add building state tracking
ALTER TABLE buildings ADD COLUMN under_construction BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE buildings ADD COLUMN construction_complete_at TIMESTAMPTZ;
ALTER TABLE buildings ADD COLUMN last_production_at TIMESTAMPTZ DEFAULT NOW();

-- Create indexes for performance
CREATE INDEX idx_buildings_port_id ON buildings(port_id);
CREATE INDEX idx_buildings_type ON buildings(type);
CREATE INDEX idx_buildings_under_construction ON buildings(under_construction);
CREATE INDEX idx_building_production_type_level ON building_production(building_type_id, level);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE buildings DROP COLUMN under_construction;
ALTER TABLE buildings DROP COLUMN construction_complete_at;
ALTER TABLE buildings DROP COLUMN last_production_at;
DROP TABLE building_production;
DROP TABLE building_types;
-- +goose StatementEnd