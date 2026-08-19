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
	Programs    ProgramStore
	Admin       AdminStore
	Workouts    WorkoutStore
	Progress    ProgressStore
	Skills      SkillsStore
	Calendar    CalendarStore
	Coach       CoachStore
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
			protected.Get("/programs", listPrograms(dependencies.Programs))
			protected.Get("/programs/{id}", getProgram(dependencies.Programs))
			protected.Get("/workouts", listWorkouts(dependencies.Workouts))
			protected.Get("/workouts/today", today(dependencies.Workouts))
			protected.Get("/workouts/{id}", getWorkout(dependencies.Workouts))
			protected.Post("/workouts/{id}/start", start(dependencies.Workouts))
			protected.Get("/workout-sessions/{id}", getWorkoutSession(dependencies.Workouts))
			protected.Post("/workout-sessions/{id}/sets", set(dependencies.Workouts))
			protected.Post("/workout-sessions/{id}/complete", complete(dependencies.Workouts))
			protected.Get("/progress", summary(dependencies.Progress))
			protected.Get("/stats", stats(dependencies.Progress))
			protected.Get("/history", history(dependencies.Progress))
			protected.Get("/achievements", achievements(dependencies.Progress))
			protected.Get("/skills", listSkills(dependencies.Skills))
			protected.Get("/skills/map", skillMap(dependencies.Skills))
			protected.Get("/skills/{id}", getSkill(dependencies.Skills))
			protected.Post("/skills/{id}/levels/{level}/complete", completeSkillLevel(dependencies.Skills))
			protected.Post("/skills/{id}/master", masterSkill(dependencies.Skills))
			protected.Get("/calendar", calendarRange(dependencies.Calendar))
			protected.Get("/calendar/today", calendarToday(dependencies.Calendar))
			protected.Get("/training-schedules", schedules(dependencies.Calendar))
			protected.Post("/training-schedules", schedules(dependencies.Calendar))
			protected.Put("/training-schedules/{id}", scheduleByID(dependencies.Calendar))
			protected.Delete("/training-schedules/{id}", scheduleByID(dependencies.Calendar))
			protected.Post("/planned-workouts", planned(dependencies.Calendar))
			protected.Get("/planned-workouts/{id}", plannedByID(dependencies.Calendar))
			protected.Put("/planned-workouts/{id}", plannedByID(dependencies.Calendar))
			protected.Delete("/planned-workouts/{id}", plannedByID(dependencies.Calendar))
			protected.Post("/planned-workouts/{id}/skip", skipPlanned(dependencies.Calendar))
			protected.With(func(next http.Handler) http.Handler { return requireCoach(dependencies.Coach, next) }).Route("/coach", func(coach chi.Router) {
				coach.Get("/me", coachMe(dependencies.Coach))
				coach.Get("/dashboard", coachDashboard(dependencies.Coach))
				coach.Get("/analytics", coachAnalytics(dependencies.Coach))
				coach.Get("/media", coachMedia(dependencies.Coach))
				coach.Get("/options", func(w http.ResponseWriter, r *http.Request) {
					u, role := coachIdentity(r)
					x, e := dependencies.Coach.Options(r.Context(), u, role)
					if e != nil {
						coachErr(w, e)
						return
					}
					writeJSON(w, 200, x)
				})
				coach.Post("/media", coachMedia(dependencies.Coach))
				coach.Post("/media/upload", coachUploadUnavailable)
				coach.Delete("/media/{id}", coachMediaDelete(dependencies.Coach))
				coach.Get("/{kind}", coachContent(dependencies.Coach))
				coach.Post("/{kind}", coachContent(dependencies.Coach))
				coach.Put("/{kind}/{id}", coachContentID(dependencies.Coach))
				coach.Get("/{kind}/{id}", coachContentID(dependencies.Coach))
				coach.Post("/{kind}/{id}/{action}", coachAction(dependencies.Coach))
			})
			protected.Group(func(adminRoutes chi.Router) {
				adminRoutes.Use(func(next http.Handler) http.Handler { return requireAdmin(dependencies.Admin, next) })
				adminRoutes.Post("/admin/lessons", createLesson(dependencies.Admin))
				adminRoutes.Put("/admin/lessons/{id}", updateLesson(dependencies.Admin))
				adminRoutes.Delete("/admin/lessons/{id}", deleteLesson(dependencies.Admin))
				adminRoutes.Post("/admin/lessons/{id}/publish", publishLesson(dependencies.Admin))
				adminRoutes.Post("/admin/exercises", createExercise(dependencies.Admin))
				adminRoutes.Put("/admin/exercises/{id}", updateExercise(dependencies.Admin))
				adminRoutes.Delete("/admin/exercises/{id}", deleteExercise(dependencies.Admin))
				adminRoutes.Post("/admin/programs", programInput(dependencies.Admin, true))
				adminRoutes.Put("/admin/programs/{id}", programInput(dependencies.Admin, false))
				adminRoutes.Delete("/admin/programs/{id}", deleteProgram(dependencies.Admin))
				adminRoutes.Post("/admin/programs/{id}/publish", publishProgram(dependencies.Admin))
				adminRoutes.Post("/admin/workouts", workoutInput(dependencies.Admin, true))
				adminRoutes.Put("/admin/workouts/{id}", workoutInput(dependencies.Admin, false))
				adminRoutes.Delete("/admin/workouts/{id}", deleteWorkout(dependencies.Admin))
				adminRoutes.Post("/admin/workouts/{id}/exercises", upsertWorkoutExercise(dependencies.Admin))
				adminRoutes.Put("/admin/workouts/{id}/exercises/{exerciseID}", upsertWorkoutExercise(dependencies.Admin))
				adminRoutes.Delete("/admin/workouts/{id}/exercises/{exerciseID}", deleteWorkoutExercise(dependencies.Admin))
			})
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
