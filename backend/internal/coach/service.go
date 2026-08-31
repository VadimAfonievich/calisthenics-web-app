package coach

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrForbidden       = errors.New("coach content forbidden")
	ErrNotFound        = errors.New("coach content not found")
	ErrInvalid         = errors.New("coach content invalid")
	ErrInUse           = errors.New("media in use")
	ErrWorkoutDayInUse = errors.New("workout day is already in use")
	ErrDependency      = errors.New("published dependency prevents lifecycle change")
)

type ValidationError struct{ Message string }

func (e *ValidationError) Error() string { return e.Message }
func (e *ValidationError) Unwrap() error { return ErrInvalid }

func invalid(message string) error { return &ValidationError{Message: message} }

type Role string

func (r Role) CanManageAll() bool { return r == "admin" || r == "super_admin" }

type Dashboard struct {
	Lessons           int `json:"lessons"`
	LessonsPublished  int `json:"lessons_published"`
	Exercises         int `json:"exercises"`
	Workouts          int `json:"workouts"`
	WorkoutsPublished int `json:"workouts_published"`
	Programs          int `json:"programs"`
	Skills            int `json:"skills"`
	Media             int `json:"media"`
}
type Analytics struct {
	TotalUsers             int      `json:"total_users"`
	ActiveUsers7D          int      `json:"active_users_7d"`
	ActiveUsers30D         int      `json:"active_users_30d"`
	TotalWorkoutsCompleted int      `json:"total_workouts_completed"`
	Workouts7D             int      `json:"workouts_7d"`
	Workouts30D            int      `json:"workouts_30d"`
	TotalLessonsCompleted  int      `json:"total_lessons_completed"`
	PopularWorkouts        []Metric `json:"popular_workouts"`
	SkillProgress          []Metric `json:"skill_progress"`
	TopAchievements        []Metric `json:"top_achievements"`
}
type Metric struct {
	Name      string  `json:"name"`
	Value     int     `json:"value"`
	Secondary float64 `json:"secondary,omitempty"`
}
type Item struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug,omitempty"`
	Status      string    `json:"status"`
	Difficulty  string    `json:"difficulty"`
	OwnerUserID *string   `json:"owner_user_id"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type Option struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Status         string   `json:"status,omitempty"`
	Difficulty     string   `json:"difficulty,omitempty"`
	Minutes        int      `json:"minutes,omitempty"`
	SortOrder      int      `json:"sort_order,omitempty"`
	OwnerUserID    *string  `json:"owner_user_id,omitempty"`
	ParentID       *string  `json:"parent_id,omitempty"`
	Description    string   `json:"description,omitempty"`
	Instructions   string   `json:"instructions,omitempty"`
	CommonMistakes string   `json:"common_mistakes,omitempty"`
	CoachTips      string   `json:"coach_tips,omitempty"`
	MovementType   string   `json:"movement_type,omitempty"`
	MuscleGroups   []string `json:"muscle_groups,omitempty"`
	Equipment      []string `json:"equipment,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	HasDemo        bool     `json:"has_demo,omitempty"`
	ExerciseCount  int      `json:"exercise_count,omitempty"`
}
type Options struct {
	Categories    []Option `json:"categories"`
	Exercises     []Option `json:"exercises"`
	Programs      []Option `json:"programs"`
	ProgramLevels []Option `json:"program_levels"`
	Workouts      []Option `json:"workouts"`
	Warmups       []Option `json:"warmups"`
	Skills        []Option `json:"skills"`
	Media         []Option `json:"media"`
}
type Block struct {
	Type    string   `json:"type"`
	Text    string   `json:"text,omitempty"`
	MediaID *string  `json:"media_id,omitempty"`
	Items   []string `json:"items,omitempty"`
}
type LessonInput struct {
	CategoryID       string  `json:"category_id"`
	Title            string  `json:"title"`
	Slug             string  `json:"slug"`
	ShortDescription string  `json:"short_description"`
	Content          string  `json:"content"`
	Difficulty       string  `json:"difficulty"`
	DurationMinutes  int     `json:"duration_minutes"`
	CoverMediaID     *string `json:"cover_media_id,omitempty"`
	Blocks           []Block `json:"blocks"`
}
type BuilderInput struct {
	Name                string            `json:"name"`
	Title               string            `json:"title"`
	Slug                string            `json:"slug"`
	Code                string            `json:"code"`
	Description         string            `json:"description"`
	Difficulty          string            `json:"difficulty"`
	Category            string            `json:"category"`
	MapGroup            string            `json:"map_group"`
	DurationWeeks       int               `json:"duration_weeks"`
	EstimatedMinutes    int               `json:"estimated_minutes"`
	DayNumber           int               `json:"day_number"`
	ProgramID           string            `json:"program_id"`
	ProgramLevelID      *string           `json:"program_level_id,omitempty"`
	MovementType        string            `json:"movement_type"`
	Instructions        string            `json:"instructions"`
	CommonMistakes      string            `json:"common_mistakes"`
	CoachTips           string            `json:"coach_tips"`
	MuscleGroups        []string          `json:"muscle_groups"`
	Equipment           []string          `json:"equipment"`
	Tags                []string          `json:"tags"`
	Icon                string            `json:"icon"`
	XPReward            int               `json:"xp_reward"`
	FinalCriterionType  string            `json:"final_criterion_type"`
	FinalCriterionValue int               `json:"final_criterion_value"`
	CoverMediaID        *string           `json:"cover_media_id,omitempty"`
	DemoMediaID         *string           `json:"demo_media_id,omitempty"`
	Exercises           []BuilderExercise `json:"exercises"`
	Levels              []BuilderLevel    `json:"levels"`
	Requirements        []string          `json:"requirements"`
	Workouts            []ProgramWorkout  `json:"workouts"`
	SortOrder           int               `json:"sort_order"`
	WarmupEnabled       *bool             `json:"warmup_enabled,omitempty"`
	WarmupWorkoutID     *string           `json:"warmup_workout_id,omitempty"`
}
type ProgramWorkout struct {
	WorkoutID string `json:"workout_id"`
	SortOrder int    `json:"sort_order"`
}
type BuilderExercise struct {
	ExerciseID            string  `json:"exercise_id"`
	Sets                  int     `json:"sets"`
	TargetReps            *int    `json:"target_reps,omitempty"`
	TargetDurationSeconds *int    `json:"target_duration_seconds,omitempty"`
	RestSeconds           int     `json:"rest_seconds"`
	Notes                 *string `json:"notes,omitempty"`
	SortOrder             int     `json:"sort_order"`
}
type BuilderLevel struct {
	LevelNumber     int              `json:"level_number"`
	Title           string           `json:"title"`
	Description     string           `json:"description"`
	Difficulty      string           `json:"difficulty"`
	UnlockRuleType  string           `json:"unlock_rule_type"`
	UnlockRuleValue int              `json:"unlock_rule_value"`
	CriterionType   string           `json:"criterion_type"`
	CriterionValue  int              `json:"criterion_value"`
	ProgramLevelID  *string          `json:"program_level_id,omitempty"`
	SortOrder       int              `json:"sort_order"`
	Workouts        []ProgramWorkout `json:"workouts,omitempty"`
}
type MediaInput struct {
	Type             string  `json:"type"`
	URL              string  `json:"url"`
	ThumbnailURL     *string `json:"thumbnail_url,omitempty"`
	OriginalFilename string  `json:"original_filename"`
	MIMEType         string  `json:"mime_type"`
	SizeBytes        int64   `json:"size_bytes"`
	Width            *int    `json:"width,omitempty"`
	Height           *int    `json:"height,omitempty"`
	DurationSeconds  *int    `json:"duration_seconds,omitempty"`
}
type Media struct {
	ID               string    `json:"id"`
	OwnerUserID      *string   `json:"owner_user_id"`
	Type             string    `json:"type"`
	Status           string    `json:"status"`
	StorageProvider  string    `json:"storage_provider"`
	StorageKey       string    `json:"storage_key"`
	URL              string    `json:"url"`
	ThumbnailURL     *string   `json:"thumbnail_url"`
	OriginalFilename string    `json:"original_filename"`
	MIMEType         string    `json:"mime_type"`
	SizeBytes        int64     `json:"size_bytes"`
	CreatedAt        time.Time `json:"created_at"`
	References       int       `json:"references"`
}
type Service struct{ pool *pgxpool.Pool }

