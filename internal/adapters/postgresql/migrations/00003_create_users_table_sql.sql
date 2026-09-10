-- +goose Up
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TYPE user_role AS ENUM ('customer', 'admin');

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email CITEXT NOT NULL,
    password_hash TEXT NOT NULL,
    phone_number TEXT,
    country TEXT NOT NULL,
    username CITEXT NOT NULL,
    role user_role NOT NULL DEFAULT 'customer',
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT users_username_length_check
        CHECK (LENGTH(username) BETWEEN 3 AND 30)
);

CREATE UNIQUE INDEX users_email_unique_idx ON users (email) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX users_username_unique_idx ON users (username) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX users_phone_unique_idx ON users (phone_number) WHERE deleted_at IS NULL AND phone_number IS NOT NULL;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS users_set_updated_at ON users;
DROP FUNCTION IF EXISTS set_updated_at();
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_role;