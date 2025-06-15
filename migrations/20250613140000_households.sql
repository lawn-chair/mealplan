-- +goose Up
-- +goose StatementBegin
CREATE TABLE households (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table to associate Clerk users with households
CREATE TABLE household_members (
    household_id INTEGER REFERENCES households(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    PRIMARY KEY (household_id, user_id)
);

-- Table for join codes
CREATE TABLE household_join_codes (
    code TEXT PRIMARY KEY,
    household_id INTEGER REFERENCES households(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL
);

-- Add household_id to plans and pantry, remove user_id
ALTER TABLE plans ADD COLUMN household_id INTEGER REFERENCES households(id);
ALTER TABLE pantry ADD COLUMN household_id INTEGER REFERENCES households(id);

-- Create a default household for all existing data
INSERT INTO households (id, name) VALUES (1, 'Default Household') ON CONFLICT DO NOTHING;
UPDATE plans SET household_id = 1 WHERE household_id IS NULL;
UPDATE pantry SET household_id = 1 WHERE household_id IS NULL;

ALTER TABLE plans ALTER COLUMN household_id SET NOT NULL;
ALTER TABLE pantry ALTER COLUMN household_id SET NOT NULL;
ALTER TABLE plans DROP COLUMN IF EXISTS user_id;
ALTER TABLE pantry DROP COLUMN IF EXISTS user_id;

-- Enforce one pantry per household
ALTER TABLE pantry ADD CONSTRAINT unique_household_pantry UNIQUE (household_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE pantry DROP CONSTRAINT IF EXISTS unique_household_pantry;
DROP TABLE IF EXISTS household_join_codes;
DROP TABLE IF EXISTS household_members;
ALTER TABLE pantry ADD COLUMN user_id TEXT;
ALTER TABLE plans ADD COLUMN user_id TEXT;
ALTER TABLE pantry DROP COLUMN IF EXISTS household_id;
ALTER TABLE plans DROP COLUMN IF EXISTS household_id;
DROP TABLE IF EXISTS households;
-- +goose StatementEnd
