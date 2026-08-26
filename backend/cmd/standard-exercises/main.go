package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type catalog struct {
	Version   int        `json:"version"`
	Exercises []exercise `json:"exercises"`
}
type exercise struct {
	Key            string   `json:"key"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Instructions   string   `json:"instructions"`
	CommonMistakes string   `json:"common_mistakes"`
	CoachTips      string   `json:"coach_tips"`
	Difficulty     string   `json:"difficulty"`
	MovementType   string   `json:"movement_type"`
	MuscleGroups   []string `json:"muscle_groups"`
	Equipment      []string `json:"equipment"`
	Tags           []string `json:"tags"`
	Media          *struct {
		ImageURL string `json:"image_url"`
		VideoURL string `json:"video_url"`
	} `json:"media"`
}
type dbExercise struct {
	ID, Key, Slug, Owner, Name, Description, Instructions, CommonMistakes, CoachTips, Difficulty, MovementType, ImageURL, VideoURL, Status string
	MuscleGroups, Equipment, Tags                                                                                                          []string
}
type action struct {
	Kind     string
	Item     exercise
	Existing *dbExercise
}
type plan struct {
	Created, Updated, Unchanged, Conflicts int
	Actions                                []action
	ConflictMessages                       []string
}

var keyPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var vocab = map[string]map[string]bool{
	"difficulty":    values("beginner", "intermediate", "advanced"),
	"movement_type": values("reps", "duration", "distance", "custom"),
	"muscle_groups": values("chest", "back", "shoulders", "biceps", "triceps", "forearms", "core", "glutes", "quadriceps", "hamstrings", "calves", "full-body"),
	"equipment":     values("none", "floor", "wall", "pull-up-bar", "parallel-bars", "rings", "resistance-band", "bench", "box"),
	"tags":          values("push", "pull", "squat", "hinge", "core", "handstand", "planche", "front-lever", "back-lever", "muscle-up", "l-sit", "mobility", "warm-up"),
}

func values(v ...string) map[string]bool {
	out := map[string]bool{}
	for _, x := range v {
		out[x] = true
	}
	return out
}
func duplicates(v []string) bool {
	seen := map[string]bool{}
	for _, x := range v {
		if seen[x] {
			return true
		}
		seen[x] = true
	}
	return false
}
func invalid(v []string, allowed map[string]bool) []string {
	var out []string
	for _, x := range v {
		if !allowed[x] {
			out = append(out, x)
		}
	}
	return out
}
func validHTTPS(raw string) bool {
	if raw == "" {
		return true
	}
	u, e := url.Parse(raw)
	return e == nil && u.Scheme == "https" && u.Host != ""
}
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func loadCatalog(path string) (catalog, error) {
	raw, e := os.ReadFile(path)
	if e != nil {
		return catalog{}, e
	}
	var c catalog
	d := json.NewDecoder(strings.NewReader(string(raw)))
	d.DisallowUnknownFields()
	if e = d.Decode(&c); e != nil {
		return catalog{}, e
	}
	if e = d.Decode(&struct{}{}); !errors.Is(e, io.EOF) {
		if e == nil {
			return catalog{}, errors.New("catalog contains multiple JSON values")
		}
		return catalog{}, e
	}
	return c, nil
}
func validateCatalog(c catalog) []string {
	var errs []string
	if c.Version != 1 {
		errs = append(errs, "version must be 1")
	}
	if len(c.Exercises) == 0 {
		errs = append(errs, "catalog must contain at least one exercise")
	}
	keys := map[string]bool{}
	for i, x := range c.Exercises {
		p := fmt.Sprintf("exercises[%d]", i)
		if !keyPattern.MatchString(x.Key) {
			errs = append(errs, p+".key is invalid")
		} else if keys[x.Key] {
			errs = append(errs, p+".key is duplicated: "+x.Key)
		}
		keys[x.Key] = true
		if strings.TrimSpace(x.Name) == "" || strings.TrimSpace(x.Description) == "" || strings.TrimSpace(x.Instructions) == "" {
			errs = append(errs, p+" requires name, description and instructions")
		}
		if !vocab["difficulty"][x.Difficulty] {
			errs = append(errs, p+".difficulty is invalid: "+x.Difficulty)
		}
		if !vocab["movement_type"][x.MovementType] {
			errs = append(errs, p+".movement_type is invalid: "+x.MovementType)
		}
		if len(x.MuscleGroups) == 0 || duplicates(x.MuscleGroups) {
			errs = append(errs, p+".muscle_groups must be non-empty and unique")
		}
		if bad := invalid(x.MuscleGroups, vocab["muscle_groups"]); len(bad) > 0 {
			errs = append(errs, p+".muscle_groups has invalid values: "+strings.Join(bad, ","))
		}
		if duplicates(x.Equipment) {
			errs = append(errs, p+".equipment must be unique")
		}
		if bad := invalid(x.Equipment, vocab["equipment"]); len(bad) > 0 {
			errs = append(errs, p+".equipment has invalid values: "+strings.Join(bad, ","))
		}
		if len(x.Tags) == 0 || duplicates(x.Tags) {
			errs = append(errs, p+".tags must be non-empty and unique")
		}
		for _, tag := range x.Tags {
			if !keyPattern.MatchString(tag) {
				errs = append(errs, p+".tags has invalid value: "+tag)
			}
		}
		if x.Media != nil && (!validHTTPS(x.Media.ImageURL) || !validHTTPS(x.Media.VideoURL)) {
			errs = append(errs, p+".media URLs must use https")
		}
	}
	return errs
}
func same(x exercise, d dbExercise) bool {
	image, video := "", ""
	if x.Media != nil {
		image, video = x.Media.ImageURL, x.Media.VideoURL
	}
	return d.Key == x.Key && d.Slug == x.Key && d.Owner == "" && d.Name == x.Name && d.Description == x.Description && d.Instructions == x.Instructions && d.CommonMistakes == x.CommonMistakes && d.CoachTips == x.CoachTips && d.Difficulty == x.Difficulty && d.MovementType == x.MovementType && d.Status == "published" && d.ImageURL == image && d.VideoURL == video && slicesEqual(d.MuscleGroups, x.MuscleGroups) && slicesEqual(d.Equipment, x.Equipment) && slicesEqual(d.Tags, x.Tags)
}
func buildPlan(c catalog, rows []dbExercise) plan {
	byKey, bySlug := map[string]*dbExercise{}, map[string]*dbExercise{}
	for i := range rows {
		r := &rows[i]
		if r.Key != "" {
			byKey[r.Key] = r
		}
		bySlug[r.Slug] = r
	}
	p := plan{}
	for _, x := range c.Exercises {
		r := byKey[x.Key]
		if r == nil {
			r = bySlug[x.Key]
		}
		if r == nil {
			p.Created++
			p.Actions = append(p.Actions, action{"create", x, nil})
			continue
		}
		if r.Owner != "" {
			p.Conflicts++
			p.ConflictMessages = append(p.ConflictMessages, fmt.Sprintf("%s: slug belongs to coach-owned exercise %s", x.Key, r.ID))
			continue
		}
		if r.Key != "" && r.Key != x.Key {
			p.Conflicts++
			p.ConflictMessages = append(p.ConflictMessages, fmt.Sprintf("%s: slug belongs to standard key %s", x.Key, r.Key))
			continue
		}
		if same(x, *r) {
			p.Unchanged++
			continue
		}
		p.Updated++
		p.Actions = append(p.Actions, action{"update", x, r})
	}
	return p
}
func readRows(ctx context.Context, pool *pgxpool.Pool, c catalog) ([]dbExercise, error) {
	keys := make([]string, len(c.Exercises))
	for i, x := range c.Exercises {
		keys[i] = x.Key
	}
	rows, e := pool.Query(ctx, `SELECT id::text,COALESCE(standard_key,''),slug,COALESCE(owner_user_id::text,''),name,description,instructions,common_mistakes,coach_tips,difficulty,movement_type,COALESCE(image_url,''),COALESCE(video_url,''),status,muscle_groups,equipment,tags FROM exercises WHERE standard_key IS NOT NULL OR slug=ANY($1)`, keys)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []dbExercise
	for rows.Next() {
		var x dbExercise
		if e = rows.Scan(&x.ID, &x.Key, &x.Slug, &x.Owner, &x.Name, &x.Description, &x.Instructions, &x.CommonMistakes, &x.CoachTips, &x.Difficulty, &x.MovementType, &x.ImageURL, &x.VideoURL, &x.Status, &x.MuscleGroups, &x.Equipment, &x.Tags); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func applyPlan(ctx context.Context, pool *pgxpool.Pool, p plan) error {
	if p.Conflicts > 0 {
		return errors.New("import blocked by conflicts")
	}
	tx, e := pool.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	for _, a := range p.Actions {
		image, video := "", ""
		if a.Item.Media != nil {
			image, video = a.Item.Media.ImageURL, a.Item.Media.VideoURL
		}
		if a.Kind == "create" {
			_, e = tx.Exec(ctx, `INSERT INTO exercises(standard_key,name,slug,description,instructions,common_mistakes,difficulty,muscle_groups,equipment,tags,video_url,image_url,owner_user_id,status,movement_type,coach_tips) VALUES($1,$2,$1,$3,$4,$5,$6,$7,$8,$9,NULLIF($10,''),NULLIF($11,''),NULL,'published',$12,$13)`, a.Item.Key, a.Item.Name, a.Item.Description, a.Item.Instructions, a.Item.CommonMistakes, a.Item.Difficulty, a.Item.MuscleGroups, a.Item.Equipment, a.Item.Tags, video, image, a.Item.MovementType, a.Item.CoachTips)
		} else {
			_, e = tx.Exec(ctx, `UPDATE exercises SET standard_key=$2,name=$3,slug=$2,description=$4,instructions=$5,common_mistakes=$6,difficulty=$7,muscle_groups=$8,equipment=$9,tags=$10,video_url=NULLIF($11,''),image_url=NULLIF($12,''),status='published',movement_type=$13,coach_tips=$14 WHERE id=$1::uuid AND owner_user_id IS NULL`, a.Existing.ID, a.Item.Key, a.Item.Name, a.Item.Description, a.Item.Instructions, a.Item.CommonMistakes, a.Item.Difficulty, a.Item.MuscleGroups, a.Item.Equipment, a.Item.Tags, video, image, a.Item.MovementType, a.Item.CoachTips)
		}
		if e != nil {
			return e
		}
	}
	return tx.Commit(ctx)
}
func printPlan(p plan) {
	fmt.Printf("PLAN: created=%d updated=%d unchanged=%d conflicts=%d\n", p.Created, p.Updated, p.Unchanged, p.Conflicts)
	sort.Strings(p.ConflictMessages)
	for _, m := range p.ConflictMessages {
		fmt.Println("CONFLICT:", m)
	}
}
func run() error {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	file := fs.String("file", "", "catalog JSON file")
	validate := fs.Bool("validate-only", false, "validate without database access")
	dry := fs.Bool("dry-run", false, "show import plan without writing")
	confirm := fs.Bool("confirm", false, "execute import")
	dbURL := fs.String("database-url", os.Getenv("DATABASE_URL"), "PostgreSQL connection URL")
	if e := fs.Parse(os.Args[1:]); e != nil {
		return e
	}
	if *file == "" {
		return errors.New("--file is required")
	}
	modes := 0
	for _, x := range []bool{*validate, *dry, *confirm} {
		if x {
			modes++
		}
	}
	if modes != 1 {
		return errors.New("choose exactly one of --validate-only, --dry-run or --confirm; no database writes were made")
	}
	c, e := loadCatalog(*file)
	if e != nil {
		return e
	}
	if errs := validateCatalog(c); len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	fmt.Printf("VALID: %d standard exercises\n", len(c.Exercises))
	if *validate {
		fmt.Println("No database changes applied.")
		return nil
	}
	if *dbURL == "" {
		return errors.New("DATABASE_URL or --database-url is required")
	}
	ctx := context.Background()
	pool, e := pgxpool.New(ctx, *dbURL)
	if e != nil {
		return e
	}
	defer pool.Close()
	if e = pool.Ping(ctx); e != nil {
		return e
	}
	rows, e := readRows(ctx, pool, c)
	if e != nil {
		return fmt.Errorf("read exercises (is migration 000013 applied?): %w", e)
	}
	p := buildPlan(c, rows)
	printPlan(p)
	if *dry {
		fmt.Println("DRY RUN: no database changes applied.")
		return nil
	}
	if p.Conflicts > 0 {
		return errors.New("confirmed import refused because conflicts exist")
	}
	if e = applyPlan(ctx, pool, p); e != nil {
		return e
	}
	fmt.Printf("IMPORTED: created=%d updated=%d unchanged=%d\n", p.Created, p.Updated, p.Unchanged)
	return nil
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
