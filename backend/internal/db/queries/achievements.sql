-- name: ListAchievements :many
SELECT * FROM achievements ORDER BY condition_type, condition_value, code;

-- name: ListUserAchievements :many
SELECT a.*, ua.unlocked_at
FROM user_achievements ua JOIN achievements a ON a.id = ua.achievement_id
WHERE ua.user_id = $1
ORDER BY ua.unlocked_at DESC;
