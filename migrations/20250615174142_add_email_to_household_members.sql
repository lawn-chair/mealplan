-- +goose Up
-- +goose StatementBegin
ALTER TABLE household_members ADD COLUMN email VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE household_members DROP COLUMN email;
-- +goose StatementEnd
