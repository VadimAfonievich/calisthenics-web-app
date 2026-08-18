package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/config"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/exercises"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/handler"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/lessons"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/users"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration error", "error", err)
		os.Exit(1)
	}

	logger := config.NewLogger(cfg.LogLevel)
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("create postgres pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pingPostgres(ctx, pool); err != nil {
		logger.Error("postgres unavailable", "error", err)
		os.Exit(1)
	}

	redisOptions, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		logger.Error("parse redis URL", "error", err)
		os.Exit(1)
	}
	redisClient := redis.NewClient(redisOptions)
	defer redisClient.Close()

	if err := pingRedis(ctx, redisClient); err != nil {
		logger.Error("redis unavailable", "error", err)
		os.Exit(1)
	}

	router := handler.NewRouter(handler.Dependencies{
		Auth:      handler.AuthDependencies{BotToken: cfg.TelegramBotToken, JWTSecret: cfg.JWTSecret, Users: users.NewStore(pool)},
		Lessons:   lessons.NewService(pool),
		Exercises: exercises.NewService(pool),
		Health: handler.HealthDependencies{
			Postgres: func(checkContext context.Context) error { return pingPostgres(checkContext, pool) },
			Redis:    func(checkContext context.Context) error { return pingRedis(checkContext, redisClient) },
		},
		Logger:      logger,
		CORSOrigins: cfg.CORSOrigins,
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           middleware.WithRequestLogging(logger, router),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("server started", "address", server.Addr, "environment", cfg.AppEnv)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	signalContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-signalContext.Done()

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
}

func pingPostgres(ctx context.Context, pool *pgxpool.Pool) error {
	pingContext, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return pool.Ping(pingContext)
}

func pingRedis(ctx context.Context, client *redis.Client) error {
	pingContext, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return client.Ping(pingContext).Err()
}