func NewService(p *pgxpool.Pool) *Service { return &Service{p} }
func tenantID(ctx context.Context) (string, error) {
	id, ok := middleware.TenantID(ctx)
	if !ok || id == "" {
		return "", ErrForbidden
	}
	role, _ := middleware.TenantRole(ctx)
	if role != "coach" {
		return "", ErrForbidden
	}
	return id, nil
}
func (s *Service) Role(ctx context.Context, user string) (Role, error) {
	var r Role
	e := s.pool.QueryRow(ctx, `SELECT role FROM admin_users WHERE user_id=$1::uuid`, user).Scan(&r)
	if errors.Is(e, pgx.ErrNoRows) {
		var coach bool
		if membershipErr := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tenant_memberships m JOIN tenants t ON t.id=m.tenant_id WHERE m.user_id=$1::uuid AND m.role='coach' AND m.status='active' AND t.status='active')`, user).Scan(&coach); membershipErr == nil && coach {
			return "coach", nil
		}
		return "user", nil
	}
	return r, e
}
func scope(role Role, user string) (string, []any) {
	if role.CanManageAll() {
		return "", nil
	}
	return " AND owner_user_id=$1::uuid", []any{user}
}
func (s *Service) Dashboard(ctx context.Context, user string, role Role) (Dashboard, error) {
	tenant, e := tenantID(ctx)
	if e != nil {
		return Dashboard{}, e
	}
	var d Dashboard
	q := `SELECT (SELECT count(*) FROM lessons WHERE tenant_id=$1::uuid),(SELECT count(*) FROM lessons WHERE tenant_id=$1::uuid AND status='published'),(SELECT count(*) FROM exercises WHERE tenant_id=$1::uuid OR (tenant_id IS NULL AND standard_key IS NOT NULL)),(SELECT count(*) FROM workouts WHERE tenant_id=$1::uuid),(SELECT count(*) FROM workouts WHERE tenant_id=$1::uuid AND status='published'),(SELECT count(*) FROM programs WHERE tenant_id=$1::uuid),(SELECT count(*) FROM skills WHERE tenant_id=$1::uuid),(SELECT count(*) FROM media_assets WHERE tenant_id=$1::uuid)`
	e = s.pool.QueryRow(ctx, q, tenant).Scan(&d.Lessons, &d.LessonsPublished, &d.Exercises, &d.Workouts, &d.WorkoutsPublished, &d.Programs, &d.Skills, &d.Media)
	return d, e
}
func (s *Service) Analytics(ctx context.Context, user string) (Analytics, error) {
	a := Analytics{
		PopularWorkouts: []Metric{},
		SkillProgress:   []Metric{},
		TopAchievements: []Metric{},
	}
	tenant, tenantErr := tenantID(ctx)
	if tenantErr != nil {
		return a, tenantErr
	}
	e := s.pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM tenant_memberships WHERE tenant_id=$1::uuid AND role='student' AND status='active'),(SELECT count(DISTINCT user_id) FROM workout_sessions WHERE tenant_id=$1::uuid AND started_at>=NOW()-INTERVAL '7 days'),(SELECT count(DISTINCT user_id) FROM workout_sessions WHERE tenant_id=$1::uuid AND started_at>=NOW()-INTERVAL '30 days'),(SELECT count(*) FROM workout_sessions WHERE tenant_id=$1::uuid AND status='completed'),(SELECT count(*) FROM workout_sessions WHERE tenant_id=$1::uuid AND status='completed' AND completed_at>=NOW()-INTERVAL '7 days'),(SELECT count(*) FROM workout_sessions WHERE tenant_id=$1::uuid AND status='completed' AND completed_at>=NOW()-INTERVAL '30 days'),(SELECT count(*) FROM user_lesson_progress WHERE tenant_id=$1::uuid AND completed)`, tenant).Scan(&a.TotalUsers, &a.ActiveUsers7D, &a.ActiveUsers30D, &a.TotalWorkoutsCompleted, &a.Workouts7D, &a.Workouts30D, &a.TotalLessonsCompleted)
	if e != nil {
		return a, e
	}
	rows, _ := s.pool.Query(ctx, `SELECT w.title,count(*)::int,COALESCE(avg(ws.duration_seconds),0)::float8 FROM workout_sessions ws JOIN workouts w ON w.id=ws.workout_id WHERE ws.tenant_id=$1::uuid AND ws.status='completed' GROUP BY w.id ORDER BY count(*) DESC LIMIT 10`, tenant)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var m Metric
			if rows.Scan(&m.Name, &m.Value, &m.Secondary) == nil {
				a.PopularWorkouts = append(a.PopularWorkouts, m)
			}
		}
	}
	rows, _ = s.pool.Query(ctx, `SELECT s.name,count(*)::int,COALESCE(avg(usp.current_level),0)::float8 FROM user_skill_progress usp JOIN skills s ON s.id=usp.skill_id WHERE usp.tenant_id=$1::uuid GROUP BY s.id ORDER BY count(*) DESC LIMIT 10`, tenant)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var m Metric
			if rows.Scan(&m.Name, &m.Value, &m.Secondary) == nil {
				a.SkillProgress = append(a.SkillProgress, m)
			}
		}
	}
	return a, nil
}
func (s *Service) metrics(ctx context.Context, q string) ([]Metric, error) {
	rows, e := s.pool.Query(ctx, q)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Metric{}
	for rows.Next() {
		var x Metric
		if e = rows.Scan(&x.Name, &x.Value, &x.Secondary); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

var tables = map[string]struct{ table, name, slug, listFrom, difficulty string }{
	"lessons":   {"lessons", "title", "slug", "lessons", "difficulty"},
	"exercises": {"exercises", "name", "slug", "exercises", "difficulty"},
	"programs":  {"programs", "name", "slug", "programs", "difficulty"},
	"workouts":  {"workouts", "title", "''", "workouts", "difficulty"},
	"skills":    {"skills", "name", "code", "skills", "difficulty"},
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)
var uuidValue = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func validID(value string) bool          { return uuidValue.MatchString(value) }
func validOptionalID(value *string) bool { return value == nil || validID(*value) }
func validMapGroup(value string) bool {
	return value == "basic" || value == "floor" || value == "bar" || value == "parallel_bars"
}
func validTags(values []string) bool {
	seen := map[string]bool{}
	for _, value := range values {
		if value == "" || slugBase(value) != value || seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}

func slugBase(value string) string {
	r := strings.NewReplacer("а", "a", "б", "b", "в", "v", "г", "g", "д", "d", "е", "e", "ё", "e", "ж", "zh", "з", "z", "и", "i", "й", "y", "к", "k", "л", "l", "м", "m", "н", "n", "о", "o", "п", "p", "р", "r", "с", "s", "т", "t", "у", "u", "ф", "f", "х", "h", "ц", "ts", "ч", "ch", "ш", "sh", "щ", "sch", "ъ", "", "ы", "y", "ь", "", "э", "e", "ю", "yu", "я", "ya")
	s := nonSlug.ReplaceAllString(r.Replace(strings.ToLower(strings.TrimSpace(value))), "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "content"
	}
	return s
}
func (s *Service) uniqueSlug(ctx context.Context, table, value string, id *string) (string, error) {
	base := slugBase(value)
	candidate := base
	var exclude any
	if id != nil {
		exclude = *id
	}
	for n := 1; n < 1000; n++ {
		var exists bool
		q := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE slug=$1 AND ($2::uuid IS NULL OR id<>$2::uuid))", table)
		if e := s.pool.QueryRow(ctx, q, candidate, exclude).Scan(&exists); e != nil {
			return "", e
		}
		if !exists {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, n+1)
	}
	return "", ErrInvalid
}

func (s *Service) Get(ctx context.Context, kind, id, user string, role Role) (map[string]any, error) {
	tenant, e := tenantID(ctx)
	if e != nil {
		return nil, e
	}
	x, ok := tables[kind]
	if !ok {
		return nil, ErrInvalid
	}
	condition := "tenant_id=$2::uuid"
	if kind == "exercises" {
		condition = "tenant_id=$2::uuid OR (tenant_id IS NULL AND owner_user_id IS NULL AND standard_key IS NOT NULL)"
	}
	q := fmt.Sprintf("SELECT to_jsonb(t) FROM %s t WHERE id=$1::uuid AND (%s)", x.table, condition)
	var raw []byte
	if e := s.pool.QueryRow(ctx, q, id, tenant).Scan(&raw); errors.Is(e, pgx.ErrNoRows) {
		return nil, ErrForbidden
	} else if e != nil {
		return nil, e
	}
	var out map[string]any
	if e := json.Unmarshal(raw, &out); e != nil {
		return nil, e
	}
	var extra []byte
	var extraErr error
	switch kind {
	case "workouts":
		extraErr = s.pool.QueryRow(ctx, `SELECT COALESCE(jsonb_agg(to_jsonb(we) ORDER BY sort_order),'[]') FROM workout_exercises we WHERE workout_id=$1::uuid`, id).Scan(&extra)
		out["exercises"] = json.RawMessage(extra)
	case "programs":
		extraErr = s.pool.QueryRow(ctx, `SELECT COALESCE(jsonb_agg(to_jsonb(pl)||jsonb_build_object('workouts',COALESCE((SELECT jsonb_agg(jsonb_build_object('workout_id',w.id,'sort_order',w.sort_order) ORDER BY w.sort_order) FROM workouts w WHERE w.program_level_id=pl.id),'[]'::jsonb)) ORDER BY pl.sort_order),'[]') FROM program_levels pl WHERE program_id=$1::uuid`, id).Scan(&extra)
		out["levels"] = json.RawMessage(extra)
		if extraErr == nil {
			extraErr = s.pool.QueryRow(ctx, `SELECT COALESCE(jsonb_agg(jsonb_build_object('workout_id',w.id,'sort_order',w.sort_order) ORDER BY w.sort_order),'[]') FROM workouts w WHERE w.program_id=$1::uuid AND w.program_level_id IS NOT NULL`, id).Scan(&extra)
			out["workouts"] = json.RawMessage(extra)
		}
	case "skills":
		extraErr = s.pool.QueryRow(ctx, `SELECT COALESCE(jsonb_agg(to_jsonb(sl) ORDER BY sort_order),'[]') FROM skill_levels sl WHERE skill_id=$1::uuid`, id).Scan(&extra)
		out["levels"] = json.RawMessage(extra)
		var req []string
		rows, e := s.pool.Query(ctx, `SELECT required_skill_id::text FROM skill_requirements WHERE skill_id=$1::uuid`, id)
		if e == nil {
			defer rows.Close()
			for rows.Next() {
				var v string
				_ = rows.Scan(&v)
				req = append(req, v)
			}
		}
		out["requirements"] = req
	}
	if extraErr != nil {
		return nil, extraErr
	}
	return out, nil
}
func (s *Service) Options(ctx context.Context, user string, role Role) (Options, error) {
	tenant, e := tenantID(ctx)
	if e != nil {
		return Options{}, e
	}
	var out Options
	sets := []struct {
		q    string
		dst  *[]Option
		rich bool
	}{
		{`SELECT id::text,name FROM lesson_categories ORDER BY sort_order,name`, &out.Categories, false},
		{`SELECT id::text,name,status,difficulty,0,0,owner_user_id::text,NULL::text FROM exercises WHERE status<>'archived' AND (tenant_id=$1::uuid OR (tenant_id IS NULL AND owner_user_id IS NULL AND standard_key IS NOT NULL)) ORDER BY name`, &out.Exercises, true},
		{`SELECT id::text,name,status,difficulty,0,0,owner_user_id::text,NULL::text FROM programs WHERE status<>'archived' AND tenant_id=$1::uuid ORDER BY name`, &out.Programs, true},
		{`SELECT pl.id::text,p.name||' · '||pl.title,p.status,pl.difficulty,0,pl.sort_order,p.owner_user_id::text,p.id::text FROM program_levels pl JOIN programs p ON p.id=pl.program_id WHERE p.status<>'archived' AND p.tenant_id=$1::uuid ORDER BY p.name,pl.sort_order`, &out.ProgramLevels, true},
		{`SELECT w.id::text,w.title,w.status,w.difficulty,w.estimated_minutes,w.sort_order,w.owner_user_id::text,w.program_level_id::text FROM workouts w WHERE w.status<>'archived' AND w.tenant_id=$1::uuid ORDER BY w.title`, &out.Workouts, true},
		{`SELECT w.id::text,w.title,w.status,w.difficulty,w.estimated_minutes,w.sort_order,w.owner_user_id::text,NULL::text FROM workouts w WHERE w.status='published' AND w.category='warmup' AND w.tenant_id=$1::uuid ORDER BY w.is_default_warmup DESC,w.title`, &out.Warmups, true},
		{`SELECT id::text,name,status,difficulty,0,sort_order,owner_user_id::text,NULL::text FROM skills WHERE status<>'archived' AND tenant_id=$1::uuid ORDER BY CASE WHEN sort_order=0 THEN 2147483647 ELSE sort_order END,name,id`, &out.Skills, true},
		{`SELECT id::text,CASE WHEN type='image' THEN 'Фото: ' ELSE 'Видео: ' END||original_filename FROM media_assets WHERE status='ready' AND tenant_id=$1::uuid ORDER BY created_at DESC`, &out.Media, false},
	}
	for index, set := range sets {
		args := []any{tenant}
		if index == 0 {
			args = nil
		}
		rows, e := s.pool.Query(ctx, set.q, args...)
		if e != nil {
			return out, e
		}
		for rows.Next() {
			var x Option
			if set.rich {
				e = rows.Scan(&x.ID, &x.Name, &x.Status, &x.Difficulty, &x.Minutes, &x.SortOrder, &x.OwnerUserID, &x.ParentID)
			} else {
				e = rows.Scan(&x.ID, &x.Name)
			}
			if e != nil {
				rows.Close()
				return out, e
			}
			*set.dst = append(*set.dst, x)
		}
		rows.Close()
	}
	rows, e := s.pool.Query(ctx, `SELECT id::text,name,status,difficulty,owner_user_id::text,description,instructions,common_mistakes,coach_tips,movement_type,muscle_groups,equipment,tags,demo_media_id IS NOT NULL FROM exercises WHERE status<>'archived' AND (tenant_id=$1::uuid OR (tenant_id IS NULL AND owner_user_id IS NULL AND standard_key IS NOT NULL)) ORDER BY name`, tenant)
	if e != nil {
		return out, e
	}
	out.Exercises = nil
	for rows.Next() {
		var x Option
		if e = rows.Scan(&x.ID, &x.Name, &x.Status, &x.Difficulty, &x.OwnerUserID, &x.Description, &x.Instructions, &x.CommonMistakes, &x.CoachTips, &x.MovementType, &x.MuscleGroups, &x.Equipment, &x.Tags, &x.HasDemo); e != nil {
			rows.Close()
			return out, e
		}
		out.Exercises = append(out.Exercises, x)
	}
	rows.Close()
	for index := range out.Warmups {
		if e = s.pool.QueryRow(ctx, `SELECT count(*) FROM workout_exercises WHERE workout_id=$1::uuid`, out.Warmups[index].ID).Scan(&out.Warmups[index].ExerciseCount); e != nil {
			return out, e
		}
	}
	return out, nil
}

func optionArgs(index int, role Role, user string) []any {
	if index == 0 {
		return nil
	}
	return []any{role.CanManageAll(), user}
}

func (s *Service) List(ctx context.Context, kind, user string, role Role, search, status string) ([]Item, error) {
	tenant, e := tenantID(ctx)
	if e != nil {
		return nil, e
	}
	x, ok := tables[kind]
	if !ok {
		return nil, ErrInvalid
	}
	if status != "" && !oneOf(status, "draft", "published", "archived") {
		return nil, ErrInvalid
	}
	where := " WHERE tenant_id=$1::uuid"
	args := []any{tenant}
	if kind == "exercises" {
		where = " WHERE (tenant_id=$1::uuid OR (tenant_id IS NULL AND owner_user_id IS NULL AND standard_key IS NOT NULL))"
	}
	if search != "" {
		args = append(args, "%"+search+"%")
		where += fmt.Sprintf(" AND "+x.name+" ILIKE $%d", len(args))
	}
	if status != "" {
		args = append(args, status)
		where += fmt.Sprintf(" AND status=$%d", len(args))
	}
	q := fmt.Sprintf(`SELECT id::text,%s,%s,status,%s,owner_user_id::text,updated_at FROM %s%s ORDER BY updated_at DESC`, x.name, x.slug, x.difficulty, x.listFrom, where)
	rows, e := s.pool.Query(ctx, q, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Item{}
	for rows.Next() {
		var i Item
		if e = rows.Scan(&i.ID, &i.Name, &i.Slug, &i.Status, &i.Difficulty, &i.OwnerUserID, &i.UpdatedAt); e != nil {
			return nil, e
		}
		out = append(out, i)
	}
	return out, rows.Err()
}
func validBlocks(blocks []Block) bool {
	allowed := map[string]bool{"heading": true, "text": true, "image": true, "video": true, "tip": true, "warning": true, "checklist": true, "divider": true}
	for _, b := range blocks {
		if !allowed[b.Type] {
			return false
		}
		if (b.Type == "image" || b.Type == "video") && b.MediaID == nil {
			return false
		}
		if b.Type == "checklist" && len(b.Items) == 0 {
			return false
		}
		if b.Type != "image" && b.Type != "video" && b.Type != "divider" && b.Type != "checklist" && strings.TrimSpace(b.Text) == "" {
			return false
		}
	}
	return true
}
func (s *Service) SaveLesson(ctx context.Context, user string, role Role, id *string, in LessonInput) (string, error) {
	tenant, tenantErr := tenantID(ctx)
	if tenantErr != nil {
		return "", tenantErr
	}
	role = "coach" // platform role must never bypass tenant ownership
	if !validBlocks(in.Blocks) || strings.TrimSpace(in.Title) == "" || in.DurationMinutes < 1 || !validID(in.CategoryID) || !validOptionalID(in.CoverMediaID) || !oneOf(in.Difficulty, "beginner", "intermediate", "advanced") {
		return "", ErrInvalid
	}
	if in.Slug == "" {
		var e error
		in.Slug, e = s.uniqueSlug(ctx, "lessons", in.Title, id)
		if e != nil {
			return "", e
		}
	}
	blocks, _ := json.Marshal(in.Blocks)
	if id == nil {
		var out string
		e := s.pool.QueryRow(ctx, `INSERT INTO lessons(category_id,title,slug,short_description,content,difficulty,duration_minutes,owner_user_id,tenant_id,status,content_blocks,cover_media_id) VALUES($1::uuid,$2,$3,$4,$5,$6,$7,$8::uuid,$9::uuid,'draft',$10::jsonb,$11::uuid) RETURNING id::text`, in.CategoryID, in.Title, in.Slug, in.ShortDescription, in.Content, in.Difficulty, in.DurationMinutes, user, tenant, blocks, in.CoverMediaID).Scan(&out)
		return out, e
	}
	var out string
	e := s.pool.QueryRow(ctx, `UPDATE lessons SET category_id=$3::uuid,title=$4,slug=$5,short_description=$6,content=$7,difficulty=$8,duration_minutes=$9,content_blocks=$10::jsonb,cover_media_id=$11::uuid WHERE id=$1::uuid AND tenant_id=$2::uuid RETURNING id::text`, *id, tenant, in.CategoryID, in.Title, in.Slug, in.ShortDescription, in.Content, in.Difficulty, in.DurationMinutes, blocks, in.CoverMediaID).Scan(&out)
	if errors.Is(e, pgx.ErrNoRows) {
		return "", ErrForbidden
	}
	return out, e
}
func validExercises(items []BuilderExercise) bool {
	seen := map[int]bool{}
	exercises := map[string]bool{}
	for _, x := range items {
		if !validID(x.ExerciseID) || exercises[x.ExerciseID] || x.Sets < 1 || x.RestSeconds < 0 || x.SortOrder < 0 || seen[x.SortOrder] || (x.TargetReps == nil) == (x.TargetDurationSeconds == nil) {
			return false
		}
		seen[x.SortOrder] = true
		exercises[x.ExerciseID] = true
		if x.TargetReps != nil && *x.TargetReps < 1 {
			return false
		}
		if x.TargetDurationSeconds != nil && *x.TargetDurationSeconds < 1 {
			return false
		}
	}
	return true
}

func validateWorkoutInput(in BuilderInput, publish bool) error {
	if publish && strings.TrimSpace(in.Title) == "" {
		return invalid("Укажите название тренировки.")
	}
	if publish && strings.TrimSpace(in.Description) == "" {
		return invalid("Укажите описание тренировки.")
	}
	if in.EstimatedMinutes < 1 {
		return invalid("Укажите длительность тренировки.")
	}
	if !oneOf(in.Category, "warmup", "morning", "strength", "skill") {
		return invalid("Выберите категорию тренировки.")
	}
	for index, exercise := range in.Exercises {
		position := index + 1
		if !validID(exercise.ExerciseID) {
			return invalid(fmt.Sprintf("Выберите упражнение в позиции %d.", position))
		}
		if exercise.Sets < 1 {
			return invalid(fmt.Sprintf("У упражнения в позиции %d не указано количество подходов.", position))
		}
		if (exercise.TargetReps == nil) == (exercise.TargetDurationSeconds == nil) {
			return invalid(fmt.Sprintf("У упражнения в позиции %d укажите повторения или длительность.", position))
		}
		if exercise.RestSeconds < 0 {
			return invalid(fmt.Sprintf("У упражнения в позиции %d некорректное время отдыха.", position))
		}
	}
	if !validExercises(in.Exercises) {
		return invalid("Проверьте порядок упражнений.")
	}
	return nil
}

func validProgramWorkouts(items []ProgramWorkout) bool {
	seenIDs, seenOrder := map[string]bool{}, map[int]bool{}
	for _, item := range items {
		if !validID(item.WorkoutID) || item.SortOrder < 0 || seenIDs[item.WorkoutID] || seenOrder[item.SortOrder] {
			return false
		}
		seenIDs[item.WorkoutID], seenOrder[item.SortOrder] = true, true
	}
	return true
}

func normalizedProgramLevels(in BuilderInput) []BuilderLevel {
	if len(in.Levels) > 0 {
		return in.Levels
	}
	return []BuilderLevel{{LevelNumber: 1, Title: "Тренировки", Description: "Этап программы", Difficulty: in.Difficulty, UnlockRuleType: "none", SortOrder: 0, Workouts: in.Workouts}}
}

func validProgramLevels(levels []BuilderLevel) bool {
	seenLevels, seenOrder, seenWorkouts := map[int]bool{}, map[int]bool{}, map[string]bool{}
	for _, level := range levels {
		if level.LevelNumber < 1 || level.SortOrder < 0 || seenLevels[level.LevelNumber] || seenOrder[level.SortOrder] || strings.TrimSpace(level.Title) == "" || strings.TrimSpace(level.Description) == "" || !oneOf(level.Difficulty, "beginner", "intermediate", "advanced") || !oneOf(level.UnlockRuleType, "none", "previous_level", "workouts_completed", "criterion") || level.UnlockRuleValue < 0 || !validProgramWorkouts(level.Workouts) {
			return false
		}
		seenLevels[level.LevelNumber], seenOrder[level.SortOrder] = true, true
		for _, workout := range level.Workouts {
			if seenWorkouts[workout.WorkoutID] {
				return false
			}
			seenWorkouts[workout.WorkoutID] = true
		}
	}
	return len(levels) > 0
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func validBuilderEnums(kind string, in BuilderInput) bool {
	if !oneOf(in.Difficulty, "beginner", "intermediate", "advanced") {
		return false
	}
	if kind == "exercises" && (!oneOf(in.MovementType, "reps", "duration", "distance", "custom") || !validTags(in.Tags)) {
		return false
	}
	if kind == "exercises" && !oneOf(in.MovementType, "reps", "duration", "distance", "custom") {
		return false
	}
	if kind == "workouts" && !oneOf(in.Category, "warmup", "morning", "strength", "skill") {
		return false
	}
	if (kind == "programs" || kind == "skills") && !oneOf(in.Category, "MORNING_ROUTINE", "WARMUP", "BASE_STRENGTH", "SKILL", "MOBILITY", "OTHER") {
		return false
	}
	if kind == "skills" && !oneOf(in.FinalCriterionType, "duration_seconds", "repetitions", "manual_confirmation", "workout_count", "exercise_reps", "exercise_duration", "skill_hold_duration", "manual_user_confirmation", "manual_coach_confirmation") {
		return false
	}
	for _, level := range in.Levels {
		if level.LevelNumber < 1 || level.SortOrder < 0 || strings.TrimSpace(level.Title) == "" {
			return false
		}
		if kind == "programs" && (!oneOf(level.Difficulty, "beginner", "intermediate", "advanced") || !oneOf(level.UnlockRuleType, "none", "previous_level", "workouts_completed", "criterion") || level.UnlockRuleValue < 0) {
			return false
		}
		if kind == "skills" && (!oneOf(level.CriterionType, "workout_completed", "duration_seconds", "repetitions", "manual_confirmation", "workout_count", "exercise_reps", "exercise_duration", "skill_hold_duration", "manual_user_confirmation", "manual_coach_confirmation") || level.CriterionValue < 1 || (level.CriterionType == "workout_completed" && level.ProgramLevelID == nil)) {
			return false
		}
	}
	return true
}

func (s *Service) SaveBuilder(ctx context.Context, kind, user string, role Role, id *string, in BuilderInput) (string, error) {
	tenant, tenantErr := tenantID(ctx)
	if tenantErr != nil {
		return "", tenantErr
	}
	_ = tenant
	role = "coach" // platform role must never bypass tenant ownership
	if id != nil {
		table, ok := tables[kind]
		if !ok {
			return "", ErrInvalid
		}
		var accessible bool
		query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE id=$1::uuid AND tenant_id=$2::uuid AND owner_user_id=$3::uuid)`, table.table)
		if e := s.pool.QueryRow(ctx, query, *id, tenant, user).Scan(&accessible); e != nil {
			return "", e
		}
		if !accessible {
			return "", ErrForbidden
		}
	}
	if kind == "workouts" {
		if err := validateWorkoutInput(in, false); err != nil {
			return "", err
		}
	}
	if (kind != "workouts" && strings.TrimSpace(in.Description) == "") || !validBuilderEnums(kind, in) || !validOptionalID(in.CoverMediaID) || !validOptionalID(in.DemoMediaID) {
		return "", ErrInvalid
	}
	if kind == "exercises" && in.DemoMediaID != nil {
		var ok bool
		if e := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM media_assets WHERE id=$1::uuid AND tenant_id=$2::uuid AND status='ready' AND mime_type IN ('video/mp4','video/webm','image/gif','image/jpeg','image/png','image/webp') AND size_bytes<=5242880 AND (type='image' OR duration_seconds BETWEEN 1 AND 6))`, *in.DemoMediaID, tenant).Scan(&ok); e != nil || !ok {
			return "", ErrInvalid
		}
	}
	if (kind == "exercises" || kind == "programs") && in.Slug == "" {
		var e error
		in.Slug, e = s.uniqueSlug(ctx, map[string]string{"exercises": "exercises", "programs": "programs"}[kind], in.Name, id)
		if e != nil {
			return "", e
		}
	}
	tx, e := s.pool.Begin(ctx)
	if e != nil {
		return "", e
	}
	defer tx.Rollback(ctx)
	var out string
	switch kind {
	case "exercises":
		if in.Name == "" || len(in.MuscleGroups) == 0 {
			return "", ErrInvalid
		}
		if id == nil {
			e = tx.QueryRow(ctx, `INSERT INTO exercises(name,slug,description,instructions,common_mistakes,difficulty,muscle_groups,equipment,tags,owner_user_id,tenant_id,status,movement_type,coach_tips,cover_media_id,demo_media_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::uuid,$11::uuid,'draft',$12,$13,$14::uuid,$15::uuid) RETURNING id::text`, in.Name, in.Slug, in.Description, in.Instructions, in.CommonMistakes, in.Difficulty, in.MuscleGroups, in.Equipment, in.Tags, user, tenant, in.MovementType, in.CoachTips, in.CoverMediaID, in.DemoMediaID).Scan(&out)
		} else {
			e = tx.QueryRow(ctx, `UPDATE exercises SET name=$4,slug=$5,description=$6,instructions=$7,common_mistakes=$8,difficulty=$9,muscle_groups=$10,equipment=$11,tags=$12,movement_type=$13,coach_tips=$14,cover_media_id=$15::uuid,demo_media_id=$16::uuid WHERE id=$1::uuid AND tenant_id=$2::uuid AND owner_user_id=$3::uuid RETURNING id::text`, *id, tenant, user, in.Name, in.Slug, in.Description, in.Instructions, in.CommonMistakes, in.Difficulty, in.MuscleGroups, in.Equipment, in.Tags, in.MovementType, in.CoachTips, in.CoverMediaID, in.DemoMediaID).Scan(&out)
		}
	case "programs":
		levels := normalizedProgramLevels(in)
		if in.Name == "" || in.DurationWeeks < 1 || !validProgramLevels(levels) {
			return "", ErrInvalid
		}
		if id == nil {
			e = tx.QueryRow(ctx, `INSERT INTO programs(name,slug,description,difficulty,duration_weeks,category,owner_user_id,tenant_id,status,cover_media_id) VALUES($1,$2,$3,$4,$5,$6,$7::uuid,$8::uuid,'draft',$9::uuid) RETURNING id::text`, in.Name, in.Slug, in.Description, in.Difficulty, in.DurationWeeks, in.Category, user, tenant, in.CoverMediaID).Scan(&out)
		} else {
			e = tx.QueryRow(ctx, `UPDATE programs SET name=$4,slug=$5,description=$6,difficulty=$7,duration_weeks=$8,category=$9,cover_media_id=$10::uuid WHERE id=$1::uuid AND tenant_id=$2::uuid AND owner_user_id=$3::uuid RETURNING id::text`, *id, tenant, user, in.Name, in.Slug, in.Description, in.Difficulty, in.DurationWeeks, in.Category, in.CoverMediaID).Scan(&out)
		}
		if e == nil {
			_, e = tx.Exec(ctx, `UPDATE workouts SET program_id=NULL,program_level_id=NULL,day_number=NULL,sort_order=0 WHERE program_id=$1::uuid`, out)
			levelNumbers := make([]int32, 0, len(levels))
			dayNumber := 1
			for _, level := range levels {
				if e != nil {
					break
				}
				levelNumbers = append(levelNumbers, int32(level.LevelNumber))
				var levelID string
				e = tx.QueryRow(ctx, `INSERT INTO program_levels(program_id,level_number,title,description,difficulty,unlock_rule_type,unlock_rule_value,sort_order) VALUES($1::uuid,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT(program_id,level_number) DO UPDATE SET title=EXCLUDED.title,description=EXCLUDED.description,difficulty=EXCLUDED.difficulty,unlock_rule_type=EXCLUDED.unlock_rule_type,unlock_rule_value=EXCLUDED.unlock_rule_value,sort_order=EXCLUDED.sort_order RETURNING id::text`, out, level.LevelNumber, level.Title, level.Description, level.Difficulty, level.UnlockRuleType, level.UnlockRuleValue, level.SortOrder).Scan(&levelID)
				for _, workout := range level.Workouts {
					if e != nil {
						break
					}
					tag, updateErr := tx.Exec(ctx, `UPDATE workouts SET program_id=$1::uuid,program_level_id=$2::uuid,day_number=$3,sort_order=$4 WHERE id=$5::uuid AND tenant_id=$6::uuid AND owner_user_id=$7::uuid`, out, levelID, dayNumber, workout.SortOrder, workout.WorkoutID, tenant, user)
					e = updateErr
					if e == nil && tag.RowsAffected() == 0 {
						return "", ErrForbidden
					}
					dayNumber++
				}
			}
			if e == nil {
				_, e = tx.Exec(ctx, `DELETE FROM program_levels WHERE program_id=$1::uuid AND NOT(level_number=ANY($2::int[]))`, out, levelNumbers)
			}
		}
	case "workouts":
		warmupEnabled := in.Category != "warmup"
		if in.WarmupEnabled != nil {
			warmupEnabled = *in.WarmupEnabled && in.Category != "warmup"
		}
		if !warmupEnabled {
			in.WarmupWorkoutID = nil
		}
		if in.WarmupWorkoutID != nil {
			if !validID(*in.WarmupWorkoutID) || (id != nil && *in.WarmupWorkoutID == *id) {
				return "", invalid("Выберите другую разминку.")
			}
			var validWarmup bool
			e = tx.QueryRow(ctx, `SELECT category='warmup' AND status<>'archived' AND NOT warmup_enabled FROM workouts WHERE id=$1::uuid AND tenant_id=$2::uuid`, *in.WarmupWorkoutID, tenant).Scan(&validWarmup)
			if errors.Is(e, pgx.ErrNoRows) {
				return "", invalid("Выбранная разминка недоступна или имеет неверную категорию.")
			}
			if e != nil {
				return "", e
			}
			if !validWarmup {
				return "", invalid("Выбранная разминка недоступна или имеет неверную категорию.")
			}
		}
		if id == nil {
			e = tx.QueryRow(ctx, `INSERT INTO workouts(title,description,difficulty,estimated_minutes,owner_user_id,tenant_id,status,cover_media_id,category,warmup_enabled,warmup_workout_id) VALUES($1,$2,$3,$4,$5::uuid,$6::uuid,'draft',$7::uuid,$8,$9,$10::uuid) RETURNING id::text`, in.Title, in.Description, in.Difficulty, in.EstimatedMinutes, user, tenant, in.CoverMediaID, in.Category, warmupEnabled, in.WarmupWorkoutID).Scan(&out)
		} else {
			e = tx.QueryRow(ctx, `UPDATE workouts SET title=$4,description=$5,difficulty=$6,estimated_minutes=$7,cover_media_id=$8::uuid,category=$9,warmup_enabled=$10,warmup_workout_id=$11::uuid,is_default_warmup=CASE WHEN $9='warmup' THEN is_default_warmup ELSE false END WHERE id=$1::uuid AND tenant_id=$2::uuid AND owner_user_id=$3::uuid RETURNING id::text`, *id, tenant, user, in.Title, in.Description, in.Difficulty, in.EstimatedMinutes, in.CoverMediaID, in.Category, warmupEnabled, in.WarmupWorkoutID).Scan(&out)
		}
		if e == nil {
			_, e = tx.Exec(ctx, `DELETE FROM workout_exercises WHERE workout_id=$1::uuid`, out)
			for _, x := range in.Exercises {
				if e != nil {
					break
				}
				_, e = tx.Exec(ctx, `INSERT INTO workout_exercises(workout_id,exercise_id,sets,target_reps,target_duration_seconds,rest_seconds,notes,sort_order) VALUES($1::uuid,$2::uuid,$3,$4,$5,$6,$7,$8)`, out, x.ExerciseID, x.Sets, x.TargetReps, x.TargetDurationSeconds, x.RestSeconds, x.Notes, x.SortOrder)
			}
		}
	case "skills":
		if in.MapGroup == "" {
			in.MapGroup = "basic"
		}
		if in.Name == "" || in.Icon == "" || in.FinalCriterionValue < 1 || len(in.Levels) == 0 || !validMapGroup(in.MapGroup) {
			return "", ErrInvalid
		}
		for _, level := range in.Levels {
			if !validOptionalID(level.ProgramLevelID) {
				return "", ErrInvalid
			}
		}
		for _, requirement := range in.Requirements {
			if !validID(requirement) {
				return "", ErrInvalid
			}
		}
		if in.Code == "" {
			in.Code = strings.ToUpper(strings.ReplaceAll(slugBase(in.Name), "-", "_"))
		}
		var currentOrder int
		if id != nil {
			_ = tx.QueryRow(ctx, `SELECT sort_order FROM skills WHERE id=$1::uuid AND tenant_id=$2::uuid`, *id, tenant).Scan(&currentOrder)
		}
		if in.SortOrder < 1 {
			e = tx.QueryRow(ctx, `SELECT COALESCE(max(sort_order),0)+1 FROM skills WHERE tenant_id=$1::uuid`, tenant).Scan(&in.SortOrder)
		}
		if e == nil && in.SortOrder != currentOrder {
			_, e = tx.Exec(ctx, `UPDATE skills SET sort_order=sort_order+1 WHERE tenant_id=$3::uuid AND sort_order >= $1 AND ($2::uuid IS NULL OR id<>$2::uuid)`, in.SortOrder, id, tenant)
		}
		if id == nil {
			e = tx.QueryRow(ctx, `INSERT INTO skills(code,name,description,category,map_group,difficulty,icon,xp_reward,final_criterion_type,final_criterion_value,owner_user_id,tenant_id,status,cover_media_id,sort_order) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::uuid,$12::uuid,'draft',$13::uuid,$14) RETURNING id::text`, in.Code, in.Name, in.Description, in.Category, in.MapGroup, in.Difficulty, in.Icon, in.XPReward, in.FinalCriterionType, in.FinalCriterionValue, user, tenant, in.CoverMediaID, in.SortOrder).Scan(&out)
		} else {
			e = tx.QueryRow(ctx, `UPDATE skills SET code=$4,name=$5,description=$6,category=$7,map_group=$8,difficulty=$9,icon=$10,xp_reward=$11,final_criterion_type=$12,final_criterion_value=$13,cover_media_id=$14::uuid,sort_order=$15 WHERE id=$1::uuid AND tenant_id=$2::uuid AND owner_user_id=$3::uuid RETURNING id::text`, *id, tenant, user, in.Code, in.Name, in.Description, in.Category, in.MapGroup, in.Difficulty, in.Icon, in.XPReward, in.FinalCriterionType, in.FinalCriterionValue, in.CoverMediaID, in.SortOrder).Scan(&out)
		}
		if e == nil {
			_, e = tx.Exec(ctx, `DELETE FROM skill_requirements WHERE skill_id=$1::uuid`, out)
			levelNumbers := make([]int32, 0, len(in.Levels))
			for _, l := range in.Levels {
				if e != nil {
					break
				}
				levelNumbers = append(levelNumbers, int32(l.LevelNumber))
				_, e = tx.Exec(ctx, `INSERT INTO skill_levels(skill_id,level_number,name,description,program_level_id,criterion_type,criterion_value,sort_order) VALUES($1::uuid,$2,$3,$4,$5::uuid,$6,$7,$8) ON CONFLICT(skill_id,level_number) DO UPDATE SET name=EXCLUDED.name,description=EXCLUDED.description,program_level_id=EXCLUDED.program_level_id,criterion_type=EXCLUDED.criterion_type,criterion_value=EXCLUDED.criterion_value,sort_order=EXCLUDED.sort_order`, out, l.LevelNumber, l.Title, l.Description, l.ProgramLevelID, l.CriterionType, l.CriterionValue, l.SortOrder)
			}
			if e == nil {
				_, e = tx.Exec(ctx, `DELETE FROM skill_levels WHERE skill_id=$1::uuid AND NOT(level_number=ANY($2::int[]))`, out, levelNumbers)
			}
			for _, req := range in.Requirements {
				if req == out {
					return "", ErrInvalid
				}
				if e != nil {
					break
				}
				var cycle bool
				e = tx.QueryRow(ctx, `WITH RECURSIVE graph(id) AS (SELECT required_skill_id FROM skill_requirements WHERE skill_id=$1::uuid UNION SELECT sr.required_skill_id FROM skill_requirements sr JOIN graph g ON sr.skill_id=g.id) SELECT EXISTS(SELECT 1 FROM graph WHERE id=$2::uuid)`, req, out).Scan(&cycle)
				if e == nil && cycle {
					return "", ErrInvalid
				}
				if e == nil {
					_, e = tx.Exec(ctx, `INSERT INTO skill_requirements(skill_id,required_skill_id,requirement_type) VALUES($1::uuid,$2::uuid,'skill_mastered')`, out, req)
				}
			}
		}
	default:
		return "", ErrInvalid
	}
	if errors.Is(e, pgx.ErrNoRows) {
		return "", ErrForbidden
	}
	if e != nil {
		return "", e
	}
	if e = tx.Commit(ctx); e != nil {
		return "", e
	}
	return out, nil
}
func (s *Service) Lifecycle(ctx context.Context, kind, id, user string, role Role, status string) error {
	tenant, e := tenantID(ctx)
	if e != nil {
		return e
	}
	role = "coach"
	x, ok := tables[kind]
	if !ok || (status != "draft" && status != "published" && status != "archived") {
		return ErrInvalid
	}
	var accessible bool
	check := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE id=$1::uuid AND tenant_id=$2::uuid AND owner_user_id=$3::uuid)`, x.table)
	if e = s.pool.QueryRow(ctx, check, id, tenant, user).Scan(&accessible); e != nil {
		return e
	}
	if !accessible {
		return ErrForbidden
	}
	if status == "published" {
		if e := s.validatePublish(ctx, kind, id); e != nil {
			return e
		}
	}
	if status == "archived" {
		if e := s.validateArchive(ctx, kind, id); e != nil {
			return e
		}
	}
	q := fmt.Sprintf(`UPDATE %s SET status=$2,published_by=CASE WHEN $2='published' THEN $3::uuid ELSE published_by END,published_at=CASE WHEN $2='published' THEN NOW() ELSE published_at END WHERE id=$1::uuid AND tenant_id=$4::uuid AND owner_user_id=$3::uuid`, x.table)
	tag, e := s.pool.Exec(ctx, q, id, status, user, tenant)
	if e != nil {
		return e
	}
	if tag.RowsAffected() == 0 {
		return ErrForbidden
	}
	return nil
}

func (s *Service) validatePublish(ctx context.Context, kind, id string) error {
	var names []string
	var valid bool
	switch kind {
	case "lessons":
		e := s.pool.QueryRow(ctx, `SELECT content<>'' OR jsonb_array_length(content_blocks)>0 FROM lessons WHERE id=$1::uuid`, id).Scan(&valid)
		if e != nil || !valid {
			return invalid("Добавьте содержимое урока перед публикацией.")
		}
	case "exercises":
		e := s.pool.QueryRow(ctx, `SELECT name<>'' AND description<>'' AND instructions<>'' AND cardinality(muscle_groups)>0 FROM exercises WHERE id=$1::uuid`, id).Scan(&valid)
		if e != nil || !valid {
			return invalid("Заполните название, описание, инструкцию и группы мышц упражнения.")
		}
	case "workouts":
		e := s.pool.QueryRow(ctx, `SELECT title<>'' AND description<>'' AND estimated_minutes>0 AND category IN ('warmup','morning','strength','skill') FROM workouts WHERE id=$1::uuid`, id).Scan(&valid)
		if e != nil || !valid {
			return invalid("Заполните название, описание, категорию и длительность тренировки.")
		}
		rows, e := s.pool.Query(ctx, `SELECT e.name FROM workout_exercises we JOIN exercises e ON e.id=we.exercise_id WHERE we.workout_id=$1::uuid AND e.status<>'published' ORDER BY we.sort_order`, id)
		if e != nil {
			return e
		}
		for rows.Next() {
			var name string
			if e = rows.Scan(&name); e != nil {
				rows.Close()
				return e
			}
			names = append(names, name)
		}
		rows.Close()
		if len(names) > 0 {
			return invalid("Сначала опубликуйте упражнения: " + strings.Join(names, ", ") + ".")
		}
		e = s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM workout_exercises WHERE workout_id=$1::uuid)`, id).Scan(&valid)
		if e != nil || !valid {
			return invalid("Добавьте хотя бы одно упражнение перед публикацией тренировки.")
		}
		var warmupEnabled bool
		var warmupID *string
		e = s.pool.QueryRow(ctx, `SELECT warmup_enabled,warmup_workout_id::text FROM workouts WHERE id=$1::uuid`, id).Scan(&warmupEnabled, &warmupID)
		if e != nil {
			return e
		}
		if warmupEnabled && warmupID != nil {
			e = s.pool.QueryRow(ctx, `SELECT category='warmup' AND status='published' AND NOT warmup_enabled AND warmup_workout_id IS NULL FROM workouts WHERE id=$1::uuid`, *warmupID).Scan(&valid)
			if e != nil || !valid {
				return invalid("Сначала опубликуйте выбранную разминку и отключите у неё собственную разминку.")
			}
		}
	case "programs":
		rows, e := s.pool.Query(ctx, `SELECT w.title FROM program_levels pl JOIN workouts w ON w.program_level_id=pl.id WHERE pl.program_id=$1::uuid AND w.status<>'published' ORDER BY pl.sort_order,w.sort_order`, id)
		if e != nil {
			return e
		}
		for rows.Next() {
			var name string
			if e = rows.Scan(&name); e != nil {
				rows.Close()
				return e
			}
			names = append(names, name)
		}
		rows.Close()
		if len(names) > 0 {
			return invalid("Сначала опубликуйте тренировки: " + strings.Join(names, ", ") + ".")
		}
		rows, e = s.pool.Query(ctx, `SELECT pl.title FROM program_levels pl WHERE pl.program_id=$1::uuid AND NOT EXISTS(SELECT 1 FROM workouts w WHERE w.program_level_id=pl.id) ORDER BY pl.sort_order`, id)
		if e != nil {
			return e
		}
		names = names[:0]
		for rows.Next() {
			var name string
			if e = rows.Scan(&name); e != nil {
				rows.Close()
				return e
			}
			names = append(names, name)
		}
		rows.Close()
		if len(names) > 0 {
			return invalid("Добавьте тренировку в каждый этап программы: " + strings.Join(names, ", ") + ".")
		}
		e = s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM program_levels pl JOIN workouts w ON w.program_level_id=pl.id WHERE pl.program_id=$1::uuid)`, id).Scan(&valid)
		if e != nil || !valid {
			return invalid("Добавьте хотя бы одну тренировку.")
		}
	case "skills":
		rows, e := s.pool.Query(ctx, `SELECT s.name FROM skill_requirements sr JOIN skills s ON s.id=sr.required_skill_id WHERE sr.skill_id=$1::uuid AND s.status<>'published' UNION ALL SELECT p.name FROM skill_levels sl JOIN program_levels pl ON pl.id=sl.program_level_id JOIN programs p ON p.id=pl.program_id WHERE sl.skill_id=$1::uuid AND p.status<>'published'`, id)
		if e != nil {
			return e
		}
		for rows.Next() {
			var name string
			if e = rows.Scan(&name); e != nil {
				rows.Close()
				return e
			}
			names = append(names, name)
		}
		rows.Close()
		if len(names) > 0 {
			return invalid("Сначала опубликуйте зависимости прогрессии: " + strings.Join(names, ", ") + ".")
		}
		e = s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM skill_levels WHERE skill_id=$1::uuid)`, id).Scan(&valid)
		if e != nil || !valid {
			return invalid("Добавьте хотя бы один этап прогрессии.")
		}
	}
	return nil
}

