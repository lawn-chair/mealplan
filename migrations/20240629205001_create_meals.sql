-- +goose Up
-- +goose StatementBegin
CREATE TABLE meals (
    id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name text,
    slug text,
    description text
);

CREATE TABLE meal_recipes (
    meal_id integer REFERENCES meals(id),
    recipe_id integer REFERENCES recipes(id)
);

CREATE TABLE meal_steps (
    id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    text text,
    "order" integer,
    meal_id integer REFERENCES meals(id)
);

CREATE TABLE meal_ingredients (
    id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    amount text,
    name text,
    meal_id integer REFERENCES meals(id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE meal_recipes;
DROP TABLE meal_steps;
DROP TABLE meal_ingredients;
DROP TABLE meals;

-- +goose StatementEnd
