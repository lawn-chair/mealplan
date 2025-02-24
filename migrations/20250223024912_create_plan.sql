-- +goose Up
-- +goose StatementBegin
CREATE TABLE plans (
    id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    start_date date,
    end_date date,
    user_id text,
    UNIQUE (start_date, user_id) 
);

CREATE TABLE plan_meals (
    id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    plan_id integer REFERENCES plans(id),
    meal_id integer REFERENCES meals(id),
    UNIQUE (plan_id, meal_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE plans;
DROP TABLE plan_meals;
-- +goose StatementEnd
