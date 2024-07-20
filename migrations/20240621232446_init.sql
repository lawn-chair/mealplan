-- +goose Up
-- +goose StatementBegin

CREATE TABLE recipes (
    id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name text,
    description text,
    slug text
);

CREATE TABLE recipe_ingredients (
    id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    recipe_id integer REFERENCES recipes(id),
    name text,
    amount text,
    calories integer
);

CREATE TABLE recipe_steps (
    id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    text text,
    "order" integer,
    recipe_id integer REFERENCES recipes(id)
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE recipe_steps;
DROP TABLE recipe_ingredients;
DROP TABLE recipes;


-- +goose StatementEnd
