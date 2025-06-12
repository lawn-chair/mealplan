-- +goose Up
-- +goose StatementBegin
-- Add tags table
CREATE TABLE IF NOT EXISTS tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(64) NOT NULL UNIQUE
);

-- Add recipe_tags join table
CREATE TABLE IF NOT EXISTS recipe_tags (
    recipe_id INTEGER NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (recipe_id, tag_id)
);

-- Add meal_tags join table
CREATE TABLE IF NOT EXISTS meal_tags (
    meal_id INTEGER NOT NULL REFERENCES meals(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (meal_id, tag_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS meal_tags;
DROP TABLE IF EXISTS recipe_tags;
DROP TABLE IF EXISTS tags;
-- +goose StatementEnd
