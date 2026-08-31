package programs

import (
	"context"
	"errors"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/access"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("program not found")
var ErrForbidden = errors.New("program unavailable")
var ErrStageLocked = errors.New("program stage is locked")

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
	ID                 string    `json:"id"`
	LevelNumber        int32     `json:"level_number"`
	Title              string    `json:"title"`
	Description        string    `json:"description"`
	Difficulty         string    `json:"difficulty"`
	UnlockRuleType     string    `json:"unlock_rule_type"`
	UnlockRuleValue    int32     `json:"unlock_rule_value"`
	Workouts           []Workout `json:"workouts"`
	Status             string    `json:"status,omitempty"`
	MasteryType        string    `json:"mastery_type"`
	MasteryValue       *int32    `json:"mastery_value,omitempty"`
	MasteryName        string    `json:"mastery_name"`
	MasteryDescription string    `json:"mastery_description"`
	MasteredAt         *string   `json:"mastered_at,omitempty"`
}
type MasteryResult struct {
	ProgramID        string `json:"program_id"`
	ProgramLevelID   string `json:"program_level_id"`
	Status           string `json:"status"`
	CurrentLevel     int32  `json:"current_level"`
	MasteredAt       string `json:"mastered_at"`
	ProgramCompleted bool   `json:"program_completed"`
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
	tenant, ok := middleware.TenantID(ctx)
	if !ok {
		return []Program{}, nil
	}
	rows, err := s.pool.Query(ctx, `SELECT p.id::text,p.name,p.slug,p.description,p.difficulty,p.duration_weeks,p.category,COALESCE(m.url,''),(SELECT count(*)::int FROM workouts w WHERE w.program_id=p.id AND w.tenant_id=$2::uuid AND w.status='published'),COALESCE(upp.status,''),COALESCE(upp.current_level,0),COALESCE((SELECT title FROM program_levels WHERE program_id=p.id AND level_number=upp.current_level),'') FROM programs p LEFT JOIN media_assets m ON m.id=p.cover_media_id LEFT JOIN user_program_progress upp ON upp.program_id=p.id AND upp.user_id=$1::uuid AND upp.tenant_id=$2::uuid WHERE p.tenant_id=$2::uuid AND p.published ORDER BY (upp.status='active') DESC,upp.updated_at DESC NULLS LAST,p.category,p.difficulty,p.name`, user, tenant)
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
	tenant, ok := middleware.TenantID(ctx)
	if !ok {
		return Program{}, ErrNotFound
	}
	var item Program
	err := s.pool.QueryRow(ctx, `SELECT p.id::text,p.name,p.slug,p.description,p.difficulty,p.duration_weeks,p.category,COALESCE(m.url,''),(SELECT count(*)::int FROM workouts w WHERE w.program_id=p.id AND w.tenant_id=$3::uuid AND w.status='published'),COALESCE(upp.status,''),COALESCE(upp.current_level,0),COALESCE((SELECT title FROM program_levels WHERE program_id=p.id AND level_number=upp.current_level),'') FROM programs p LEFT JOIN media_assets m ON m.id=p.cover_media_id LEFT JOIN user_program_progress upp ON upp.program_id=p.id AND upp.user_id=$2::uuid AND upp.tenant_id=$3::uuid WHERE p.id=$1::uuid AND p.tenant_id=$3::uuid AND p.published`, id, user, tenant).Scan(&item.ID, &item.Name, &item.Slug, &item.Description, &item.Difficulty, &item.DurationWeeks, &item.Category, &item.CoverMediaURL, &item.WorkoutCount, &item.ProgressStatus, &item.CurrentLevel, &item.CurrentStage)
	if errors.Is(err, pgx.ErrNoRows) {
		return Program{}, ErrNotFound
	}
	if err != nil {
		return Program{}, err
	}
	rows, e := s.pool.Query(ctx, `SELECT pl.id::text,pl.level_number,pl.title,pl.description,pl.difficulty,pl.unlock_rule_type,pl.unlock_rule_value,pl.mastery_type,pl.mastery_value,pl.mastery_name,pl.mastery_description,mastery.mastered_at::text,w.id::text,w.title,w.description,w.estimated_minutes,w.difficulty,w.category FROM program_levels pl JOIN programs p ON p.id=pl.program_id LEFT JOIN user_program_level_mastery mastery ON mastery.program_level_id=pl.id AND mastery.program_id=p.id AND mastery.user_id=$3::uuid AND mastery.tenant_id=$2::uuid LEFT JOIN workouts w ON w.program_level_id=pl.id AND w.tenant_id=$2::uuid AND w.status='published' WHERE pl.program_id=$1::uuid AND p.tenant_id=$2::uuid ORDER BY pl.sort_order,pl.level_number,w.sort_order,w.day_number,w.id`, id, tenant, user)
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
		if e = rows.Scan(&level.ID, &level.LevelNumber, &level.Title, &level.Description, &level.Difficulty, &level.UnlockRuleType, &level.UnlockRuleValue, &level.MasteryType, &level.MasteryValue, &level.MasteryName, &level.MasteryDescription, &level.MasteredAt, &workoutID, &title, &description, &minutes, &difficulty, &category); e != nil {
			return Program{}, e
		}
		index, exists := levelIndex[level.ID]
		if !exists {
			level.Workouts = []Workout{}
			if level.MasteredAt != nil {
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
				copy := workout
				if item.NextWorkout == nil {
					item.NextWorkout = &copy
				}
			}
		}
	}
	return item, rows.Err()
}

