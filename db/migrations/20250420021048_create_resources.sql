-- +goose Up
-- +goose StatementBegin
CREATE TABLE resources (
    port_id INTEGER PRIMARY KEY NOT NULL REFERENCES ports(id) ON DELETE CASCADE,
    type TEXT NOT NULL, -- e.g. 'trade_office', 'shipyard', etc.
    level INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE resources;
-- +goose StatementEnd
