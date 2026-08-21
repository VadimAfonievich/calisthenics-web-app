package programs

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("program not found")

type Program struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	Description   string    `json:"description"`
	Difficulty    string    `json:"difficulty"`
	DurationWeeks int32     `json:"duration_weeks"`
	Category      string    `json:"category"`
	CoverMediaURL string    `json:"cover_media_url,omitempty"`
	WorkoutCount  int32     `json:"workout_count"`
	Workouts      []Workout `json:"workouts,omitempty"`
}
type Workout struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	EstimatedMinutes int32  `json:"estimated_minutes"`
	Difficulty       string `json:"difficulty"`
	Category         string `json:"category"`
}

type Service struct{ pool *pgxpool.Pool }

func NewService(pool *pgxpool.Pool) *Service { return &Service{pool: pool} }

func (s *Service) List(ctx context.Context) ([]Program, error) {
	rows, err := s.pool.Query(ctx, `SELECT p.id::text,p.name,p.slug,p.description,p.difficulty,p.duration_weeks,p.category,COALESCE(m.url,''),(SELECT count(*)::int FROM workouts w WHERE w.program_id=p.id AND w.status='published') FROM programs p LEFT JOIN media_assets m ON m.id=p.cover_media_id WHERE p.published ORDER BY p.category,p.difficulty,p.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Program{}
	for rows.Next() {
		var item Program
		if err = rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Description, &item.Difficulty, &item.DurationWeeks, &item.Category, &item.CoverMediaURL, &item.WorkoutCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) Get(ctx context.Context, id string) (Program, error) {
	var item Program
	err := s.pool.QueryRow(ctx, `SELECT p.id::text,p.name,p.slug,p.description,p.difficulty,p.duration_weeks,p.category,COALESCE(m.url,''),(SELECT count(*)::int FROM workouts w WHERE w.program_id=p.id AND w.status='published') FROM programs p LEFT JOIN media_assets m ON m.id=p.cover_media_id WHERE p.id=$1::uuid AND p.published`, id).Scan(&item.ID, &item.Name, &item.Slug, &item.Description, &item.Difficulty, &item.DurationWeeks, &item.Category, &item.CoverMediaURL, &item.WorkoutCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return Program{}, ErrNotFound
	}
	if err != nil {
		return Program{}, err
	}
	rows, e := s.pool.Query(ctx, `SELECT w.id::text,w.title,w.description,w.estimated_minutes,p.difficulty,w.category FROM workouts w JOIN programs p ON p.id=w.program_id WHERE w.program_id=$1::uuid AND w.status='published' ORDER BY w.sort_order,w.day_number,w.id`, id)
	if e != nil {
		return Program{}, e
	}
	defer rows.Close()
	item.Workouts = []Workout{}
	for rows.Next() {
		var workout Workout
		if e = rows.Scan(&workout.ID, &workout.Title, &workout.Description, &workout.EstimatedMinutes, &workout.Difficulty, &workout.Category); e != nil {
			return Program{}, e
		}
		item.Workouts = append(item.Workouts, workout)
	}
	return item, rows.Err()
}