func (s *Service) Start(ctx context.Context, user, id string) (Progress, error) {
	tenant, ok := middleware.TenantID(ctx)
	if !ok {
		return Progress{}, ErrNotFound
	}
	var out Progress
	var owner *string
	if err := s.pool.QueryRow(ctx, `SELECT owner_user_id::text FROM programs WHERE id=$1::uuid AND tenant_id=$2::uuid AND published`, id, tenant).Scan(&owner); errors.Is(err, pgx.ErrNoRows) {
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
	err = s.pool.QueryRow(ctx, `INSERT INTO user_program_progress(user_id,tenant_id,program_id,current_level,status) SELECT $1::uuid,$3::uuid,p.id,COALESCE((SELECT min(level_number) FROM program_levels WHERE program_id=p.id),1),'active' FROM programs p WHERE p.id=$2::uuid AND p.tenant_id=$3::uuid AND p.published ON CONFLICT(user_id,tenant_id,program_id) DO UPDATE SET updated_at=user_program_progress.updated_at RETURNING program_id::text,status,current_level,started_at::text,completed_at::text`, user, id, tenant).Scan(&out.ProgramID, &out.Status, &out.CurrentLevel, &out.StartedAt, &out.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, ErrNotFound
	}
	return out, err
}

func (s *Service) ConfirmMastery(ctx context.Context, user, program, levelID string) (MasteryResult, error) {
	tenant, ok := middleware.TenantID(ctx)
	if !ok {
		return MasteryResult{}, ErrNotFound
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return MasteryResult{}, err
	}
	defer tx.Rollback(ctx)
	var result MasteryResult
	var levelNumber int32
	err = tx.QueryRow(ctx, `SELECT m.mastered_at::text,upp.status,upp.current_level FROM user_program_level_mastery m JOIN user_program_progress upp ON upp.user_id=m.user_id AND upp.tenant_id=m.tenant_id AND upp.program_id=m.program_id WHERE m.user_id=$1::uuid AND m.tenant_id=$2::uuid AND m.program_id=$3::uuid AND m.program_level_id=$4::uuid`, user, tenant, program, levelID).Scan(&result.MasteredAt, &result.Status, &result.CurrentLevel)
	if err == nil {
		result.ProgramID = program
		result.ProgramLevelID = levelID
		result.ProgramCompleted = result.Status == "completed"
		return result, tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return result, err
	}
	err = tx.QueryRow(ctx, `SELECT pl.level_number FROM program_levels pl JOIN programs p ON p.id=pl.program_id JOIN user_program_progress upp ON upp.program_id=p.id AND upp.user_id=$1::uuid AND upp.tenant_id=$4::uuid WHERE p.id=$2::uuid AND pl.id=$3::uuid AND p.tenant_id=$4::uuid AND p.published AND upp.status='active' AND upp.current_level=pl.level_number FOR UPDATE OF upp`, user, program, levelID, tenant).Scan(&levelNumber)
	if errors.Is(err, pgx.ErrNoRows) {
		return result, ErrStageLocked
	}
	if err != nil {
		return result, err
	}
	err = tx.QueryRow(ctx, `INSERT INTO user_program_level_mastery(user_id,tenant_id,program_id,program_level_id) VALUES($1::uuid,$2::uuid,$3::uuid,$4::uuid) ON CONFLICT(user_id,tenant_id,program_id,program_level_id) DO UPDATE SET mastered_at=user_program_level_mastery.mastered_at RETURNING mastered_at::text`, user, tenant, program, levelID).Scan(&result.MasteredAt)
	if err != nil {
		return result, err
	}
	var next *int32
	err = tx.QueryRow(ctx, `SELECT min(level_number)::int FROM program_levels WHERE program_id=$1::uuid AND level_number>$2`, program, levelNumber).Scan(&next)
	if err != nil {
		return result, err
	}
	if next == nil {
		result.ProgramCompleted = true
		result.Status = "completed"
		result.CurrentLevel = levelNumber
		_, err = tx.Exec(ctx, `UPDATE user_program_progress SET status='completed',completed_at=COALESCE(completed_at,NOW()) WHERE user_id=$1::uuid AND tenant_id=$2::uuid AND program_id=$3::uuid`, user, tenant, program)
	} else {
		result.Status = "active"
		result.CurrentLevel = *next
		_, err = tx.Exec(ctx, `UPDATE user_program_progress SET current_level=$4 WHERE user_id=$1::uuid AND tenant_id=$2::uuid AND program_id=$3::uuid`, user, tenant, program, *next)
	}
	if err != nil {
		return result, err
	}
	result.ProgramID = program
	result.ProgramLevelID = levelID
	return result, tx.Commit(ctx)
}
