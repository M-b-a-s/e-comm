-- +goose Up

CREATE TABLE otp_codes (
    email CITEXT PRIMARY KEY,
    hashed_code TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0
);

-- +goose Down

DROP TABLE IF EXISTS otp_codes;