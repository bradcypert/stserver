-- +goose Up
-- +goose StatementBegin
CREATE TABLE buildings (
    id SERIAL PRIMARY KEY,
    port_id INTEGER NOT NULL REFERENCES ports(id) ON DELETE CASCADE,
    type TEXT NOT NULL, -- e.g. 'trade_office', 'shipyard', etc.
    level INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE buildings;
-- +goose StatementEnd
