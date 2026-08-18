package admin

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("content not found")

type Lesson struct {
	CategoryID       string `json:"category_id"`
	Title            string `json:"title"`
	Slug             string `json:"slug"`
	ShortDescription string `json:"short_description"`
	Content          string `json:"content"`
	Difficulty       string `json:"difficulty"`
	DurationMinutes  int    `json:"duration_minutes"`
	SortOrder        int    `json:"sort_order"`
	Published        bool   `json:"published"`
}
type Exercise struct {
	Name           string   `json:"name"`
	Slug           string   `json:"slug"`
	Description    string   `json:"description"`
	Instructions   string   `json:"instructions"`
	CommonMistakes string   `json:"common_mistakes"`
	Difficulty     string   `json:"difficulty"`
	MuscleGroups   []string `json:"muscle_groups"`
	Equipment      []string `json:"equipment"`
	VideoURL       *string  `json:"video_url"`
	ImageURL       *string  `json:"image_url"`
}
type Program struct {
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	Description   string `json:"description"`
	Difficulty    string `json:"difficulty"`
	DurationWeeks int    `json:"duration_weeks"`
	Published     bool   `json:"published"`
}
type Workout struct {
	ProgramID        string `json:"program_id"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	DayNumber        int    `json:"day_number"`
	EstimatedMinutes int    `json:"estimated_minutes"`
	SortOrder        int    `json:"sort_order"`
}
type WorkoutExercise struct {
	ExerciseID            string  `json:"exercise_id"`
	Sets                  int     `json:"sets"`
	RestSeconds           int     `json:"rest_seconds"`
	SortOrder             int     `json:"sort_order"`
	TargetReps            *int    `json:"target_reps"`
	TargetDurationSeconds *int    `json:"target_duration_seconds"`
	Notes                 *string `json:"notes"`
}
type Service struct{ pool *pgxpool.Pool }

func NewService(pool *pgxpool.Pool) *Service { return &Service{pool: pool} }
func (s *Service) IsAdmin(ctx context.Context, userID string) (bool, error) {
	var ok bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM admin_users WHERE user_id=$1::uuid)`, userID).Scan(&ok)
	return ok, err
}
func execOne(ctx context.Context, p *pgxpool.Pool, q string, args ...any) (string, error) {
	var id string
	err := p.QueryRow(ctx, q, args...).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return id, err
}
func (s *Service) CreateLesson(c context.Context, x Lesson) (string, error) {
	return execOne(c, s.pool, `INSERT INTO lessons(category_id,title,slug,short_description,content,difficulty,duration_minutes,sort_order,published) VALUES($1::uuid,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id::text`, x.CategoryID, x.Title, x.Slug, x.ShortDescription, x.Content, x.Difficulty, x.DurationMinutes, x.SortOrder, x.Published)
}
func (s *Service) UpdateLesson(c context.Context, id string, x Lesson) (string, error) {
	return execOne(c, s.pool, `UPDATE lessons SET category_id=$2::uuid,title=$3,slug=$4,short_description=$5,content=$6,difficulty=$7,duration_minutes=$8,sort_order=$9,published=$10 WHERE id=$1::uuid RETURNING id::text`, id, x.CategoryID, x.Title, x.Slug, x.ShortDescription, x.Content, x.Difficulty, x.DurationMinutes, x.SortOrder, x.Published)
}
func (s *Service) PublishLesson(c context.Context, id string, published bool) (string, error) {
	return execOne(c, s.pool, `UPDATE lessons SET published=$2 WHERE id=$1::uuid RETURNING id::text`, id, published)
}
func (s *Service) DeleteLesson(c context.Context, id string) error {
	tag, e := s.pool.Exec(c, `DELETE FROM lessons WHERE id=$1::uuid`, id)
	if e == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return e
}
func (s *Service) CreateExercise(c context.Context, x Exercise) (string, error) {
	return execOne(c, s.pool, `INSERT INTO exercises(name,slug,description,instructions,common_mistakes,difficulty,muscle_groups,equipment,video_url,image_url) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id::text`, x.Name, x.Slug, x.Description, x.Instructions, x.CommonMistakes, x.Difficulty, x.MuscleGroups, x.Equipment, x.VideoURL, x.ImageURL)
}
func (s *Service) UpdateExercise(c context.Context, id string, x Exercise) (string, error) {
	return execOne(c, s.pool, `UPDATE exercises SET name=$2,slug=$3,description=$4,instructions=$5,common_mistakes=$6,difficulty=$7,muscle_groups=$8,equipment=$9,video_url=$10,image_url=$11 WHERE id=$1::uuid RETURNING id::text`, id, x.Name, x.Slug, x.Description, x.Instructions, x.CommonMistakes, x.Difficulty, x.MuscleGroups, x.Equipment, x.VideoURL, x.ImageURL)
}
func (s *Service) DeleteExercise(c context.Context, id string) error {
	tag, e := s.pool.Exec(c, `DELETE FROM exercises WHERE id=$1::uuid`, id)
	if e == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return e
}
func (s *Service) CreateProgram(c context.Context, x Program) (string, error) {
	return execOne(c, s.pool, `INSERT INTO programs(name,slug,description,difficulty,duration_weeks,published) VALUES($1,$2,$3,$4,$5,$6) RETURNING id::text`, x.Name, x.Slug, x.Description, x.Difficulty, x.DurationWeeks, x.Published)
}
func (s *Service) UpdateProgram(c context.Context, id string, x Program) (string, error) {
	return execOne(c, s.pool, `UPDATE programs SET name=$2,slug=$3,description=$4,difficulty=$5,duration_weeks=$6,published=$7 WHERE id=$1::uuid RETURNING id::text`, id, x.Name, x.Slug, x.Description, x.Difficulty, x.DurationWeeks, x.Published)
}
func (s *Service) PublishProgram(c context.Context, id string, published bool) (string, error) {
	return execOne(c, s.pool, `UPDATE programs SET published=$2 WHERE id=$1::uuid RETURNING id::text`, id, published)
}
func (s *Service) DeleteProgram(c context.Context, id string) error {
	tag, e := s.pool.Exec(c, `DELETE FROM programs WHERE id=$1::uuid`, id)
	if e == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return e
}
func (s *Service) CreateWorkout(c context.Context, x Workout) (string, error) {
	return execOne(c, s.pool, `INSERT INTO workouts(program_id,day_number,title,description,estimated_minutes,sort_order) VALUES($1::uuid,$2,$3,$4,$5,$6) RETURNING id::text`, x.ProgramID, x.DayNumber, x.Title, x.Description, x.EstimatedMinutes, x.SortOrder)
}
func (s *Service) UpdateWorkout(c context.Context, id string, x Workout) (string, error) {
	return execOne(c, s.pool, `UPDATE workouts SET program_id=$2::uuid,day_number=$3,title=$4,description=$5,estimated_minutes=$6,sort_order=$7 WHERE id=$1::uuid RETURNING id::text`, id, x.ProgramID, x.DayNumber, x.Title, x.Description, x.EstimatedMinutes, x.SortOrder)
}
func (s *Service) DeleteWorkout(c context.Context, id string) error {
	tag, e := s.pool.Exec(c, `DELETE FROM workouts WHERE id=$1::uuid`, id)
	if e == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return e
}
func (s *Service) UpsertWorkoutExercise(c context.Context, workoutID string, x WorkoutExercise) (string, error) {
	if (x.TargetReps == nil) == (x.TargetDurationSeconds == nil) {
		return "", fmt.Errorf("exactly one target is required")
	}
	return execOne(c, s.pool, `INSERT INTO workout_exercises(workout_id,exercise_id,sets,target_reps,target_duration_seconds,rest_seconds,sort_order,notes) VALUES($1::uuid,$2::uuid,$3,$4,$5,$6,$7,$8) ON CONFLICT(workout_id,exercise_id) DO UPDATE SET sets=EXCLUDED.sets,target_reps=EXCLUDED.target_reps,target_duration_seconds=EXCLUDED.target_duration_seconds,rest_seconds=EXCLUDED.rest_seconds,sort_order=EXCLUDED.sort_order,notes=EXCLUDED.notes RETURNING id::text`, workoutID, x.ExerciseID, x.Sets, x.TargetReps, x.TargetDurationSeconds, x.RestSeconds, x.SortOrder, x.Notes)
}
func (s *Service) DeleteWorkoutExercise(c context.Context, workoutID, exerciseID string) error {
	tag, e := s.pool.Exec(c, `DELETE FROM workout_exercises WHERE workout_id=$1::uuid AND exercise_id=$2::uuid`, workoutID, exerciseID)
	if e == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return e
}
