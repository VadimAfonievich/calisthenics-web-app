package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	AppEnv                 string
	CORSOrigins            []string
	DatabaseURL            string
	JWTSecret              string
	LogLevel               slog.Level
	Port                   string
	RedisURL               string
	TelegramBotToken       string
	TelegramBotUsername    string
	TelegramWebAppURL      string
	ObjectStorageProvider  string
	ObjectStorageEndpoint  string
	ObjectStorageRegion    string
	ObjectStorageBucket    string
	ObjectStorageAccessKey string
	ObjectStorageSecretKey string
	ObjectStoragePublicURL string
	MaxImageUploadBytes    int64
	MaxVideoUploadBytes    int64
}

func Load() (Config, error) {
	logLevel := new(slog.Level)
	if err := logLevel.UnmarshalText([]byte(envOrDefault("LOG_LEVEL", "INFO"))); err != nil {
		return Config{}, fmt.Errorf("invalid LOG_LEVEL: %w", err)
	}

	cfg := Config{
		AppEnv:                 envOrDefault("APP_ENV", "development"),
		CORSOrigins:            splitCSV(envOrDefault("CORS_ORIGINS", "http://localhost:5173")),
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		JWTSecret:              os.Getenv("JWT_SECRET"),
		LogLevel:               *logLevel,
		Port:                   envOrDefault("PORT", "8080"),
		RedisURL:               os.Getenv("REDIS_URL"),
		TelegramBotToken:       os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramBotUsername:    os.Getenv("TELEGRAM_BOT_USERNAME"),
		TelegramWebAppURL:      os.Getenv("TELEGRAM_WEBAPP_URL"),
		ObjectStorageProvider:  envOrDefault("OBJECT_STORAGE_PROVIDER", "unavailable"),
		ObjectStorageEndpoint:  os.Getenv("OBJECT_STORAGE_ENDPOINT"),
		ObjectStorageRegion:    os.Getenv("OBJECT_STORAGE_REGION"),
		ObjectStorageBucket:    os.Getenv("OBJECT_STORAGE_BUCKET"),
		ObjectStorageAccessKey: os.Getenv("OBJECT_STORAGE_ACCESS_KEY"),
		ObjectStorageSecretKey: os.Getenv("OBJECT_STORAGE_SECRET_KEY"),
		ObjectStoragePublicURL: os.Getenv("OBJECT_STORAGE_PUBLIC_URL"),
		MaxImageUploadBytes:    10 << 20,
		MaxVideoUploadBytes:    500 << 20,
	}

	if cfg.DatabaseURL == "" || cfg.RedisURL == "" || cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("DATABASE_URL, REDIS_URL, and JWT_SECRET must be set")
	}
	if len(cfg.CORSOrigins) == 0 {
		return Config{}, fmt.Errorf("CORS_ORIGINS must contain at least one origin")
	}
	return cfg, nil
}

func NewLogger(level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if origin := strings.TrimSpace(part); origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}
