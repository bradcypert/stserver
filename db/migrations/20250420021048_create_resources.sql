-- +goose Up
-- +goose StatementBegin
CREATE TABLE resources (
    port_id INTEGER PRIMARY KEY NOT NULL REFERENCES ports(id) ON DELETE CASCADE,
    wood INTEGER NOT NULL DEFAULT 0,
    iron INTEGER NOT NULL DEFAULT 0,
    rum INTEGER NOT NULL DEFAULT 0,
    sugar INTEGER NOT NULL DEFAULT 0,
    tobacco INTEGER NOT NULL DEFAULT 0,
    cotton INTEGER NOT NULL DEFAULT 0,
    coffee INTEGER NOT NULL DEFAULT 0,
    grain INTEGER NOT NULL DEFAULT 0,
    gold INTEGER NOT NULL DEFAULT 0,
    silver INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE resources;
-- +goose StatementEnd
