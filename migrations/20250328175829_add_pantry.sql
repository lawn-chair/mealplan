-- +goose Up
-- +goose StatementBegin
CREATE TABLE pantry (
    id SERIAL PRIMARY KEY,
    user_id text NOT NULL,
    UNIQUE (user_id)
);

CREATE TABLE pantry_items (
    id SERIAL PRIMARY KEY,
    pantry_id integer,
    item_name text NOT NULL,
    UNIQUE (pantry_id, item_name)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE pantry_items;
DROP TABLE pantry;
-- +goose StatementEnd
