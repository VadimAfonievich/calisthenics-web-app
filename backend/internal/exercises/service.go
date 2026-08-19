package exercises

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("exercise not found")

type Exercise struct {
	ID, Name, Description, Instructions, CommonMistakes, Difficulty string
	MuscleGroups, Equipment                                         []string
}
type Service struct{ pool *pgxpool.Pool }

func NewService(pool *pgxpool.Pool) *Service { return &Service{pool} }
func (s *Service) List(ctx context.Context, difficulty, muscle string) ([]Exercise, error) {
	rows, e := s.pool.Query(ctx, `SELECT id::text,name,description,instructions,common_mistakes,difficulty,muscle_groups,equipment FROM exercises WHERE status='published' AND ($1='' OR difficulty=$1) AND ($2='' OR $2=ANY(muscle_groups)) ORDER BY difficulty,name`, difficulty, muscle)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	items := []Exercise{}
	for rows.Next() {
		var x Exercise
		if e = rows.Scan(&x.ID, &x.Name, &x.Description, &x.Instructions, &x.CommonMistakes, &x.Difficulty, &x.MuscleGroups, &x.Equipment); e != nil {
			return nil, e
		}
		items = append(items, x)
	}
	return items, rows.Err()
}
func (s *Service) Get(ctx context.Context, id string) (Exercise, error) {
	var x Exercise
	e := s.pool.QueryRow(ctx, `SELECT id::text,name,description,instructions,common_mistakes,difficulty,muscle_groups,equipment FROM exercises WHERE id=$1::uuid AND status='published'`, id).Scan(&x.ID, &x.Name, &x.Description, &x.Instructions, &x.CommonMistakes, &x.Difficulty, &x.MuscleGroups, &x.Equipment)
	if errors.Is(e, pgx.ErrNoRows) {
		return Exercise{}, ErrNotFound
	}
	return x, e
}
