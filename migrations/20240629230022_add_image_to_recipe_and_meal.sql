-- +goose Up
-- +goose StatementBegin
ALTER TABLE recipes ADD COLUMN image text;
ALTER TABLE meals ADD COLUMN image text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE recipes DROP COLUMN image;
ALTER TABLE meals DROP COLUMN image;
-- +goose StatementEnd
