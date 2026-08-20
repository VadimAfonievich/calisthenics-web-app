package lessons

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const LessonCompletionXP = 20

var ErrNotFound = errors.New("lesson not found")

type Lesson struct {
	ID               string  `json:"id"`
	CategoryID       string  `json:"category_id"`
	CategoryName     string  `json:"category_name"`
	Title            string  `json:"title"`
	ShortDescription string  `json:"short_description"`
	Content          string  `json:"content"`
	ContentBlocks    []Block `json:"content_blocks"`
	CoverMediaURL    string  `json:"cover_media_url,omitempty"`
	Difficulty       string  `json:"difficulty"`
	DurationMinutes  int32   `json:"duration_minutes"`
	Completed        bool    `json:"completed"`
	ProgressPercent  int16   `json:"progress_percent"`
}
type Block struct {
	Type     string   `json:"type"`
	Text     string   `json:"text,omitempty"`
	MediaID  *string  `json:"media_id,omitempty"`
	URL      string   `json:"url,omitempty"`
	MIMEType string   `json:"mime_type,omitempty"`
	Alt      string   `json:"alt,omitempty"`
	Items    []string `json:"items,omitempty"`
}
type Completion struct {
	XPearned         int32 `json:"xp_earned"`
	TotalXP          int32 `json:"total_xp"`
	AlreadyCompleted bool  `json:"already_completed"`
}
type Service struct{ pool *pgxpool.Pool }

func NewService(pool *pgxpool.Pool) *Service { return &Service{pool: pool} }

func (s *Service) List(ctx context.Context, userID string) ([]Lesson, error) {
	rows, err := s.pool.Query(ctx, lessonSelect+` WHERE l.published ORDER BY c.sort_order,l.sort_order,l.title`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Lesson{}
	for rows.Next() {
		var item Lesson
		var blocks []byte
		if err := rows.Scan(&item.ID, &item.CategoryID, &item.CategoryName, &item.Title, &item.ShortDescription, &item.Content, &blocks, &item.CoverMediaURL, &item.Difficulty, &item.DurationMinutes, &item.Completed, &item.ProgressPercent); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(blocks, &item.ContentBlocks); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (s *Service) Get(ctx context.Context, userID, lessonID string) (Lesson, error) {
	var item Lesson
	var blocks []byte
	err := s.pool.QueryRow(ctx, lessonSelect+` WHERE l.id=$2::uuid AND l.published`, userID, lessonID).Scan(&item.ID, &item.CategoryID, &item.CategoryName, &item.Title, &item.ShortDescription, &item.Content, &blocks, &item.CoverMediaURL, &item.Difficulty, &item.DurationMinutes, &item.Completed, &item.ProgressPercent)
	if errors.Is(err, pgx.ErrNoRows) {
		return Lesson{}, ErrNotFound
	}
	if err == nil {
		err = json.Unmarshal(blocks, &item.ContentBlocks)
	}
	return item, err
}

const lessonSelect = `SELECT l.id::text,l.category_id::text,c.name,l.title,l.short_description,l.content,
COALESCE((SELECT jsonb_agg(CASE WHEN b.value ? 'media_id' THEN b.value || jsonb_build_object('url',m.url,'mime_type',m.mime_type,'alt',m.original_filename) ELSE b.value END ORDER BY b.ordinality) FROM jsonb_array_elements(l.content_blocks) WITH ORDINALITY b(value,ordinality) LEFT JOIN media_assets m ON m.id::text=b.value->>'media_id'),'[]'::jsonb),
COALESCE(cm.url,''),l.difficulty,l.duration_minutes,COALESCE(p.completed,false),COALESCE(p.progress_percent,0)
FROM lessons l JOIN lesson_categories c ON c.id=l.category_id LEFT JOIN media_assets cm ON cm.id=l.cover_media_id LEFT JOIN user_lesson_progress p ON p.lesson_id=l.id AND p.user_id=$1::uuid`

func (s *Service) Complete(ctx context.Context, userID, lessonID string) (Completion, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Completion{}, err
	}
	defer tx.Rollback(ctx)
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM lessons WHERE id=$1::uuid AND published)`, lessonID).Scan(&exists); err != nil || !exists {
		return Completion{}, ErrNotFound
	}
	var wasCompleted bool
	err = tx.QueryRow(ctx, `SELECT completed FROM user_lesson_progress WHERE user_id=$1::uuid AND lesson_id=$2::uuid FOR UPDATE`, userID, lessonID).Scan(&wasCompleted)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Completion{}, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO user_lesson_progress (user_id,lesson_id,completed,progress_percent,completed_at) VALUES ($1::uuid,$2::uuid,true,100,NOW()) ON CONFLICT (user_id,lesson_id) DO UPDATE SET completed=true,progress_percent=100,completed_at=COALESCE(user_lesson_progress.completed_at,NOW())`, userID, lessonID)
	if err != nil {
		return Completion{}, err
	}
	result := Completion{AlreadyCompleted: wasCompleted}
	if !wasCompleted {
		result.XPearned = LessonCompletionXP
		err = tx.QueryRow(ctx, `UPDATE profiles SET xp=xp+$2 WHERE user_id=$1::uuid RETURNING xp`, userID, LessonCompletionXP).Scan(&result.TotalXP)
	} else {
		err = tx.QueryRow(ctx, `SELECT xp FROM profiles WHERE user_id=$1::uuid`, userID).Scan(&result.TotalXP)
	}
	if err != nil {
		return Completion{}, fmt.Errorf("update lesson XP: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Completion{}, err
	}
	return result, nil
}
