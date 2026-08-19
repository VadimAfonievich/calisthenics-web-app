package users

import (
	"context"
	"fmt"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/auth"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID                  string   `json:"id"`
	TelegramID          int64    `json:"telegram_id"`
	Username            string   `json:"username,omitempty"`
	FirstName           string   `json:"first_name"`
	LastName            string   `json:"last_name,omitempty"`
	PhotoURL            string   `json:"photo_url,omitempty"`
	DisplayName         string   `json:"display_name"`
	Level               int32    `json:"level"`
	XP                  int32    `json:"xp"`
	CurrentStreak       int32    `json:"current_streak"`
	LongestStreak       int32    `json:"longest_streak"`
	PreferredDifficulty string   `json:"preferred_difficulty"`
	Timezone            string   `json:"timezone"`
	Role                string   `json:"role"`
	AvailableModes      []string `json:"available_modes"`
}

func AvailableModes(role string) []string {
	if role == "coach" || role == "admin" || role == "super_admin" {
		return []string{"student", "coach"}
	}
	return []string{"student"}
}

type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) UpsertTelegramUser(ctx context.Context, telegramUser auth.TelegramUser) (User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return User{}, fmt.Errorf("begin user upsert: %w", err)
	}
	defer tx.Rollback(ctx) // no-op after Commit
	var id pgtype.UUID
	err = tx.QueryRow(ctx, `INSERT INTO users (telegram_id, username, first_name, last_name, photo_url)
VALUES ($1, NULLIF($2, ''), $3, NULLIF($4, ''), NULLIF($5, ''))
ON CONFLICT (telegram_id) DO UPDATE SET username = EXCLUDED.username, first_name = EXCLUDED.first_name, last_name = EXCLUDED.last_name, photo_url = EXCLUDED.photo_url
RETURNING id`, telegramUser.ID, telegramUser.Username, telegramUser.FirstName, telegramUser.LastName, telegramUser.PhotoURL).Scan(&id)
	if err != nil {
		return User{}, fmt.Errorf("upsert user: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO profiles (user_id, display_name) VALUES ($1, $2) ON CONFLICT (user_id) DO NOTHING`, id, telegramUser.FirstName)
	if err != nil {
		return User{}, fmt.Errorf("create profile: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO user_progress (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING`, id)
	if err != nil {
		return User{}, fmt.Errorf("create user progress: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return User{}, fmt.Errorf("commit user upsert: %w", err)
	}
	return s.GetByID(ctx, id.String())
}

func (s *Store) GetByID(ctx context.Context, id string) (User, error) {
	var user User
	err := s.pool.QueryRow(ctx, `SELECT u.id::text, u.telegram_id, COALESCE(u.username, ''), u.first_name, COALESCE(u.last_name, ''), COALESCE(u.photo_url, ''), p.display_name, p.level, p.xp, p.current_streak, p.longest_streak, p.preferred_difficulty, p.timezone, COALESCE(a.role, 'user')
FROM users u JOIN profiles p ON p.user_id = u.id LEFT JOIN admin_users a ON a.user_id=u.id WHERE u.id = $1::uuid`, id).Scan(&user.ID, &user.TelegramID, &user.Username, &user.FirstName, &user.LastName, &user.PhotoURL, &user.DisplayName, &user.Level, &user.XP, &user.CurrentStreak, &user.LongestStreak, &user.PreferredDifficulty, &user.Timezone, &user.Role)
	if err != nil {
		return User{}, err
	}
	user.AvailableModes = AvailableModes(user.Role)
	return user, nil
}
