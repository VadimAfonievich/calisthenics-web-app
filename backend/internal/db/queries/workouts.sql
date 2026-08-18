-- name: GetWorkoutByID :one
SELECT * FROM workouts WHERE id = $1;

-- name: GetTodayWorkout :one
SELECT w.*
FROM workouts w
JOIN programs p ON p.id = w.program_id
WHERE p.published = true
ORDER BY p.difficulty, w.day_number
LIMIT 1;

-- name: ListWorkoutExercises :many
SELECT we.*, e.name AS exercise_name, e.slug AS exercise_slug, e.difficulty AS exercise_difficulty
FROM workout_exercises we
JOIN exercises e ON e.id = we.exercise_id
WHERE we.workout_id = $1
ORDER BY we.sort_order, e.name;

-- name: CreateWorkoutSession :one
INSERT INTO workout_sessions (user_id, workout_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetWorkoutSession :one
SELECT * FROM workout_sessions WHERE id = $1;

-- name: CompleteWorkoutSession :one
UPDATE workout_sessions
SET status = 'completed', completed_at = NOW(), duration_seconds = $2, xp_earned = $3
WHERE id = $1 AND status = 'started'
RETURNING *;

-- name: CreateExerciseSet :one
INSERT INTO exercise_sets (session_id, exercise_id, set_number, reps, duration_seconds, weight, completed)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;
