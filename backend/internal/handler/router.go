package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type HealthDependencies struct {
	Postgres func(context.Context) error
	Redis    func(context.Context) error
}

type Dependencies struct {
	Auth        AuthDependencies
	Lessons     LessonStore
	Exercises   ExerciseStore
	Workouts    WorkoutStore
	Progress    ProgressStore
	Health      HealthDependencies
	Logger      *slog.Logger
	CORSOrigins []string
}

func NewRouter(dependencies Dependencies) http.Handler {
	router := chi.NewRouter()
	router.Use(chimiddleware.Recoverer)
	router.Use(cors(dependencies.CORSOrigins))
	router.Use(securityHeaders)
	router.Get("/healthz", healthHandler(dependencies.Health))
	router.Get("/openapi.yaml", openAPIHandler)
	router.Route("/api/v1", func(api chi.Router) {
		api.Post("/auth/telegram", telegramAuthHandler(dependencies.Auth))
		api.With(func(next http.Handler) http.Handler { return middleware.RequireAuth(dependencies.Auth.JWTSecret, next) }).Group(func(protected chi.Router) {
			protected.Get("/me", meHandler(dependencies.Auth))
			protected.Get("/lessons", listLessons(dependencies.Lessons))
			protected.Get("/lessons/{id}", getLesson(dependencies.Lessons))
			protected.Post("/lessons/{id}/complete", completeLesson(dependencies.Lessons))
			protected.Get("/exercises", listExercises(dependencies.Exercises))
			protected.Get("/exercises/{id}", getExercise(dependencies.Exercises))
			protected.Get("/workouts/today", today(dependencies.Workouts))
			protected.Get("/workouts/{id}", getWorkout(dependencies.Workouts))
			protected.Post("/workouts/{id}/start", start(dependencies.Workouts))
			protected.Post("/workout-sessions/{id}/sets", set(dependencies.Workouts))
			protected.Post("/workout-sessions/{id}/complete", complete(dependencies.Workouts))
			protected.Get("/progress", summary(dependencies.Progress))
			protected.Get("/stats", stats(dependencies.Progress))
			protected.Get("/history", history(dependencies.Progress))
			protected.Get("/achievements", achievements(dependencies.Progress))
		})
	})
	return router
}

func healthHandler(dependencies HealthDependencies) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if err := dependencies.Postgres(request.Context()); err != nil {
			writeError(writer, http.StatusServiceUnavailable, "POSTGRES_UNAVAILABLE", "Database is unavailable")
			return
		}
		if err := dependencies.Redis(request.Context()); err != nil {
			writeError(writer, http.StatusServiceUnavailable, "REDIS_UNAVAILABLE", "Cache is unavailable")
			return
		}
		writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func openAPIHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	_, _ = writer.Write([]byte(openAPI))
}

func writeError(writer http.ResponseWriter, status int, code, message string) {
	writeJSON(writer, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

func writeJSON(writer http.ResponseWriter, status int, payload any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(payload)
}
