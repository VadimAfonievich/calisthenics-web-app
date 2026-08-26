package exercises

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("exercise not found")

type Exercise struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Instructions   string     `json:"instructions"`
	CommonMistakes string     `json:"common_mistakes"`
	Difficulty     string     `json:"difficulty"`
	CoachTips      string     `json:"coach_tips,omitempty"`
	CoverMediaURL  string     `json:"cover_media_url,omitempty"`
	MuscleGroups   []string   `json:"muscle_groups"`
	Equipment      []string   `json:"equipment"`
	Tags           []string   `json:"tags"`
	DemoMedia      *DemoMedia `json:"demo_media,omitempty"`
}
type DemoMedia struct {
	URL       string `json:"url"`
	Type      string `json:"type"`
	MIMEType  string `json:"mime_type"`
	PosterURL string `json:"poster_url,omitempty"`
}
type Service struct{ pool *pgxpool.Pool }

func NewService(pool *pgxpool.Pool) *Service { return &Service{pool} }
func (s *Service) List(ctx context.Context, difficulty, muscle, movement, equipment, tag, search string) ([]Exercise, error) {
	rows, e := s.pool.Query(ctx, `SELECT e.id::text,e.name,e.description,e.instructions,e.common_mistakes,e.difficulty,e.muscle_groups,e.equipment,e.tags,e.coach_tips,COALESCE(m.url,''),dm.url,dm.type,dm.mime_type,dm.thumbnail_url FROM exercises e LEFT JOIN media_assets m ON m.id=e.cover_media_id LEFT JOIN media_assets dm ON dm.id=e.demo_media_id WHERE e.status='published' AND ($1='' OR e.difficulty=$1) AND ($2='' OR $2=ANY(e.muscle_groups)) AND ($3='' OR e.movement_type=$3) AND ($4='' OR $4=ANY(e.equipment)) AND ($5='' OR $5=ANY(e.tags)) AND ($6='' OR e.name ILIKE '%'||$6||'%' OR e.description ILIKE '%'||$6||'%') ORDER BY e.difficulty,e.name`, difficulty, muscle, movement, equipment, tag, strings.TrimSpace(search))
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	items := []Exercise{}
	for rows.Next() {
		var x Exercise
		var u, kind, mime, poster *string
		if e = rows.Scan(&x.ID, &x.Name, &x.Description, &x.Instructions, &x.CommonMistakes, &x.Difficulty, &x.MuscleGroups, &x.Equipment, &x.Tags, &x.CoachTips, &x.CoverMediaURL, &u, &kind, &mime, &poster); e != nil {
			return nil, e
		}
		if u != nil {
			x.DemoMedia = &DemoMedia{URL: *u, Type: *kind, MIMEType: *mime}
			if poster != nil {
				x.DemoMedia.PosterURL = *poster
			}
		}
		items = append(items, x)
	}
	return items, rows.Err()
}
func (s *Service) Get(ctx context.Context, id string) (Exercise, error) {
	var x Exercise
	var u, kind, mime, poster *string
	e := s.pool.QueryRow(ctx, `SELECT e.id::text,e.name,e.description,e.instructions,e.common_mistakes,e.difficulty,e.muscle_groups,e.equipment,e.tags,e.coach_tips,COALESCE(m.url,''),dm.url,dm.type,dm.mime_type,dm.thumbnail_url FROM exercises e LEFT JOIN media_assets m ON m.id=e.cover_media_id LEFT JOIN media_assets dm ON dm.id=e.demo_media_id WHERE e.id=$1::uuid AND e.status='published'`, id).Scan(&x.ID, &x.Name, &x.Description, &x.Instructions, &x.CommonMistakes, &x.Difficulty, &x.MuscleGroups, &x.Equipment, &x.Tags, &x.CoachTips, &x.CoverMediaURL, &u, &kind, &mime, &poster)
	if e == nil && u != nil {
		x.DemoMedia = &DemoMedia{URL: *u, Type: *kind, MIMEType: *mime}
		if poster != nil {
			x.DemoMedia.PosterURL = *poster
		}
	}
	if errors.Is(e, pgx.ErrNoRows) {
		return Exercise{}, ErrNotFound
	}
	return x, e
}