func (s *Service) validateArchive(ctx context.Context, kind, id string) error {
	var count int
	var label string
	switch kind {
	case "exercises":
		label = "опубликованных тренировках"
		_ = s.pool.QueryRow(ctx, `SELECT count(DISTINCT w.id) FROM workout_exercises we JOIN workouts w ON w.id=we.workout_id WHERE we.exercise_id=$1::uuid AND w.status='published'`, id).Scan(&count)
	case "workouts":
		label = "опубликованных программах"
		_ = s.pool.QueryRow(ctx, `SELECT count(DISTINCT p.id) FROM workouts w JOIN program_levels pl ON pl.id=w.program_level_id JOIN programs p ON p.id=pl.program_id WHERE w.id=$1::uuid AND p.status='published'`, id).Scan(&count)
	case "skills":
		label = "опубликованных прогрессиях"
		_ = s.pool.QueryRow(ctx, `SELECT count(DISTINCT s.id) FROM skill_requirements sr JOIN skills s ON s.id=sr.skill_id WHERE sr.required_skill_id=$1::uuid AND s.status='published'`, id).Scan(&count)
	}
	if count > 0 {
		return fmt.Errorf("%w: Материал используется в %d %s. Сначала снимите зависимый контент с публикации.", ErrDependency, count, label)
	}
	return nil
}
func (s *Service) Duplicate(ctx context.Context, kind, id, user string, role Role) (string, error) {
	tenant, e := tenantID(ctx)
	if e != nil {
		return "", e
	}
	role = "coach"
	x, ok := tables[kind]
	if !ok {
		return "", ErrInvalid
	}
	if kind == "lessons" {
		var out string
		e := s.pool.QueryRow(ctx, `INSERT INTO lessons(category_id,title,slug,short_description,content,difficulty,duration_minutes,sort_order,published,owner_user_id,status,content_blocks,cover_media_id,tenant_id) SELECT category_id,title||' — копия',slug||'-copy-'||substr(gen_random_uuid()::text,1,8),short_description,content,difficulty,duration_minutes,sort_order,false,$2::uuid,'draft',content_blocks,cover_media_id,$3::uuid FROM lessons WHERE id=$1::uuid AND tenant_id=$3::uuid AND owner_user_id=$2::uuid RETURNING id::text`, id, user, tenant).Scan(&out)
		if errors.Is(e, pgx.ErrNoRows) {
			return "", ErrForbidden
		}
		return out, e
	}
	if kind == "exercises" {
		var out string
		e := s.pool.QueryRow(ctx, `INSERT INTO exercises(name,slug,description,instructions,common_mistakes,difficulty,muscle_groups,equipment,tags,video_url,image_url,owner_user_id,status,movement_type,coach_tips,cover_media_id,tenant_id) SELECT name||' — копия',slug||'-copy-'||substr(gen_random_uuid()::text,1,8),description,instructions,common_mistakes,difficulty,muscle_groups,equipment,tags,video_url,image_url,$2::uuid,'draft',movement_type,coach_tips,cover_media_id,$3::uuid FROM exercises WHERE id=$1::uuid AND tenant_id=$3::uuid AND owner_user_id=$2::uuid RETURNING id::text`, id, user, tenant).Scan(&out)
		if errors.Is(e, pgx.ErrNoRows) {
			return "", ErrForbidden
		}
		return out, e
	}
	if kind == "workouts" {
		tx, e := s.pool.Begin(ctx)
		if e != nil {
			return "", e
		}
		defer tx.Rollback(ctx)
		var out string
		e = tx.QueryRow(ctx, `INSERT INTO workouts(title,description,difficulty,estimated_minutes,sort_order,owner_user_id,status,cover_media_id,category,warmup_enabled,warmup_workout_id,tenant_id) SELECT title||' — копия',description,difficulty,estimated_minutes,0,$2::uuid,'draft',cover_media_id,category,warmup_enabled,warmup_workout_id,$3::uuid FROM workouts w WHERE id=$1::uuid AND tenant_id=$3::uuid AND owner_user_id=$2::uuid RETURNING id::text`, id, user, tenant).Scan(&out)
		if errors.Is(e, pgx.ErrNoRows) {
			return "", ErrForbidden
		}
		if e != nil {
			return "", e
		}
		_, e = tx.Exec(ctx, `INSERT INTO workout_exercises(workout_id,exercise_id,sets,target_reps,target_duration_seconds,rest_seconds,sort_order,notes) SELECT $1::uuid,exercise_id,sets,target_reps,target_duration_seconds,rest_seconds,sort_order,notes FROM workout_exercises WHERE workout_id=$2::uuid`, out, id)
		if e != nil {
			return "", e
		}
		return out, tx.Commit(ctx)
	}
	if kind == "programs" {
		tx, e := s.pool.Begin(ctx)
		if e != nil {
			return "", e
		}
		defer tx.Rollback(ctx)
		var out string
		e = tx.QueryRow(ctx, `INSERT INTO programs(name,slug,description,difficulty,duration_weeks,published,category,owner_user_id,status,cover_media_id,coach_description,tenant_id) SELECT name||' — копия',slug||'-copy-'||substr(gen_random_uuid()::text,1,8),description,difficulty,duration_weeks,false,category,$2::uuid,'draft',cover_media_id,coach_description,$3::uuid FROM programs WHERE id=$1::uuid AND tenant_id=$3::uuid AND owner_user_id=$2::uuid RETURNING id::text`, id, user, tenant).Scan(&out)
		if errors.Is(e, pgx.ErrNoRows) {
			return "", ErrForbidden
		}
		if e != nil {
			return "", e
		}
		_, e = tx.Exec(ctx, `INSERT INTO program_levels(program_id,level_number,title,description,difficulty,unlock_rule_type,unlock_rule_value,sort_order) SELECT $1::uuid,level_number,title,description,difficulty,unlock_rule_type,unlock_rule_value,sort_order FROM program_levels WHERE program_id=$2::uuid`, out, id)
		if e != nil {
			return "", e
		}
		return out, tx.Commit(ctx)
	}
	if kind == "skills" {
		tx, e := s.pool.Begin(ctx)
		if e != nil {
			return "", e
		}
		defer tx.Rollback(ctx)
		var out string
		e = tx.QueryRow(ctx, `INSERT INTO skills(code,name,description,category,map_group,difficulty,icon,xp_reward,final_criterion_type,final_criterion_value,owner_user_id,status,cover_media_id,tenant_id) SELECT code||'_COPY_'||upper(substr(replace(gen_random_uuid()::text,'-',''),1,6)),name||' — копия',description,category,map_group,difficulty,icon,xp_reward,final_criterion_type,final_criterion_value,$2::uuid,'draft',cover_media_id,$3::uuid FROM skills WHERE id=$1::uuid AND tenant_id=$3::uuid AND owner_user_id=$2::uuid RETURNING id::text`, id, user, tenant).Scan(&out)
		if errors.Is(e, pgx.ErrNoRows) {
			return "", ErrForbidden
		}
		if e != nil {
			return "", e
		}
		_, e = tx.Exec(ctx, `INSERT INTO skill_levels(skill_id,level_number,name,description,program_level_id,criterion_type,criterion_value,sort_order) SELECT $1::uuid,level_number,name,description,program_level_id,criterion_type,criterion_value,sort_order FROM skill_levels WHERE skill_id=$2::uuid`, out, id)
		if e != nil {
			return "", e
		}
		return out, tx.Commit(ctx)
	}
	return "", fmt.Errorf("%w: duplicate %s is not available", ErrInvalid, x.table)
}
func (s *Service) ListMedia(ctx context.Context, user string, role Role) ([]Media, error) {
	tenant, tenantErr := tenantID(ctx)
	if tenantErr != nil {
		return nil, tenantErr
	}
	rows, e := s.pool.Query(ctx, `SELECT m.id::text,m.owner_user_id::text,m.type,m.status,m.storage_provider,m.storage_key,m.url,m.thumbnail_url,m.original_filename,m.mime_type,m.size_bytes,m.created_at,(SELECT count(*) FROM lessons WHERE tenant_id=$1::uuid AND (cover_media_id=m.id OR content_blocks @> jsonb_build_array(jsonb_build_object('media_id',m.id::text))))+(SELECT count(*) FROM exercises WHERE tenant_id=$1::uuid AND (cover_media_id=m.id OR demo_media_id=m.id))+(SELECT count(*) FROM workouts WHERE tenant_id=$1::uuid AND cover_media_id=m.id)+(SELECT count(*) FROM programs WHERE tenant_id=$1::uuid AND cover_media_id=m.id)+(SELECT count(*) FROM skills WHERE tenant_id=$1::uuid AND cover_media_id=m.id) FROM media_assets m WHERE m.tenant_id=$1::uuid ORDER BY m.created_at DESC`, tenant)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Media{}
	for rows.Next() {
		var m Media
		if e = rows.Scan(&m.ID, &m.OwnerUserID, &m.Type, &m.Status, &m.StorageProvider, &m.StorageKey, &m.URL, &m.ThumbnailURL, &m.OriginalFilename, &m.MIMEType, &m.SizeBytes, &m.CreatedAt, &m.References); e != nil {
			return nil, e
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
func (s *Service) CreateExternalMedia(ctx context.Context, user string, in MediaInput) (Media, error) {
	tenant, tenantErr := tenantID(ctx)
	if tenantErr != nil {
		return Media{}, tenantErr
	}
	isHTTPS := strings.HasPrefix(in.URL, "https://")
	isLocalImage := in.Type == "image" && strings.HasPrefix(in.URL, "data:"+in.MIMEType+";base64,") && in.SizeBytes <= 10<<20
	if (in.Type != "image" && in.Type != "video") || (!isHTTPS && !isLocalImage) || !validMime(in.Type, in.MIMEType) || in.SizeBytes < 0 {
		return Media{}, ErrInvalid
	}
	key := fmt.Sprintf("tenants/%s/media/%d", tenant, time.Now().UnixNano())
	var id string
	e := s.pool.QueryRow(ctx, `INSERT INTO media_assets(owner_user_id,tenant_id,type,storage_provider,storage_key,url,thumbnail_url,original_filename,mime_type,size_bytes,width,height,duration_seconds) VALUES($1::uuid,$2::uuid,$3,'external',$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id::text`, user, tenant, in.Type, key, in.URL, in.ThumbnailURL, safeName(in.OriginalFilename), in.MIMEType, in.SizeBytes, in.Width, in.Height, in.DurationSeconds).Scan(&id)
	if e != nil {
		return Media{}, e
	}
	items, e := s.ListMedia(ctx, user, "coach")
	for _, m := range items {
		if m.ID == id {
			return m, nil
		}
	}
	return Media{}, e
}
func validMime(kind, mime string) bool {
	if kind == "image" {
		return mime == "image/jpeg" || mime == "image/png" || mime == "image/webp" || mime == "image/gif"
	}
	return mime == "video/mp4" || mime == "video/webm"
}
func safeName(v string) string {
	v = strings.ReplaceAll(strings.ReplaceAll(v, "/", "_"), "\\", "_")
	if len(v) > 200 {
		v = v[len(v)-200:]
	}
	if strings.TrimSpace(v) == "" {
		return "media"
	}
	return v
}
func (s *Service) DeleteMedia(ctx context.Context, user, id string, role Role) error {
	tenant, tenantErr := tenantID(ctx)
	if tenantErr != nil {
		return tenantErr
	}
	var refs int
	e := s.pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM lessons WHERE tenant_id=$2::uuid AND (cover_media_id=$1::uuid OR content_blocks @> jsonb_build_array(jsonb_build_object('media_id',$1::text))))+(SELECT count(*) FROM exercises WHERE tenant_id=$2::uuid AND (cover_media_id=$1::uuid OR demo_media_id=$1::uuid))+(SELECT count(*) FROM workouts WHERE tenant_id=$2::uuid AND cover_media_id=$1::uuid)+(SELECT count(*) FROM programs WHERE tenant_id=$2::uuid AND cover_media_id=$1::uuid)+(SELECT count(*) FROM skills WHERE tenant_id=$2::uuid AND cover_media_id=$1::uuid)`, id, tenant).Scan(&refs)
	if e != nil {
		return e
	}
	if refs > 0 {
		return ErrInUse
	}
	tag, e := s.pool.Exec(ctx, `DELETE FROM media_assets WHERE id=$1::uuid AND tenant_id=$2::uuid`, id, tenant)
	if e != nil {
		return e
	}
	if tag.RowsAffected() == 0 {
		return ErrForbidden
	}
	return nil
}
