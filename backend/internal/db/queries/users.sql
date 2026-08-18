-- name: GetUserByID :one
SELECT u.*, p.display_name, p.level, p.xp, p.current_streak, p.longest_streak, p.preferred_difficulty, p.timezone
FROM users u JOIN profiles p ON p.user_id = u.id
WHERE u.id = $1;

-- name: GetUserByTelegramID :one
SELECT u.*, p.display_name, p.level, p.xp, p.current_streak, p.longest_streak, p.preferred_difficulty, p.timezone
FROM users u JOIN profiles p ON p.user_id = u.id
WHERE u.telegram_id = $1;

-- name: CreateUser :one
INSERT INTO users (telegram_id, username, first_name, last_name, photo_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: CreateProfile :one
INSERT INTO profiles (user_id, display_name, timezone)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateUserProfile :one
UPDATE profiles
SET display_name = $2, preferred_difficulty = $3, timezone = $4
WHERE user_id = $1
RETURNING *;
