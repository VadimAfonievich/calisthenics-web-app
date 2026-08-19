package coach

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrForbidden = errors.New("coach content forbidden")
	ErrNotFound  = errors.New("coach content not found")
	ErrInvalid   = errors.New("coach content invalid")
	ErrInUse     = errors.New("media in use")
)

type Role string

func (r Role) CanManageAll() bool { return r == "admin" || r == "super_admin" }

type Dashboard struct {
	Lessons          int `json:"lessons"`
	LessonsPublished int `json:"lessons_published"`
	Exercises        int `json:"exercises"`
	Workouts         int `json:"workouts"`
	Programs         int `json:"programs"`
	Skills           int `json:"skills"`
	Media            int `json:"media"`
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
	ID   string `json:"id"`
	Name string `json:"name"`
}
type Options struct {
	Categories []Option `json:"categories"`
	Exercises  []Option `json:"exercises"`
	Programs   []Option `json:"programs"`
	Workouts   []Option `json:"workouts"`
	Skills     []Option `json:"skills"`
	Media      []Option `json:"media"`
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
	Icon                string            `json:"icon"`
	XPReward            int               `json:"xp_reward"`
	FinalCriterionType  string            `json:"final_criterion_type"`
	FinalCriterionValue int               `json:"final_criterion_value"`
	CoverMediaID        *string           `json:"cover_media_id,omitempty"`
	Exercises           []BuilderExercise `json:"exercises"`
	Levels              []BuilderLevel    `json:"levels"`
	Requirements        []string          `json:"requirements"`
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
	LevelNumber     int     `json:"level_number"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Difficulty      string  `json:"difficulty"`
	UnlockRuleType  string  `json:"unlock_rule_type"`
	UnlockRuleValue int     `json:"unlock_rule_value"`
	CriterionType   string  `json:"criterion_type"`
	CriterionValue  int     `json:"criterion_value"`
	ProgramLevelID  *string `json:"program_level_id,omitempty"`
	SortOrder       int     `json:"sort_order"`
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
func (s *Service) Role(ctx context.Context, user string) (Role, error) {
	var r Role
	e := s.pool.QueryRow(ctx, `SELECT role FROM admin_users WHERE user_id=$1::uuid`, user).Scan(&r)
	if errors.Is(e, pgx.ErrNoRows) {
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
	var d Dashboard
	q := `SELECT (SELECT count(*) FROM lessons WHERE ($2 OR owner_user_id=$1::uuid)),(SELECT count(*) FROM lessons WHERE status='published' AND ($2 OR owner_user_id=$1::uuid)),(SELECT count(*) FROM exercises WHERE ($2 OR owner_user_id=$1::uuid)),(SELECT count(*) FROM workouts WHERE ($2 OR owner_user_id=$1::uuid)),(SELECT count(*) FROM programs WHERE ($2 OR owner_user_id=$1::uuid)),(SELECT count(*) FROM skills WHERE ($2 OR owner_user_id=$1::uuid)),(SELECT count(*) FROM media_assets WHERE ($2 OR owner_user_id=$1::uuid))`
	e := s.pool.QueryRow(ctx, q, user, role.CanManageAll()).Scan(&d.Lessons, &d.LessonsPublished, &d.Exercises, &d.Workouts, &d.Programs, &d.Skills, &d.Media)
	return d, e
}
func (s *Service) Analytics(ctx context.Context) (Analytics, error) {
	var a Analytics
	e := s.pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM users),(SELECT count(DISTINCT user_id) FROM workout_sessions WHERE started_at>=NOW()-INTERVAL '7 days'),(SELECT count(DISTINCT user_id) FROM workout_sessions WHERE started_at>=NOW()-INTERVAL '30 days'),(SELECT count(*) FROM workout_sessions WHERE status='completed'),(SELECT count(*) FROM workout_sessions WHERE status='completed' AND completed_at>=NOW()-INTERVAL '7 days'),(SELECT count(*) FROM workout_sessions WHERE status='completed' AND completed_at>=NOW()-INTERVAL '30 days'),(SELECT count(*) FROM user_lesson_progress WHERE completed)`).Scan(&a.TotalUsers, &a.ActiveUsers7D, &a.ActiveUsers30D, &a.TotalWorkoutsCompleted, &a.Workouts7D, &a.Workouts30D, &a.TotalLessonsCompleted)
	if e != nil {
		return a, e
	}
	a.PopularWorkouts, _ = s.metrics(ctx, `SELECT w.title,count(*)::int,COALESCE(avg(ws.duration_seconds),0)::float8 FROM workout_sessions ws JOIN workouts w ON w.id=ws.workout_id WHERE ws.status='completed' GROUP BY w.id ORDER BY count(*) DESC LIMIT 10`)
	a.SkillProgress, _ = s.metrics(ctx, `SELECT s.name,count(*)::int,COALESCE(avg(usp.current_level),0)::float8 FROM user_skill_progress usp JOIN skills s ON s.id=usp.skill_id GROUP BY s.id ORDER BY count(*) DESC LIMIT 10`)
	a.TopAchievements, _ = s.metrics(ctx, `SELECT a.title,count(*)::int,0::float8 FROM user_achievements ua JOIN achievements a ON a.id=ua.achievement_id GROUP BY a.id ORDER BY count(*) DESC LIMIT 10`)
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

