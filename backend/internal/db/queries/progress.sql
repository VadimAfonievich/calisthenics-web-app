-- name: GetUserProgress :one
SELECT up.*, p.xp, p.level, p.current_streak, p.longest_streak
FROM user_progress up JOIN profiles p ON p.user_id = up.user_id
WHERE up.user_id = $1;

-- name: GetWorkoutHistory :many
SELECT ws.*, w.title AS workout_title, p.name AS program_name
FROM workout_sessions ws
JOIN workouts w ON w.id = ws.workout_id
JOIN programs p ON p.id = w.program_id
WHERE ws.user_id = $1 AND ws.status = 'completed'
ORDER BY ws.started_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUserExerciseStats :many
SELECT ues.*, e.name AS exercise_name, e.slug AS exercise_slug
FROM user_exercise_stats ues
JOIN exercises e ON e.id = ues.exercise_id
WHERE ues.user_id = $1
ORDER BY ues.last_performed_at DESC NULLS LAST;
