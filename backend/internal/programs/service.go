package programs

import (
	"context"
	"errors"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/access"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("program not found")
var ErrForbidden = errors.New("program unavailable")

type Progress struct {
	ProgramID    string  `json:"program_id"`
	Status       string  `json:"status"`
	CurrentLevel int32   `json:"current_level"`
	StartedAt    string  `json:"started_at"`
	CompletedAt  *string `json:"completed_at,omitempty"`
}

type Program struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description"`
	Difficulty     string    `json:"difficulty"`
	DurationWeeks  int32     `json:"duration_weeks"`
	Category       string    `json:"category"`
	CoverMediaURL  string    `json:"cover_media_url,omitempty"`
	WorkoutCount   int32     `json:"workout_count"`
	Workouts       []Workout `json:"workouts,omitempty"`
	Levels         []Level   `json:"levels,omitempty"`
	ProgressStatus string    `json:"progress_status,omitempty"`
	CurrentLevel   int32     `json:"current_level,omitempty"`
	CurrentStage   string    `json:"current_stage,omitempty"`
	NextWorkout    *Workout  `json:"next_workout,omitempty"`
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
	Status          string    `json:"status,omitempty"`
}
type Workout struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	EstimatedMinutes int32  `json:"estimated_minutes"`
	Difficulty       string `json:"difficulty"`
	Category         string `json:"category"`
}

type Service struct {
	pool   *pgxpool.Pool
	access access.Service
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool, access: access.PublishedContent{}}
}

