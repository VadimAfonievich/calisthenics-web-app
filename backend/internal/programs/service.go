package programs

import (
	"context"
	"errors"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
}

type Service struct{ queries *sqlc.Queries }

func NewService(pool *pgxpool.Pool) *Service { return &Service{queries: sqlc.New(pool)} }

func programFromRow(row sqlc.Program) Program {
	id, _ := row.ID.Value()
	return Program{ID: id.(string), Name: row.Name, Slug: row.Slug, Description: row.Description, Difficulty: row.Difficulty, DurationWeeks: row.DurationWeeks}
}

func (s *Service) List(ctx context.Context) ([]Program, error) {
	rows, err := s.queries.ListPublishedPrograms(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]Program, 0, len(rows))
	for _, row := range rows {
		items = append(items, programFromRow(row))
	}
	return items, nil
}

func (s *Service) Get(ctx context.Context, id string) (Program, error) {
	var uuid pgtype.UUID
	if err := uuid.Scan(id); err != nil {
		return Program{}, ErrNotFound
	}
	row, err := s.queries.GetProgramByID(ctx, uuid)
	if errors.Is(err, pgx.ErrNoRows) {
		return Program{}, ErrNotFound
	}
	if err != nil {
		return Program{}, err
	}
	return programFromRow(row), nil
}
