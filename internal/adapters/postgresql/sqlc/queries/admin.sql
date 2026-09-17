-- name: CreateAdminUser :one
INSERT INTO users (name, email, password_hash, phone_number, country, username, role, email_verified)
VALUES ($1, $2, $3, $4, $5, $6, 'admin', TRUE)
ON CONFLICT (email) WHERE deleted_at IS NULL DO UPDATE
SET role = 'admin', email_verified = TRUE
RETURNING *;