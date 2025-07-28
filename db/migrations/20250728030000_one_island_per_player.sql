-- +goose Up
-- +goose StatementBegin

-- Add unique constraint to ensure one port per player
ALTER TABLE ports ADD CONSTRAINT unique_player_port UNIQUE (player_id);

-- Update ports table to make it clearer this represents islands
ALTER TABLE ports ADD COLUMN IF NOT EXISTS island_type TEXT DEFAULT 'tropical';
ALTER TABLE ports ADD COLUMN IF NOT EXISTS starting_resources_initialized BOOLEAN DEFAULT FALSE;

-- Add some starting resource values for new islands
INSERT INTO resources (port_id, wood, iron, gold, grain) 
SELECT p.id, 100, 20, 50, 25 
FROM ports p 
WHERE NOT EXISTS (SELECT 1 FROM resources r WHERE r.port_id = p.id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE ports DROP CONSTRAINT IF EXISTS unique_player_port;
ALTER TABLE ports DROP COLUMN IF EXISTS island_type;
ALTER TABLE ports DROP COLUMN IF EXISTS starting_resources_initialized;
-- +goose StatementEnd