-- +goose Up
-- +goose StatementBegin
CREATE TABLE shopping_status (
    plan_id integer REFERENCES plans(id) ON DELETE CASCADE,
    status jsonb
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE shopping_status;
-- +goose StatementEnd
