-- +goose Up
-- +goose StatementBegin
CREATE TABLE ports (
    id SERIAL PRIMARY KEY,
    player_id INTEGER NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    x INTEGER NOT NULL, -- grid location
    y INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE ports;
-- +goose StatementEnd