var tables = map[string]struct{ table, name, slug string }{"lessons": {"lessons", "title", "slug"}, "exercises": {"exercises", "name", "slug"}, "programs": {"programs", "name", "slug"}, "workouts": {"workouts", "title", "''"}, "skills": {"skills", "name", "code"}}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

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
	x, ok := tables[kind]
	if !ok {
		return nil, ErrInvalid
	}
	q := fmt.Sprintf("SELECT to_jsonb(t) FROM %s t WHERE id=$1::uuid AND ($2 OR owner_user_id=$3::uuid)", x.table)
	var raw []byte
	if e := s.pool.QueryRow(ctx, q, id, role.CanManageAll(), user).Scan(&raw); errors.Is(e, pgx.ErrNoRows) {
		return nil, ErrForbidden
	} else if e != nil {
		return nil, e
	}
	var out map[string]any
	if e := json.Unmarshal(raw, &out); e != nil {
		return nil, e
	}
	var extra []byte
	switch kind {
	case "workouts":
		_ = s.pool.QueryRow(ctx, `SELECT COALESCE(jsonb_agg(to_jsonb(we) ORDER BY sort_order),'[]') FROM workout_exercises we WHERE workout_id=$1::uuid`, id).Scan(&extra)
		out["exercises"] = json.RawMessage(extra)
	case "programs":
		_ = s.pool.QueryRow(ctx, `SELECT COALESCE(jsonb_agg(to_jsonb(pl) ORDER BY sort_order),'[]') FROM program_levels pl WHERE program_id=$1::uuid`, id).Scan(&extra)
		out["levels"] = json.RawMessage(extra)
	case "skills":
		_ = s.pool.QueryRow(ctx, `SELECT COALESCE(jsonb_agg(to_jsonb(sl) ORDER BY sort_order),'[]') FROM skill_levels sl WHERE skill_id=$1::uuid`, id).Scan(&extra)
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
	return out, nil
}
func (s *Service) Options(ctx context.Context, user string, role Role) (Options, error) {
	var out Options
	sets := []struct {
		q   string
		dst *[]Option
	}{{`SELECT id::text,name FROM lesson_categories ORDER BY sort_order,name`, &out.Categories}, {`SELECT id::text,name FROM exercises WHERE status<>'archived' AND ($1 OR owner_user_id=$2::uuid OR owner_user_id IS NULL) ORDER BY name`, &out.Exercises}, {`SELECT id::text,name FROM programs WHERE status<>'archived' AND ($1 OR owner_user_id=$2::uuid OR owner_user_id IS NULL) ORDER BY name`, &out.Programs}, {`SELECT id::text,title FROM workouts WHERE status<>'archived' AND ($1 OR owner_user_id=$2::uuid OR owner_user_id IS NULL) ORDER BY title`, &out.Workouts}, {`SELECT id::text,name FROM skills WHERE status<>'archived' AND ($1 OR owner_user_id=$2::uuid OR owner_user_id IS NULL) ORDER BY name`, &out.Skills}, {`SELECT id::text,original_filename FROM media_assets WHERE status='ready' AND ($1 OR owner_user_id=$2::uuid) ORDER BY created_at DESC`, &out.Media}}
	for _, set := range sets {
		rows, e := s.pool.Query(ctx, set.q, role.CanManageAll(), user)
		if e != nil {
			return out, e
		}
		for rows.Next() {
			var x Option
			if e = rows.Scan(&x.ID, &x.Name); e != nil {
				rows.Close()
				return out, e
			}
			*set.dst = append(*set.dst, x)
		}
		rows.Close()
	}
	return out, nil
}

