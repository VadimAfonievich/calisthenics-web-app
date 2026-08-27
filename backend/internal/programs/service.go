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
	Levels        []Level   `json:"levels,omitempty"`
}
type Level struct {
	ID              string    `json:"id"`
	LevelNumber     int32     `json:"level_number"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Difficulty      string    `json:"difficulty"`
	UnlockRuleType  string    `json:"unlock_rule_type"`
	UnlockRuleValue int32     `json:"unlock_rule_value"`
	Workouts        []Workout `json:"workouts"`
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
	rows, e := s.pool.Query(ctx, `SELECT pl.id::text,pl.level_number,pl.title,pl.description,pl.difficulty,pl.unlock_rule_type,pl.unlock_rule_value,w.id::text,w.title,w.description,w.estimated_minutes,w.difficulty,w.category FROM program_levels pl LEFT JOIN workouts w ON w.program_level_id=pl.id AND w.status='published' WHERE pl.program_id=$1::uuid ORDER BY pl.sort_order,pl.level_number,w.sort_order,w.day_number,w.id`, id)
	if e != nil {
		return Program{}, e
	}
	defer rows.Close()
	item.Workouts = []Workout{}
	item.Levels = []Level{}
	levelIndex := map[string]int{}
	for rows.Next() {
		var level Level
		var workoutID, title, description, difficulty, category *string
		var minutes *int32
		if e = rows.Scan(&level.ID, &level.LevelNumber, &level.Title, &level.Description, &level.Difficulty, &level.UnlockRuleType, &level.UnlockRuleValue, &workoutID, &title, &description, &minutes, &difficulty, &category); e != nil {
			return Program{}, e
		}
		index, exists := levelIndex[level.ID]
		if !exists {
			level.Workouts = []Workout{}
			item.Levels = append(item.Levels, level)
			index = len(item.Levels) - 1
			levelIndex[level.ID] = index
		}
		if workoutID != nil {
			workout := Workout{ID: *workoutID, Title: *title, Description: *description, EstimatedMinutes: *minutes, Difficulty: *difficulty, Category: *category}
			item.Levels[index].Workouts = append(item.Levels[index].Workouts, workout)
			item.Workouts = append(item.Workouts, workout)
		}
	}
	return item, rows.Err()
}
