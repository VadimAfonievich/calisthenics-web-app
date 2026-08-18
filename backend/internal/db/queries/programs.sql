-- name: ListPublishedPrograms :many
SELECT * FROM programs WHERE published = true ORDER BY difficulty, name;

-- name: GetProgramByID :one
SELECT * FROM programs WHERE id = $1 AND published = true;