func (s *Service) List(ctx context.Context, kind, user string, role Role, search, status string) ([]Item, error) {
	x, ok := tables[kind]
	if !ok {
		return nil, ErrInvalid
	}
	where := " WHERE 1=1"
	args := []any{}
	if !role.CanManageAll() {
		args = append(args, user)
		where += fmt.Sprintf(" AND owner_user_id=$%d::uuid", len(args))
	}
	if search != "" {
		args = append(args, "%"+search+"%")
		where += fmt.Sprintf(" AND "+x.name+" ILIKE $%d", len(args))
	}
	if status != "" {
		args = append(args, status)
		where += fmt.Sprintf(" AND status=$%d", len(args))
	}
	q := fmt.Sprintf(`SELECT id::text,%s,%s,status,difficulty,owner_user_id::text,updated_at FROM %s%s ORDER BY updated_at DESC`, x.name, x.slug, x.table, where)
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
	if !validBlocks(in.Blocks) || strings.TrimSpace(in.Title) == "" || in.DurationMinutes < 1 || strings.TrimSpace(in.CategoryID) == "" {
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
		e := s.pool.QueryRow(ctx, `INSERT INTO lessons(category_id,title,slug,short_description,content,difficulty,duration_minutes,owner_user_id,status,content_blocks,cover_media_id) VALUES($1::uuid,$2,$3,$4,$5,$6,$7,$8::uuid,'draft',$9::jsonb,$10::uuid) RETURNING id::text`, in.CategoryID, in.Title, in.Slug, in.ShortDescription, in.Content, in.Difficulty, in.DurationMinutes, user, blocks, in.CoverMediaID).Scan(&out)
		return out, e
	}
	var out string
	e := s.pool.QueryRow(ctx, `UPDATE lessons SET category_id=$3::uuid,title=$4,slug=$5,short_description=$6,content=$7,difficulty=$8,duration_minutes=$9,content_blocks=$10::jsonb,cover_media_id=$11::uuid WHERE id=$1::uuid AND ($2 OR owner_user_id=$12::uuid) RETURNING id::text`, *id, role.CanManageAll(), in.CategoryID, in.Title, in.Slug, in.ShortDescription, in.Content, in.Difficulty, in.DurationMinutes, blocks, in.CoverMediaID, user).Scan(&out)
	if errors.Is(e, pgx.ErrNoRows) {
		return "", ErrForbidden
	}
	return out, e
}
func validExercises(items []BuilderExercise) bool {
	seen := map[int]bool{}
	for _, x := range items {
		if x.Sets < 1 || x.RestSeconds < 0 || x.SortOrder < 0 || seen[x.SortOrder] || (x.TargetReps == nil) == (x.TargetDurationSeconds == nil) {
			return false
		}
		seen[x.SortOrder] = true
		if x.TargetReps != nil && *x.TargetReps < 1 {
			return false
		}
		if x.TargetDurationSeconds != nil && *x.TargetDurationSeconds < 1 {
			return false
		}
	}
	return true
}
func (s *Service) SaveBuilder(ctx context.Context, kind, user string, role Role, id *string, in BuilderInput) (string, error) {
	if strings.TrimSpace(in.Description) == "" || strings.TrimSpace(in.Difficulty) == "" {
		return "", ErrInvalid
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
			e = tx.QueryRow(ctx, `INSERT INTO exercises(name,slug,description,instructions,common_mistakes,difficulty,muscle_groups,equipment,owner_user_id,status,movement_type,coach_tips,cover_media_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9::uuid,'draft',$10,$11,$12::uuid) RETURNING id::text`, in.Name, in.Slug, in.Description, in.Instructions, in.CommonMistakes, in.Difficulty, in.MuscleGroups, in.Equipment, user, in.MovementType, in.CoachTips, in.CoverMediaID).Scan(&out)
		} else {
			e = tx.QueryRow(ctx, `UPDATE exercises SET name=$4,slug=$5,description=$6,instructions=$7,common_mistakes=$8,difficulty=$9,muscle_groups=$10,equipment=$11,movement_type=$12,coach_tips=$13,cover_media_id=$14::uuid WHERE id=$1::uuid AND ($2 OR owner_user_id=$3::uuid) RETURNING id::text`, *id, role.CanManageAll(), user, in.Name, in.Slug, in.Description, in.Instructions, in.CommonMistakes, in.Difficulty, in.MuscleGroups, in.Equipment, in.MovementType, in.CoachTips, in.CoverMediaID).Scan(&out)
		}
	case "programs":
		if in.Name == "" || in.DurationWeeks < 1 {
			return "", ErrInvalid
		}
		if id == nil {
			e = tx.QueryRow(ctx, `INSERT INTO programs(name,slug,description,difficulty,duration_weeks,category,owner_user_id,status,cover_media_id) VALUES($1,$2,$3,$4,$5,$6,$7::uuid,'draft',$8::uuid) RETURNING id::text`, in.Name, in.Slug, in.Description, in.Difficulty, in.DurationWeeks, in.Category, user, in.CoverMediaID).Scan(&out)
		} else {
			e = tx.QueryRow(ctx, `UPDATE programs SET name=$4,slug=$5,description=$6,difficulty=$7,duration_weeks=$8,category=$9,cover_media_id=$10::uuid WHERE id=$1::uuid AND ($2 OR owner_user_id=$3::uuid) RETURNING id::text`, *id, role.CanManageAll(), user, in.Name, in.Slug, in.Description, in.Difficulty, in.DurationWeeks, in.Category, in.CoverMediaID).Scan(&out)
		}
		if e == nil && len(in.Levels) > 0 {
			_, e = tx.Exec(ctx, `DELETE FROM program_levels WHERE program_id=$1::uuid`, out)
			for _, l := range in.Levels {
				if e != nil {
					break
				}
				_, e = tx.Exec(ctx, `INSERT INTO program_levels(program_id,level_number,title,description,difficulty,unlock_rule_type,unlock_rule_value,sort_order) VALUES($1::uuid,$2,$3,$4,$5,$6,$7,$8)`, out, l.LevelNumber, l.Title, l.Description, l.Difficulty, l.UnlockRuleType, l.UnlockRuleValue, l.SortOrder)
			}
		}
	case "workouts":
		if in.Title == "" || in.EstimatedMinutes < 1 || in.ProgramID == "" || !validExercises(in.Exercises) {
			return "", ErrInvalid
		}
		if id == nil {
			e = tx.QueryRow(ctx, `INSERT INTO workouts(program_id,program_level_id,day_number,title,description,estimated_minutes,owner_user_id,status,cover_media_id) VALUES($1::uuid,$2::uuid,$3,$4,$5,$6,$7::uuid,'draft',$8::uuid) RETURNING id::text`, in.ProgramID, in.ProgramLevelID, in.DayNumber, in.Title, in.Description, in.EstimatedMinutes, user, in.CoverMediaID).Scan(&out)
		} else {
			e = tx.QueryRow(ctx, `UPDATE workouts SET program_id=$4::uuid,program_level_id=$5::uuid,day_number=$6,title=$7,description=$8,estimated_minutes=$9,cover_media_id=$10::uuid WHERE id=$1::uuid AND ($2 OR owner_user_id=$3::uuid) RETURNING id::text`, *id, role.CanManageAll(), user, in.ProgramID, in.ProgramLevelID, in.DayNumber, in.Title, in.Description, in.EstimatedMinutes, in.CoverMediaID).Scan(&out)
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
		if in.Name == "" || in.Icon == "" || in.FinalCriterionValue < 1 || len(in.Levels) == 0 {
			return "", ErrInvalid
		}
		if in.Code == "" {
			in.Code = strings.ToUpper(strings.ReplaceAll(slugBase(in.Name), "-", "_"))
		}
		if id == nil {
			e = tx.QueryRow(ctx, `INSERT INTO skills(code,name,description,category,difficulty,icon,xp_reward,final_criterion_type,final_criterion_value,owner_user_id,status,cover_media_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::uuid,'draft',$11::uuid) RETURNING id::text`, in.Code, in.Name, in.Description, in.Category, in.Difficulty, in.Icon, in.XPReward, in.FinalCriterionType, in.FinalCriterionValue, user, in.CoverMediaID).Scan(&out)
		} else {
			e = tx.QueryRow(ctx, `UPDATE skills SET code=$4,name=$5,description=$6,category=$7,difficulty=$8,icon=$9,xp_reward=$10,final_criterion_type=$11,final_criterion_value=$12,cover_media_id=$13::uuid WHERE id=$1::uuid AND ($2 OR owner_user_id=$3::uuid) RETURNING id::text`, *id, role.CanManageAll(), user, in.Code, in.Name, in.Description, in.Category, in.Difficulty, in.Icon, in.XPReward, in.FinalCriterionType, in.FinalCriterionValue, in.CoverMediaID).Scan(&out)
		}
		if e == nil {
			_, e = tx.Exec(ctx, `DELETE FROM skill_requirements WHERE skill_id=$1::uuid`, out)
			if e == nil {
				_, e = tx.Exec(ctx, `DELETE FROM skill_levels WHERE skill_id=$1::uuid`, out)
			}
			for _, l := range in.Levels {
				if e != nil {
					break
				}
				_, e = tx.Exec(ctx, `INSERT INTO skill_levels(skill_id,level_number,name,description,program_level_id,criterion_type,criterion_value,sort_order) VALUES($1::uuid,$2,$3,$4,$5::uuid,$6,$7,$8)`, out, l.LevelNumber, l.Title, l.Description, l.ProgramLevelID, l.CriterionType, l.CriterionValue, l.SortOrder)
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
	x, ok := tables[kind]
	if !ok || (status != "draft" && status != "published" && status != "archived") {
		return ErrInvalid
	}
	if status == "published" {
		var valid bool
		switch kind {
		case "lessons":
			_ = s.pool.QueryRow(ctx, `SELECT content<>'' OR jsonb_array_length(content_blocks)>0 FROM lessons WHERE id=$1::uuid`, id).Scan(&valid)
		case "workouts":
			_ = s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM workout_exercises WHERE workout_id=$1::uuid)`, id).Scan(&valid)
		case "programs":
			_ = s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM program_levels pl JOIN workouts w ON w.program_level_id=pl.id WHERE pl.program_id=$1::uuid)`, id).Scan(&valid)
		case "skills":
			_ = s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM skill_levels sl JOIN workouts w ON w.program_level_id=sl.program_level_id WHERE sl.skill_id=$1::uuid)`, id).Scan(&valid)
		default:
			valid = true
		}
		if !valid {
			return ErrInvalid
		}
	}
	q := fmt.Sprintf(`UPDATE %s SET status=$2,published_by=CASE WHEN $2='published' THEN $3::uuid ELSE published_by END,published_at=CASE WHEN $2='published' THEN NOW() ELSE published_at END WHERE id=$1::uuid AND ($4 OR owner_user_id=$3::uuid)`, x.table)
	tag, e := s.pool.Exec(ctx, q, id, status, user, role.CanManageAll())
	if e != nil {
		return e
	}
	if tag.RowsAffected() == 0 {
		return ErrForbidden
	}
	return nil
}
func (s *Service) Duplicate(ctx context.Context, kind, id, user string, role Role) (string, error) {
	x, ok := tables[kind]
	if !ok {
		return "", ErrInvalid
	}
	if kind == "lessons" {
		var out string
		e := s.pool.QueryRow(ctx, `INSERT INTO lessons(category_id,title,slug,short_description,content,difficulty,duration_minutes,sort_order,published,owner_user_id,status,content_blocks,cover_media_id) SELECT category_id,title||' — копия',slug||'-copy-'||substr(gen_random_uuid()::text,1,8),short_description,content,difficulty,duration_minutes,sort_order,false,$2::uuid,'draft',content_blocks,cover_media_id FROM lessons WHERE id=$1::uuid AND ($3 OR owner_user_id=$2::uuid) RETURNING id::text`, id, user, role.CanManageAll()).Scan(&out)
		if errors.Is(e, pgx.ErrNoRows) {
			return "", ErrForbidden
		}
		return out, e
	}
	if kind == "exercises" {
		var out string
		e := s.pool.QueryRow(ctx, `INSERT INTO exercises(name,slug,description,instructions,common_mistakes,difficulty,muscle_groups,equipment,video_url,image_url,owner_user_id,status,movement_type,coach_tips,cover_media_id) SELECT name||' — копия',slug||'-copy-'||substr(gen_random_uuid()::text,1,8),description,instructions,common_mistakes,difficulty,muscle_groups,equipment,video_url,image_url,$2::uuid,'draft',movement_type,coach_tips,cover_media_id FROM exercises WHERE id=$1::uuid AND ($3 OR owner_user_id=$2::uuid) RETURNING id::text`, id, user, role.CanManageAll()).Scan(&out)
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
		e = tx.QueryRow(ctx, `INSERT INTO workouts(program_id,program_level_id,day_number,title,description,estimated_minutes,sort_order,owner_user_id,status,cover_media_id) SELECT program_id,program_level_id,(SELECT COALESCE(max(day_number),0)+1 FROM workouts w2 WHERE w2.program_id=w.program_id),title||' — копия',description,estimated_minutes,sort_order,$2::uuid,'draft',cover_media_id FROM workouts w WHERE id=$1::uuid AND ($3 OR owner_user_id=$2::uuid) RETURNING id::text`, id, user, role.CanManageAll()).Scan(&out)
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
		e = tx.QueryRow(ctx, `INSERT INTO programs(name,slug,description,difficulty,duration_weeks,published,category,owner_user_id,status,cover_media_id,coach_description) SELECT name||' — копия',slug||'-copy-'||substr(gen_random_uuid()::text,1,8),description,difficulty,duration_weeks,false,category,$2::uuid,'draft',cover_media_id,coach_description FROM programs WHERE id=$1::uuid AND ($3 OR owner_user_id=$2::uuid) RETURNING id::text`, id, user, role.CanManageAll()).Scan(&out)
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
		e = tx.QueryRow(ctx, `INSERT INTO skills(code,name,description,category,difficulty,icon,xp_reward,final_criterion_type,final_criterion_value,owner_user_id,status,cover_media_id) SELECT code||'_COPY_'||upper(substr(replace(gen_random_uuid()::text,'-',''),1,6)),name||' — копия',description,category,difficulty,icon,xp_reward,final_criterion_type,final_criterion_value,$2::uuid,'draft',cover_media_id FROM skills WHERE id=$1::uuid AND ($3 OR owner_user_id=$2::uuid) RETURNING id::text`, id, user, role.CanManageAll()).Scan(&out)
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
	rows, e := s.pool.Query(ctx, `SELECT m.id::text,m.owner_user_id::text,m.type,m.status,m.storage_provider,m.storage_key,m.url,m.thumbnail_url,m.original_filename,m.mime_type,m.size_bytes,m.created_at,(SELECT count(*) FROM lessons WHERE cover_media_id=m.id OR content_blocks @> jsonb_build_array(jsonb_build_object('media_id',m.id::text)))+(SELECT count(*) FROM exercises WHERE cover_media_id=m.id)+(SELECT count(*) FROM workouts WHERE cover_media_id=m.id)+(SELECT count(*) FROM programs WHERE cover_media_id=m.id)+(SELECT count(*) FROM skills WHERE cover_media_id=m.id) FROM media_assets m WHERE $2 OR m.owner_user_id=$1::uuid ORDER BY m.created_at DESC`, user, role.CanManageAll())
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
	if (in.Type != "image" && in.Type != "video") || !strings.HasPrefix(in.URL, "https://") || !validMime(in.Type, in.MIMEType) || in.SizeBytes < 0 {
		return Media{}, ErrInvalid
	}
	key := fmt.Sprintf("external/%s/%d", user, time.Now().UnixNano())
	var id string
	e := s.pool.QueryRow(ctx, `INSERT INTO media_assets(owner_user_id,type,storage_provider,storage_key,url,thumbnail_url,original_filename,mime_type,size_bytes,width,height,duration_seconds) VALUES($1::uuid,$2,'external',$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id::text`, user, in.Type, key, in.URL, in.ThumbnailURL, safeName(in.OriginalFilename), in.MIMEType, in.SizeBytes, in.Width, in.Height, in.DurationSeconds).Scan(&id)
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
		return mime == "image/jpeg" || mime == "image/png" || mime == "image/webp"
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
	var refs int
	e := s.pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM lessons WHERE cover_media_id=$1::uuid OR content_blocks @> jsonb_build_array(jsonb_build_object('media_id',$1::text)))+(SELECT count(*) FROM exercises WHERE cover_media_id=$1::uuid)+(SELECT count(*) FROM workouts WHERE cover_media_id=$1::uuid)+(SELECT count(*) FROM programs WHERE cover_media_id=$1::uuid)+(SELECT count(*) FROM skills WHERE cover_media_id=$1::uuid)`, id).Scan(&refs)
	if e != nil {
		return e
	}
	if refs > 0 {
		return ErrInUse
	}
	tag, e := s.pool.Exec(ctx, `DELETE FROM media_assets WHERE id=$1::uuid AND ($2 OR owner_user_id=$3::uuid)`, id, role.CanManageAll(), user)
	if e != nil {
		return e
	}
	if tag.RowsAffected() == 0 {
		return ErrForbidden
	}
	return nil
}
