-- name: ListExercises :many
SELECT * FROM exercises ORDER BY difficulty, name;

-- name: GetExerciseByID :one
SELECT * FROM exercises WHERE id = $1;