func (s *Service) List(ctx context.Context, user string) ([]Program, error) {
	rows, err := s.pool.Query(ctx, `SELECT p.id::text,p.name,p.slug,p.description,p.difficulty,p.duration_weeks,p.category,COALESCE(m.url,''),(SELECT count(*)::int FROM workouts w WHERE w.program_id=p.id AND w.status='published'),COALESCE(upp.status,''),COALESCE(upp.current_level,0),COALESCE((SELECT title FROM program_levels WHERE program_id=p.id AND level_number=upp.current_level),'') FROM programs p LEFT JOIN media_assets m ON m.id=p.cover_media_id LEFT JOIN user_program_progress upp ON upp.program_id=p.id AND upp.user_id=$1::uuid WHERE p.published ORDER BY (upp.status='active') DESC,upp.updated_at DESC NULLS LAST,p.category,p.difficulty,p.name`, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Program{}
	for rows.Next() {
		var item Program
		if err = rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Description, &item.Difficulty, &item.DurationWeeks, &item.Category, &item.CoverMediaURL, &item.WorkoutCount, &item.ProgressStatus, &item.CurrentLevel, &item.CurrentStage); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) Get(ctx context.Context, user, id string) (Program, error) {
	_ = s.refresh(ctx, user, id)
	var item Program
	err := s.pool.QueryRow(ctx, `SELECT p.id::text,p.name,p.slug,p.description,p.difficulty,p.duration_weeks,p.category,COALESCE(m.url,''),(SELECT count(*)::int FROM workouts w WHERE w.program_id=p.id AND w.status='published'),COALESCE(upp.status,''),COALESCE(upp.current_level,0),COALESCE((SELECT title FROM program_levels WHERE program_id=p.id AND level_number=upp.current_level),'') FROM programs p LEFT JOIN media_assets m ON m.id=p.cover_media_id LEFT JOIN user_program_progress upp ON upp.program_id=p.id AND upp.user_id=$2::uuid WHERE p.id=$1::uuid AND p.published`, id, user).Scan(&item.ID, &item.Name, &item.Slug, &item.Description, &item.Difficulty, &item.DurationWeeks, &item.Category, &item.CoverMediaURL, &item.WorkoutCount, &item.ProgressStatus, &item.CurrentLevel, &item.CurrentStage)
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
			if item.ProgressStatus == "completed" || (item.ProgressStatus == "active" && level.LevelNumber < item.CurrentLevel) {
				level.Status = "completed"
			} else if item.ProgressStatus == "active" && level.LevelNumber == item.CurrentLevel {
				level.Status = "current"
			} else {
				level.Status = "locked"
			}
			item.Levels = append(item.Levels, level)
			index = len(item.Levels) - 1
			levelIndex[level.ID] = index
		}
		if workoutID != nil {
			workout := Workout{ID: *workoutID, Title: *title, Description: *description, EstimatedMinutes: *minutes, Difficulty: *difficulty, Category: *category}
			item.Levels[index].Workouts = append(item.Levels[index].Workouts, workout)
			item.Workouts = append(item.Workouts, workout)
			if level.Status == "current" && item.NextWorkout == nil {
				var done bool
				_ = s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM workout_sessions WHERE user_id=$1::uuid AND workout_id=$2::uuid AND status='completed')`, user, workout.ID).Scan(&done)
				if !done {
					copy := workout
					item.NextWorkout = &copy
				}
			}
		}
	}
	return item, rows.Err()
}

func (s *Service) Start(ctx context.Context, user, id string) (Progress, error) {
	var out Progress
	var owner *string
	if err := s.pool.QueryRow(ctx, `SELECT owner_user_id::text FROM programs WHERE id=$1::uuid AND published`, id).Scan(&owner); errors.Is(err, pgx.ErrNoRows) {
		return out, ErrNotFound
	} else if err != nil {
		return out, err
	}
	allowed, err := s.access.CanAccessContent(ctx, user, owner)
	if err != nil {
		return out, err
	}
	if !allowed {
		return out, ErrForbidden
	}
	err = s.pool.QueryRow(ctx, `INSERT INTO user_program_progress(user_id,program_id,current_level,status) SELECT $1::uuid,p.id,COALESCE((SELECT min(level_number) FROM program_levels WHERE program_id=p.id),1),'active' FROM programs p WHERE p.id=$2::uuid AND p.published ON CONFLICT(user_id,program_id) DO UPDATE SET updated_at=user_program_progress.updated_at RETURNING program_id::text,status,current_level,started_at::text,completed_at::text`, user, id).Scan(&out.ProgramID, &out.Status, &out.CurrentLevel, &out.StartedAt, &out.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, ErrNotFound
	}
	return out, err
}

func (s *Service) refresh(ctx context.Context, user, program string) error {
	var active bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM user_program_progress WHERE user_id=$1::uuid AND program_id=$2::uuid AND status='active')`, user, program).Scan(&active); err != nil || !active {
		return err
	}
	rows, err := s.pool.Query(ctx, `SELECT pl.level_number,count(w.id)::int,count(DISTINCT ws.workout_id)::int FROM program_levels pl LEFT JOIN workouts w ON w.program_level_id=pl.id AND w.status='published' LEFT JOIN workout_sessions ws ON ws.workout_id=w.id AND ws.user_id=$1::uuid AND ws.status='completed' WHERE pl.program_id=$2::uuid GROUP BY pl.id ORDER BY pl.sort_order,pl.level_number`, user, program)
	if err != nil {
		return err
	}
	defer rows.Close()
	var current int32
	allComplete := true
	for rows.Next() {
		var level, total, done int32
		if err = rows.Scan(&level, &total, &done); err != nil {
			return err
		}
		if current == 0 && (total == 0 || done < total) {
			current = level
			allComplete = false
		}
	}
	if err = rows.Err(); err != nil {
		return err
	}
	if allComplete {
		_, err = s.pool.Exec(ctx, `UPDATE user_program_progress SET status='completed',completed_at=COALESCE(completed_at,NOW()) WHERE user_id=$1::uuid AND program_id=$2::uuid AND status='active'`, user, program)
	} else {
		_, err = s.pool.Exec(ctx, `UPDATE user_program_progress SET current_level=$3 WHERE user_id=$1::uuid AND program_id=$2::uuid AND status='active'`, user, program, current)
	}
	return err
}
