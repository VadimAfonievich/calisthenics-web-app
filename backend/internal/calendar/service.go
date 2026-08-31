package calendar

import (
	"context"
	"errors"
	"fmt"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound  = errors.New("calendar item not found")
	ErrForbidden = errors.New("calendar item forbidden")
	ErrInvalid   = errors.New("invalid calendar input")
)

type ScheduleInput struct {
	WorkoutID     string  `json:"workout_id"`
	Weekdays      []int16 `json:"weekdays"`
	PreferredTime *string `json:"preferred_time"`
	StartDate     string  `json:"start_date"`
	EndDate       *string `json:"end_date"`
	Timezone      *string `json:"timezone"`
	Active        *bool   `json:"active,omitempty"`
}
type Schedule struct {
	ID            string  `json:"id"`
	WorkoutID     string  `json:"workout_id"`
	WorkoutTitle  string  `json:"workout_title"`
	Weekdays      []int16 `json:"weekdays"`
	PreferredTime *string `json:"preferred_time"`
	Timezone      string  `json:"timezone"`
	StartDate     string  `json:"start_date"`
	EndDate       *string `json:"end_date"`
	Active        bool    `json:"active"`
}
type PlannedInput struct {
	WorkoutID        string  `json:"workout_id"`
	ScheduledDate    string  `json:"scheduled_date"`
	ScheduledTime    *string `json:"scheduled_time"`
	SourceScheduleID *string `json:"source_schedule_id,omitempty"`
}
type PlannedWorkout struct {
	ID               string  `json:"id"`
	WorkoutID        string  `json:"workout_id"`
	WorkoutTitle     string  `json:"workout_title"`
	ScheduledDate    string  `json:"scheduled_date"`
	ScheduledTime    *string `json:"scheduled_time"`
	Timezone         string  `json:"timezone"`
	SourceScheduleID *string `json:"source_schedule_id"`
	Status           string  `json:"status"`
}
type Occurrence struct {
	Date               string     `json:"date"`
	Time               *string    `json:"time"`
	WorkoutID          string     `json:"workout_id"`
	WorkoutTitle       string     `json:"workout_title"`
	ScheduleID         *string    `json:"schedule_id"`
	PlannedWorkoutID   *string    `json:"planned_workout_id"`
	Status             string     `json:"status"`
	CompletedSessionID *string    `json:"completed_session_id"`
	Category           string     `json:"category"`
	Difficulty         string     `json:"difficulty"`
	EstimatedMinutes   int32      `json:"estimated_minutes"`
	DurationSeconds    *int32     `json:"duration_seconds,omitempty"`
	XPEarned           *int32     `json:"xp_earned,omitempty"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
}
type Service struct{ pool *pgxpool.Pool }

func NewService(p *pgxpool.Pool) *Service { return &Service{p} }

func parseInput(in ScheduleInput, fallbackTZ string) (string, error) {
	tz := fallbackTZ
	if in.Timezone != nil {
		tz = *in.Timezone
	}
	if _, e := time.LoadLocation(tz); e != nil {
		return "", ErrInvalid
	}
	start, e := time.Parse("2006-01-02", in.StartDate)
	if e != nil {
		return "", ErrInvalid
	}
	if in.EndDate != nil {
		end, e := time.Parse("2006-01-02", *in.EndDate)
		if e != nil || end.Before(start) {
			return "", ErrInvalid
		}
	}
	if len(in.Weekdays) == 0 {
		return "", ErrInvalid
	}
	seen := map[int16]bool{}
	for _, d := range in.Weekdays {
		if d < 1 || d > 7 || seen[d] {
			return "", ErrInvalid
		}
		seen[d] = true
	}
	if in.PreferredTime != nil {
		if _, e = time.Parse("15:04", *in.PreferredTime); e != nil {
			return "", ErrInvalid
		}
	}
	return tz, nil
}
func (s *Service) profileTZ(ctx context.Context, user string) (string, error) {
	var z string
	e := s.pool.QueryRow(ctx, `SELECT timezone FROM profiles WHERE user_id=$1::uuid`, user).Scan(&z)
	return z, e
}
func (s *Service) CreateSchedule(ctx context.Context, user string, in ScheduleInput) (Schedule, error) {
	tenant, ok := middleware.TenantID(ctx)
	if !ok {
		return Schedule{}, ErrNotFound
	}
	fallback, e := s.profileTZ(ctx, user)
	if e != nil {
		return Schedule{}, e
	}
	tz, e := parseInput(in, fallback)
	if e != nil {
		return Schedule{}, e
	}
	tx, e := s.pool.Begin(ctx)
	if e != nil {
		return Schedule{}, e
	}
	defer tx.Rollback(ctx)
	var out Schedule
	active := true
	if in.Active != nil {
		active = *in.Active
	}
	e = tx.QueryRow(ctx, `INSERT INTO user_training_schedules(user_id,tenant_id,workout_id,preferred_time,timezone,active,start_date,end_date) SELECT $1::uuid,$8::uuid,w.id,$3::time,$4,$5,$6::date,$7::date FROM workouts w WHERE w.id=$2::uuid AND w.tenant_id=$8::uuid AND w.status='published' RETURNING id::text,workout_id::text,to_char(preferred_time,'HH24:MI'),timezone,start_date::text,end_date::text,active`, user, in.WorkoutID, in.PreferredTime, tz, active, in.StartDate, in.EndDate, tenant).Scan(&out.ID, &out.WorkoutID, &out.PreferredTime, &out.Timezone, &out.StartDate, &out.EndDate, &out.Active)
	if errors.Is(e, pgx.ErrNoRows) {
		return out, ErrNotFound
	}
	if e != nil {
		return out, e
	}
	for _, d := range in.Weekdays {
		if _, e = tx.Exec(ctx, `INSERT INTO user_training_schedule_days(schedule_id,weekday) VALUES($1::uuid,$2)`, out.ID, d); e != nil {
			return out, e
		}
	}
	out.Weekdays = append([]int16(nil), in.Weekdays...)
	sort.Slice(out.Weekdays, func(i, j int) bool { return out.Weekdays[i] < out.Weekdays[j] })
	if e = tx.Commit(ctx); e != nil {
		return out, e
	}
	return out, nil
}
func (s *Service) ListSchedules(ctx context.Context, user string) ([]Schedule, error) {
	tenant, ok := middleware.TenantID(ctx)
	if !ok {
		return []Schedule{}, nil
	}
	rows, e := s.pool.Query(ctx, `SELECT s.id::text,s.workout_id::text,w.title,to_char(s.preferred_time,'HH24:MI'),s.timezone,s.start_date::text,s.end_date::text,s.active,COALESCE(array_agg(d.weekday ORDER BY d.weekday) FILTER(WHERE d.weekday IS NOT NULL),'{}') FROM user_training_schedules s JOIN workouts w ON w.id=s.workout_id LEFT JOIN user_training_schedule_days d ON d.schedule_id=s.id WHERE s.user_id=$1::uuid AND s.tenant_id=$2::uuid GROUP BY s.id,w.title ORDER BY s.active DESC,s.created_at`, user, tenant)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Schedule{}
	for rows.Next() {
		var x Schedule
		if e = rows.Scan(&x.ID, &x.WorkoutID, &x.WorkoutTitle, &x.PreferredTime, &x.Timezone, &x.StartDate, &x.EndDate, &x.Active, &x.Weekdays); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (s *Service) UpdateSchedule(ctx context.Context, user, id string, in ScheduleInput) (Schedule, error) {
	tenant, ok := middleware.TenantID(ctx)
	if !ok {
		return Schedule{}, ErrNotFound
	}
	fallback, e := s.profileTZ(ctx, user)
	if e != nil {
		return Schedule{}, e
	}
	tz, e := parseInput(in, fallback)
	if e != nil {
		return Schedule{}, e
	}
	tx, e := s.pool.Begin(ctx)
	if e != nil {
		return Schedule{}, e
	}
	defer tx.Rollback(ctx)
	active := true
	if in.Active != nil {
		active = *in.Active
	}
	var out Schedule
	e = tx.QueryRow(ctx, `UPDATE user_training_schedules s SET workout_id=$3::uuid,preferred_time=$4::time,timezone=$5,active=$6,start_date=$7::date,end_date=$8::date,updated_at=NOW() WHERE id=$1::uuid AND user_id=$2::uuid AND tenant_id=$9::uuid AND EXISTS(SELECT 1 FROM workouts w WHERE w.id=$3::uuid AND w.tenant_id=$9::uuid AND w.status='published') RETURNING id::text,workout_id::text,to_char(preferred_time,'HH24:MI'),timezone,start_date::text,end_date::text,active`, id, user, in.WorkoutID, in.PreferredTime, tz, active, in.StartDate, in.EndDate, tenant).Scan(&out.ID, &out.WorkoutID, &out.PreferredTime, &out.Timezone, &out.StartDate, &out.EndDate, &out.Active)
	if errors.Is(e, pgx.ErrNoRows) {
		return out, ErrNotFound
	}
	if e != nil {
		return out, e
	}
	_, e = tx.Exec(ctx, `DELETE FROM user_training_schedule_days WHERE schedule_id=$1::uuid`, id)
	if e != nil {
		return out, e
	}
	for _, d := range in.Weekdays {
		if _, e = tx.Exec(ctx, `INSERT INTO user_training_schedule_days(schedule_id,weekday) VALUES($1::uuid,$2)`, id, d); e != nil {
			return out, e
		}
	}
	out.Weekdays = in.Weekdays
	if e = tx.Commit(ctx); e != nil {
		return out, e
	}
	return out, nil
}
func (s *Service) DeleteSchedule(ctx context.Context, user, id string) error {
	tenant, ok := middleware.TenantID(ctx)
	if !ok {
		return ErrNotFound
	}
	tag, e := s.pool.Exec(ctx, `UPDATE user_training_schedules SET active=false,updated_at=NOW() WHERE id=$1::uuid AND user_id=$2::uuid AND tenant_id=$3::uuid`, id, user, tenant)
	if e != nil {
		return e
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func validatePlanned(in PlannedInput) error {
	if _, e := time.Parse("2006-01-02", in.ScheduledDate); e != nil {
		return ErrInvalid
	}
	if in.ScheduledTime != nil {
		if _, e := time.Parse("15:04", *in.ScheduledTime); e != nil {
			return ErrInvalid
		}
	}
	return nil
}
func (s *Service) CreatePlanned(ctx context.Context, user string, in PlannedInput) (PlannedWorkout, error) {
	tenant, ok := middleware.TenantID(ctx)
	if !ok {
		return PlannedWorkout{}, ErrNotFound
	}
	if e := validatePlanned(in); e != nil {
		return PlannedWorkout{}, e
	}
	tz, e := s.profileTZ(ctx, user)
	if e != nil {
		return PlannedWorkout{}, e
	}
	var x PlannedWorkout
	e = s.pool.QueryRow(ctx, `INSERT INTO user_planned_workouts(user_id,tenant_id,workout_id,scheduled_date,scheduled_time,timezone,source_schedule_id) SELECT $1::uuid,$7::uuid,w.id,$3::date,$4::time,$5,$6::uuid FROM workouts w WHERE w.id=$2::uuid AND w.tenant_id=$7::uuid AND ($6::uuid IS NULL OR EXISTS(SELECT 1 FROM user_training_schedules s WHERE s.id=$6::uuid AND s.user_id=$1::uuid AND s.tenant_id=$7::uuid AND s.workout_id=w.id)) ON CONFLICT (user_id,source_schedule_id,scheduled_date) WHERE source_schedule_id IS NOT NULL DO UPDATE SET scheduled_time=EXCLUDED.scheduled_time,tenant_id=EXCLUDED.tenant_id RETURNING id::text,workout_id::text,scheduled_date::text,to_char(scheduled_time,'HH24:MI'),timezone,source_schedule_id::text,status`, user, in.WorkoutID, in.ScheduledDate, in.ScheduledTime, tz, in.SourceScheduleID, tenant).Scan(&x.ID, &x.WorkoutID, &x.ScheduledDate, &x.ScheduledTime, &x.Timezone, &x.SourceScheduleID, &x.Status)
	if e != nil {
		return x, e
	}
	return s.GetPlanned(ctx, user, x.ID)
}
func (s *Service) GetPlanned(ctx context.Context, user, id string) (PlannedWorkout, error) {
	tenant, ok := middleware.TenantID(ctx)
	if !ok {
		return PlannedWorkout{}, ErrNotFound
	}
	var x PlannedWorkout
	e := s.pool.QueryRow(ctx, `SELECT p.id::text,p.workout_id::text,w.title,p.scheduled_date::text,to_char(p.scheduled_time,'HH24:MI'),p.timezone,p.source_schedule_id::text,p.status FROM user_planned_workouts p JOIN workouts w ON w.id=p.workout_id WHERE p.id=$1::uuid AND p.user_id=$2::uuid AND p.tenant_id=$3::uuid`, id, user, tenant).Scan(&x.ID, &x.WorkoutID, &x.WorkoutTitle, &x.ScheduledDate, &x.ScheduledTime, &x.Timezone, &x.SourceScheduleID, &x.Status)
	if errors.Is(e, pgx.ErrNoRows) {
		return x, ErrNotFound
	}
	return x, e
}
func (s *Service) UpdatePlanned(ctx context.Context, user, id string, in PlannedInput) (PlannedWorkout, error) {
	tenant, ok := middleware.TenantID(ctx)
	if !ok {
		return PlannedWorkout{}, ErrNotFound
	}
	if e := validatePlanned(in); e != nil {
		return PlannedWorkout{}, e
	}
	tag, e := s.pool.Exec(ctx, `UPDATE user_planned_workouts SET workout_id=$3::uuid,scheduled_date=$4::date,scheduled_time=$5::time,updated_at=NOW() WHERE id=$1::uuid AND user_id=$2::uuid AND tenant_id=$6::uuid AND status='scheduled' AND EXISTS(SELECT 1 FROM workouts w WHERE w.id=$3::uuid AND w.tenant_id=$6::uuid AND w.status='published')`, id, user, in.WorkoutID, in.ScheduledDate, in.ScheduledTime, tenant)
	if e != nil {
		return PlannedWorkout{}, e
	}
	if tag.RowsAffected() == 0 {
		return PlannedWorkout{}, ErrNotFound
	}
	return s.GetPlanned(ctx, user, id)
}
func (s *Service) DeletePlanned(ctx context.Context, user, id string) error {
	return s.setStatus(ctx, user, id, "cancelled")
}
func (s *Service) SkipPlanned(ctx context.Context, user, id string) error {
	return s.setStatus(ctx, user, id, "skipped")
}
func (s *Service) setStatus(ctx context.Context, user, id, status string) error {
	tenant, ok := middleware.TenantID(ctx)
	if !ok {
		return ErrNotFound
	}
	tag, e := s.pool.Exec(ctx, `UPDATE user_planned_workouts SET status=$3,updated_at=NOW() WHERE id=$1::uuid AND user_id=$2::uuid AND tenant_id=$4::uuid AND status='scheduled'`, id, user, status, tenant)
	if e != nil {
		return e
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) Calendar(ctx context.Context, user, from, to string) ([]Occurrence, error) {
	tenant, ok := middleware.TenantID(ctx)
	if !ok {
		return []Occurrence{}, nil
	}
	f, fe := time.Parse("2006-01-02", from)
	t, te := time.Parse("2006-01-02", to)
	if fe != nil || te != nil || t.Before(f) || t.Sub(f) > 366*24*time.Hour {
		return nil, ErrInvalid
	}
	tz, e := s.profileTZ(ctx, user)
	if e != nil {
		return nil, e
	}
	loc, e := time.LoadLocation(tz)
	if e != nil {
		return nil, e
	}
	today := time.Now().In(loc).Format("2006-01-02")
	rows, e := s.pool.Query(ctx, `SELECT s.id::text,s.workout_id::text,w.title,to_char(s.preferred_time,'HH24:MI'),s.start_date::text,s.end_date::text,w.category,w.difficulty,w.estimated_minutes,array_agg(d.weekday) FROM user_training_schedules s JOIN user_training_schedule_days d ON d.schedule_id=s.id JOIN workouts w ON w.id=s.workout_id WHERE s.user_id=$1::uuid AND s.tenant_id=$4::uuid AND s.active AND s.start_date<=$3::date AND (s.end_date IS NULL OR s.end_date>=$2::date) GROUP BY s.id,w.id`, user, from, to, tenant)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Occurrence{}
	for rows.Next() {
		var sid, wid, title, start string
		var tm, end *string
		var cat, diff string
		var mins int32
		var days []int16
		if e = rows.Scan(&sid, &wid, &title, &tm, &start, &end, &cat, &diff, &mins, &days); e != nil {
			return nil, e
		}
		set := map[int]bool{}
		for _, d := range days {
			set[int(d)] = true
		}
		for d := f; !d.After(t); d = d.AddDate(0, 0, 1) {
			ds := d.Format("2006-01-02")
			weekday := ((int(d.Weekday()) + 6) % 7) + 1
			if ds < start || (end != nil && ds > *end) || !set[weekday] {
				continue
			}
			id := sid
			status := "scheduled"
			if ds < today {
				status = "missed"
			}
			out = append(out, Occurrence{Date: ds, Time: tm, WorkoutID: wid, WorkoutTitle: title, ScheduleID: &id, Status: status, Category: cat, Difficulty: diff, EstimatedMinutes: mins})
		}
	}
	prows, e := s.pool.Query(ctx, `SELECT pw.id::text,pw.workout_id::text,w.title,pw.scheduled_date::text,to_char(pw.scheduled_time,'HH24:MI'),pw.source_schedule_id::text,pw.status,w.category,w.difficulty,w.estimated_minutes,ws.id::text,ws.duration_seconds,ws.xp_earned,ws.completed_at FROM user_planned_workouts pw JOIN workouts w ON w.id=pw.workout_id LEFT JOIN workout_sessions ws ON ws.planned_workout_id=pw.id AND ws.tenant_id=$4::uuid AND ws.status='completed' WHERE pw.user_id=$1::uuid AND pw.tenant_id=$4::uuid AND pw.scheduled_date BETWEEN $2::date AND $3::date`, user, from, to, tenant)
	if e != nil {
		return nil, e
	}
	defer prows.Close()
	overrides := map[string]Occurrence{}
	for prows.Next() {
		var x Occurrence
		var persisted string
		if e = prows.Scan(&persisted, &x.WorkoutID, &x.WorkoutTitle, &x.Date, &x.Time, &x.ScheduleID, &x.Status, &x.Category, &x.Difficulty, &x.EstimatedMinutes, &x.CompletedSessionID, &x.DurationSeconds, &x.XPEarned, &x.CompletedAt); e != nil {
			return nil, e
		}
		x.PlannedWorkoutID = &persisted
		if x.Status == "scheduled" && x.Date < today {
			x.Status = "missed"
		}
		if x.ScheduleID != nil {
			overrides[*x.ScheduleID+"/"+x.Date] = x
		} else {
			out = append(out, x)
		}
	}
	for i := range out {
		if out[i].ScheduleID != nil {
			if x, ok := overrides[*out[i].ScheduleID+"/"+out[i].Date]; ok {
				out[i] = x
				delete(overrides, *out[i].ScheduleID+"/"+out[i].Date)
			}
		}
	}
	for _, x := range overrides {
		out = append(out, x)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Date == out[j].Date {
			return fmt.Sprint(out[i].Time) < fmt.Sprint(out[j].Time)
		}
		return out[i].Date < out[j].Date
	})
	return out, nil
}
func (s *Service) Today(ctx context.Context, user string) ([]Occurrence, error) {
	tz, e := s.profileTZ(ctx, user)
	if e != nil {
		return nil, e
	}
	loc, e := time.LoadLocation(tz)
	if e != nil {
		return nil, e
	}
	d := time.Now().In(loc).Format("2006-01-02")
	return s.Calendar(ctx, user, d, d)
}
