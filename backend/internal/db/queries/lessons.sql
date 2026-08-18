-- name: ListPublishedLessons :many
SELECT l.*, c.name AS category_name, c.slug AS category_slug
FROM lessons l JOIN lesson_categories c ON c.id = l.category_id
WHERE l.published = true
ORDER BY c.sort_order, l.sort_order, l.title;

-- name: GetLessonByID :one
SELECT l.*, c.name AS category_name, c.slug AS category_slug
FROM lessons l JOIN lesson_categories c ON c.id = l.category_id
WHERE l.id = $1 AND l.published = true;

-- name: GetLessonProgress :one
SELECT * FROM user_lesson_progress WHERE user_id = $1 AND lesson_id = $2;

-- name: UpsertLessonProgress :one
INSERT INTO user_lesson_progress (user_id, lesson_id, completed, progress_percent, completed_at)
VALUES ($1, $2, $3, $4, CASE WHEN $3 THEN NOW() ELSE NULL END)
ON CONFLICT (user_id, lesson_id) DO UPDATE
SET completed = EXCLUDED.completed,
    progress_percent = EXCLUDED.progress_percent,
    completed_at = CASE WHEN EXCLUDED.completed THEN COALESCE(user_lesson_progress.completed_at, NOW()) ELSE NULL END
RETURNING *;
