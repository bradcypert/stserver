-- name: CreateUser :one
INSERT INTO users (email, password_hash, email_verification_token, email_verification_expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByVerificationToken :one
SELECT * FROM users WHERE email_verification_token = $1;

-- name: VerifyUserEmail :exec
UPDATE users 
SET email_verified = TRUE, 
    email_verification_token = NULL, 
    email_verification_expires_at = NULL,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE users 
SET password_hash = $1, 
    updated_at = NOW()
WHERE id = $2;

-- name: UpdatePlayerUserID :exec
UPDATE players 
SET user_id = $1
WHERE id = $2;