-- name: SaveOTP :exec
INSERT INTO otp_codes (email, hashed_code, expires_at, attempts)
VALUES ($1, $2, $3, 0)
ON CONFLICT (email) DO UPDATE
SET hashed_code = EXCLUDED.hashed_code,
    expires_at = EXCLUDED.expires_at,
    attempts = 0;

-- name: GetOTP :one
SELECT hashed_code, expires_at, attempts
FROM otp_codes
WHERE email = $1;

-- name: IncrementOTPAttempts :exec
UPDATE otp_codes
SET attempts = attempts + 1
WHERE email = $1;

-- name: DeleteOTP :exec
DELETE FROM otp_codes
WHERE email = $1;