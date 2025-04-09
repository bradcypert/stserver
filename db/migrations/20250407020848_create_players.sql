-- +goose Up
-- +goose StatementBegin
CREATE TABLE factions (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE
);

INSERT INTO factions (name) VALUES
('Unaffiliated'),
('British'),
('French'),
('Spanish'),
('Dutch');

CREATE TABLE players (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    faction INTEGER NOT NULL REFERENCES factions(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE players;
DROP TABLE factions;
-- +goose StatementEnd
