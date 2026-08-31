package main

import (
	"context"
	"log"
	"os"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/systemworkouts"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err = systemworkouts.Seed(context.Background(), pool); err != nil {
		log.Fatal(err)
	}
	log.Printf("seeded %d system workouts", len(systemworkouts.Catalog))
}
