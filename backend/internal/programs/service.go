package programs

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("program not found")

type Program struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	Description   string `json:"description"`
	Difficulty    string `json:"difficulty"`
	DurationWeeks int32  `json:"duration_weeks"`
	Category      string `json:"category"`
	CoverMediaURL string `json:"cover_media_url,omitempty"`
}

type Service struct{ pool *pgxpool.Pool }

func NewService(pool *pgxpool.Pool) *Service { return &Service{pool: pool} }

func (s *Service) List(ctx context.Context) ([]Program, error) {
	rows, err := s.pool.Query(ctx, `SELECT p.id::text,p.name,p.slug,p.description,p.difficulty,p.duration_weeks,p.category,COALESCE(m.url,'') FROM programs p LEFT JOIN media_assets m ON m.id=p.cover_media_id WHERE p.published ORDER BY p.category,p.difficulty,p.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Program{}
	for rows.Next() {
		var item Program
		if err = rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Description, &item.Difficulty, &item.DurationWeeks, &item.Category, &item.CoverMediaURL); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) Get(ctx context.Context, id string) (Program, error) {
	var item Program
	err := s.pool.QueryRow(ctx, `SELECT p.id::text,p.name,p.slug,p.description,p.difficulty,p.duration_weeks,p.category,COALESCE(m.url,'') FROM programs p LEFT JOIN media_assets m ON m.id=p.cover_media_id WHERE p.id=$1::uuid AND p.published`, id).Scan(&item.ID, &item.Name, &item.Slug, &item.Description, &item.Difficulty, &item.DurationWeeks, &item.Category, &item.CoverMediaURL)
	if errors.Is(err, pgx.ErrNoRows) {
		return Program{}, ErrNotFound
	}
	if err != nil {
		return Program{}, err
	}
	return item, nil
}
