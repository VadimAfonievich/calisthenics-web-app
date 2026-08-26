package exercises

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Run explicitly against an isolated migrated database populated by the standard importer.
func TestStandardLibraryFiltersIntegration(t *testing.T) {
	databaseURL := os.Getenv("STANDARD_EXERCISES_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("STANDARD_EXERCISES_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	service := NewService(pool)
	cases := []struct {
		name string
		args [6]string
	}{
		{"difficulty", [6]string{"beginner"}},
		{"muscle_group", [6]string{"", "chest"}},
		{"movement_type", [6]string{"", "", "duration"}},
		{"equipment", [6]string{"", "", "", "pull-up-bar"}},
		{"tag", [6]string{"", "", "", "", "handstand"}},
		{"search", [6]string{"", "", "", "", "", "стойка"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			items, err := service.List(ctx, tc.args[0], tc.args[1], tc.args[2], tc.args[3], tc.args[4], tc.args[5])
			if err != nil {
				t.Fatal(err)
			}
			if len(items) == 0 {
				t.Fatal("expected at least one exercise")
			}
		})
	}
}
