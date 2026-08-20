package exercises

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("exercise not found")

type Exercise struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Instructions   string   `json:"instructions"`
	CommonMistakes string   `json:"common_mistakes"`
	Difficulty     string   `json:"difficulty"`
	CoachTips      string   `json:"coach_tips,omitempty"`
	CoverMediaURL  string   `json:"cover_media_url,omitempty"`
	MuscleGroups   []string `json:"muscle_groups"`
	Equipment      []string `json:"equipment"`
}
type Service struct{ pool *pgxpool.Pool }

func NewService(pool *pgxpool.Pool) *Service { return &Service{pool} }
func (s *Service) List(ctx context.Context, difficulty, muscle string) ([]Exercise, error) {
	rows, e := s.pool.Query(ctx, `SELECT e.id::text,e.name,e.description,e.instructions,e.common_mistakes,e.difficulty,e.muscle_groups,e.equipment,e.coach_tips,COALESCE(m.url,'') FROM exercises e LEFT JOIN media_assets m ON m.id=e.cover_media_id WHERE e.status='published' AND ($1='' OR e.difficulty=$1) AND ($2='' OR $2=ANY(e.muscle_groups)) ORDER BY e.difficulty,e.name`, difficulty, muscle)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	items := []Exercise{}
	for rows.Next() {
		var x Exercise
		if e = rows.Scan(&x.ID, &x.Name, &x.Description, &x.Instructions, &x.CommonMistakes, &x.Difficulty, &x.MuscleGroups, &x.Equipment, &x.CoachTips, &x.CoverMediaURL); e != nil {
			return nil, e
		}
		items = append(items, x)
	}
	return items, rows.Err()
}
func (s *Service) Get(ctx context.Context, id string) (Exercise, error) {
	var x Exercise
	e := s.pool.QueryRow(ctx, `SELECT e.id::text,e.name,e.description,e.instructions,e.common_mistakes,e.difficulty,e.muscle_groups,e.equipment,e.coach_tips,COALESCE(m.url,'') FROM exercises e LEFT JOIN media_assets m ON m.id=e.cover_media_id WHERE e.id=$1::uuid AND e.status='published'`, id).Scan(&x.ID, &x.Name, &x.Description, &x.Instructions, &x.CommonMistakes, &x.Difficulty, &x.MuscleGroups, &x.Equipment, &x.CoachTips, &x.CoverMediaURL)
	if errors.Is(e, pgx.ErrNoRows) {
		return Exercise{}, ErrNotFound
	}
	return x, e
}
