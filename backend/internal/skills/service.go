package skills

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound  = errors.New("skill not found")
	ErrLocked    = errors.New("skill or level is locked")
	ErrCriterion = errors.New("skill criterion is not met")
)

type Skill struct {
	ID                  string `json:"id"`
	Code                string `json:"code"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	Category            string `json:"category"`
	Difficulty          string `json:"difficulty"`
	Icon                string `json:"icon"`
	XPReward            int32  `json:"xp_reward"`
	FinalCriterionType  string `json:"final_criterion_type"`
	FinalCriterionValue int32  `json:"final_criterion_value"`
	Status              string `json:"status"`
	CurrentLevel        int32  `json:"current_level"`
	TotalLevels         int32  `json:"total_levels"`
	ProgressPercent     int32  `json:"progress_percent"`
	CoverMediaURL       string `json:"cover_media_url,omitempty"`
}
type Workout struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	EstimatedMinutes int32  `json:"estimated_minutes"`
}
type Level struct {
	ID             string    `json:"id"`
	LevelNumber    int32     `json:"level_number"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	CriterionType  string    `json:"criterion_type"`
	CriterionValue int32     `json:"criterion_value"`
	Status         string    `json:"status"`
	ProgressValue  int32     `json:"progress_value"`
	Workouts       []Workout `json:"workouts"`
}
type Detail struct {
	Skill    Skill       `json:"skill"`
	Levels   []Level     `json:"levels"`
	Criteria []Criterion `json:"criteria"`
}
type Criterion struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Title       string `json:"title"`
	Type        string `json:"criterion_type"`
	TargetValue int32  `json:"target_value"`
	Confirmed   bool   `json:"confirmed"`
}
type Requirement struct {
	SkillID         string `json:"skill_id"`
	RequiredSkillID string `json:"required_skill_id"`
	Type            string `json:"requirement_type"`
	Value           int32  `json:"requirement_value"`
}
type Map struct {
	Nodes        []Skill       `json:"nodes"`
	Requirements []Requirement `json:"requirements"`
}
type Mastery struct {
	SkillID         string `json:"skill_id"`
	Status          string `json:"status"`
	XPEarned        int32  `json:"xp_earned"`
	Achievement     string `json:"achievement,omitempty"`
	AlreadyMastered bool   `json:"already_mastered"`
}

type Service struct{ pool *pgxpool.Pool }

func NewService(pool *pgxpool.Pool) *Service { return &Service{pool} }

const listSQL = `SELECT s.id::text,s.code,s.name,s.description,s.category,s.difficulty,s.icon,s.xp_reward,s.final_criterion_type,s.final_criterion_value,COALESCE(m.url,''),
CASE WHEN usp.status IS NOT NULL THEN usp.status WHEN EXISTS(SELECT 1 FROM skill_requirements r LEFT JOIN user_skill_progress required ON required.user_id=$1::uuid AND required.skill_id=r.required_skill_id WHERE r.skill_id=s.id AND (required.status IS DISTINCT FROM 'mastered' OR (r.requirement_type='skill_level' AND required.current_level<r.requirement_value))) THEN 'locked' ELSE 'available' END,
COALESCE(usp.current_level,1),GREATEST(COUNT(sl.id),(SELECT count(*) FROM skill_criteria c WHERE c.skill_id=s.id))::int,GREATEST(COALESCE(COUNT(ul.skill_level_id) FILTER(WHERE ul.status='completed'),0),(SELECT count(*) FROM user_skill_criteria uc JOIN skill_criteria c ON c.id=uc.criterion_id WHERE c.skill_id=s.id AND uc.user_id=$1::uuid))::int
FROM skills s LEFT JOIN media_assets m ON m.id=s.cover_media_id LEFT JOIN user_skill_progress usp ON usp.skill_id=s.id AND usp.user_id=$1::uuid LEFT JOIN skill_levels sl ON sl.skill_id=s.id LEFT JOIN user_skill_level_progress ul ON ul.skill_level_id=sl.id AND ul.user_id=$1::uuid WHERE s.status='published' AND NOT s.hidden GROUP BY s.id,m.url,usp.status,usp.current_level ORDER BY CASE WHEN s.sort_order=0 THEN 2147483647 ELSE s.sort_order END,s.name,s.id`

func scanSkill(rows pgx.Rows) (Skill, error) {
	var x Skill
	var completed int32
	err := rows.Scan(&x.ID, &x.Code, &x.Name, &x.Description, &x.Category, &x.Difficulty, &x.Icon, &x.XPReward, &x.FinalCriterionType, &x.FinalCriterionValue, &x.CoverMediaURL, &x.Status, &x.CurrentLevel, &x.TotalLevels, &completed)
	if x.TotalLevels > 0 {
		x.ProgressPercent = completed * 100 / x.TotalLevels
	}
	if x.Status == "mastered" {
		x.ProgressPercent = 100
	}
	return x, err
}
func (s *Service) List(ctx context.Context, user string) ([]Skill, error) {
	rows, e := s.pool.Query(ctx, listSQL, user)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Skill{}
	for rows.Next() {
		x, e := scanSkill(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (s *Service) Map(ctx context.Context, user string) (Map, error) {
	nodes, e := s.List(ctx, user)
	if e != nil {
		return Map{}, e
	}
	rows, e := s.pool.Query(ctx, `SELECT skill_id::text,required_skill_id::text,requirement_type,requirement_value FROM skill_requirements ORDER BY skill_id,required_skill_id`)
	if e != nil {
		return Map{}, e
	}
	defer rows.Close()
	reqs := []Requirement{}
	for rows.Next() {
		var r Requirement
		if e = rows.Scan(&r.SkillID, &r.RequiredSkillID, &r.Type, &r.Value); e != nil {
			return Map{}, e
		}
		reqs = append(reqs, r)
	}
	return Map{nodes, reqs}, rows.Err()
}
func (s *Service) Get(ctx context.Context, user, id string) (Detail, error) {
	items, e := s.List(ctx, user)
	if e != nil {
		return Detail{}, e
	}
	var skill *Skill
	for i := range items {
		if items[i].ID == id {
			skill = &items[i]
			break
		}
	}
	if skill == nil {
		return Detail{}, ErrNotFound
	}
	rows, e := s.pool.Query(ctx, `SELECT sl.id::text,sl.level_number,sl.name,sl.description,sl.criterion_type,sl.criterion_value,COALESCE(ulp.status,''),COALESCE(ulp.progress_value,0),w.id::text,w.title,w.estimated_minutes FROM skill_levels sl LEFT JOIN user_skill_level_progress ulp ON ulp.skill_level_id=sl.id AND ulp.user_id=$2::uuid LEFT JOIN workouts w ON w.program_level_id=sl.program_level_id WHERE sl.skill_id=$1::uuid ORDER BY sl.level_number,w.sort_order`, id, user)
	if e != nil {
		return Detail{}, e
	}
	defer rows.Close()
	levels := []Level{}
	index := map[string]int{}
	for rows.Next() {
		var l Level
		var wid, title *string
		var minutes *int32
		if e = rows.Scan(&l.ID, &l.LevelNumber, &l.Name, &l.Description, &l.CriterionType, &l.CriterionValue, &l.Status, &l.ProgressValue, &wid, &title, &minutes); e != nil {
			return Detail{}, e
		}
		i, ok := index[l.ID]
		if !ok {
			if l.Status == "" {
				if skill.Status == "locked" || l.LevelNumber > skill.CurrentLevel {
					l.Status = "locked"
				} else {
					l.Status = "available"
				}
			}
			l.Workouts = []Workout{}
			levels = append(levels, l)
			i = len(levels) - 1
			index[l.ID] = i
		}
		if wid != nil {
			levels[i].Workouts = append(levels[i].Workouts, Workout{*wid, *title, *minutes})
		}
	}
	if e = rows.Err(); e != nil {
		return Detail{}, e
	}
	criteriaRows, e := s.pool.Query(ctx, `SELECT c.id::text,c.code,c.title,c.criterion_type,c.target_value,uc.criterion_id IS NOT NULL FROM skill_criteria c LEFT JOIN user_skill_criteria uc ON uc.criterion_id=c.id AND uc.user_id=$2::uuid WHERE c.skill_id=$1::uuid ORDER BY c.sort_order`, id, user)
	if e != nil {
		return Detail{}, e
	}
	defer criteriaRows.Close()
	criteria := []Criterion{}
	for criteriaRows.Next() {
		var c Criterion
		if e = criteriaRows.Scan(&c.ID, &c.Code, &c.Title, &c.Type, &c.TargetValue, &c.Confirmed); e != nil {
			return Detail{}, e
		}
		criteria = append(criteria, c)
	}
	return Detail{*skill, levels, criteria}, criteriaRows.Err()
}

func (s *Service) ConfirmCriterion(ctx context.Context, user, skillID, criterionID string) (Mastery, error) {
	tx, e := s.pool.Begin(ctx)
	if e != nil {
		return Mastery{}, e
	}
	defer tx.Rollback(ctx)
	var code string
	var reward int32
	e = tx.QueryRow(ctx, `SELECT s.code,s.xp_reward FROM skill_criteria c JOIN skills s ON s.id=c.skill_id WHERE c.id=$1::uuid AND c.skill_id=$2::uuid AND s.status='published'`, criterionID, skillID).Scan(&code, &reward)
	if errors.Is(e, pgx.ErrNoRows) {
		return Mastery{}, ErrNotFound
	}
	if e != nil {
		return Mastery{}, e
	}
	_, e = tx.Exec(ctx, `INSERT INTO user_skill_criteria(user_id,criterion_id) VALUES($1::uuid,$2::uuid) ON CONFLICT DO NOTHING`, user, criterionID)
	if e != nil {
		return Mastery{}, e
	}
	var total, done int32
	e = tx.QueryRow(ctx, `SELECT count(*)::int,count(uc.criterion_id)::int FROM skill_criteria c LEFT JOIN user_skill_criteria uc ON uc.criterion_id=c.id AND uc.user_id=$2::uuid WHERE c.skill_id=$1::uuid`, skillID, user).Scan(&total, &done)
	if e != nil {
		return Mastery{}, e
	}
	if total == 0 || done < total {
		if e = tx.Commit(ctx); e != nil {
			return Mastery{}, e
		}
		return Mastery{SkillID: skillID, Status: "in_progress"}, nil
	}
	var prior string
	e = tx.QueryRow(ctx, `INSERT INTO user_skill_progress(user_id,skill_id,current_level,status,started_at,completed_at) VALUES($1::uuid,$2::uuid,1,'mastered',NOW(),NOW()) ON CONFLICT(user_id,skill_id) DO UPDATE SET status='mastered',completed_at=COALESCE(user_skill_progress.completed_at,NOW()),updated_at=NOW() RETURNING status`, user, skillID).Scan(&prior)
	if e != nil {
		return Mastery{}, e
	}
	var awarded bool
	e = tx.QueryRow(ctx, `INSERT INTO user_achievements(user_id,achievement_id) SELECT $1::uuid,id FROM achievements WHERE code='CALISTHENICS_BASE_READY' ON CONFLICT DO NOTHING RETURNING true`, user).Scan(&awarded)
	if errors.Is(e, pgx.ErrNoRows) {
		e = nil
	}
	if e != nil {
		return Mastery{}, e
	}
	if awarded {
		_, e = tx.Exec(ctx, `UPDATE profiles SET xp=xp+$2,level=(SELECT level FROM levels WHERE min_xp<=xp+$2 ORDER BY min_xp DESC LIMIT 1) WHERE user_id=$1::uuid`, user, reward)
		if e != nil {
			return Mastery{}, e
		}
	}
	if e = tx.Commit(ctx); e != nil {
		return Mastery{}, e
	}
	return Mastery{SkillID: skillID, Status: "mastered", XPEarned: map[bool]int32{true: reward}[awarded], Achievement: map[bool]string{true: "CALISTHENICS_BASE_READY"}[awarded], AlreadyMastered: !awarded}, nil
}
func (s *Service) CompleteLevel(ctx context.Context, user, skillID string, levelNumber, value int32) error {
	tx, e := s.pool.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	var levelID, criterion string
	var required int32
	e = tx.QueryRow(ctx, `SELECT id::text,criterion_type,criterion_value FROM skill_levels WHERE skill_id=$1::uuid AND level_number=$2`, skillID, levelNumber).Scan(&levelID, &criterion, &required)
	if errors.Is(e, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if e != nil {
		return e
	}
	var unmet bool
	e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM skill_requirements r LEFT JOIN user_skill_progress p ON p.user_id=$1::uuid AND p.skill_id=r.required_skill_id WHERE r.skill_id=$2::uuid AND p.status IS DISTINCT FROM 'mastered') OR EXISTS(SELECT 1 FROM skill_levels prior LEFT JOIN user_skill_level_progress up ON up.user_id=$1::uuid AND up.skill_level_id=prior.id WHERE prior.skill_id=$2::uuid AND prior.level_number<$3 AND up.status IS DISTINCT FROM 'completed')`, user, skillID, levelNumber).Scan(&unmet)
	if e != nil {
		return e
	}
	if unmet {
		return ErrLocked
	}
	if criterion == "workout_completed" {
		e = tx.QueryRow(ctx, `SELECT COUNT(*)::int FROM workout_sessions ws JOIN workouts w ON w.id=ws.workout_id JOIN skill_levels sl ON sl.program_level_id=w.program_level_id WHERE ws.user_id=$1::uuid AND ws.status='completed' AND sl.id=$2::uuid`, user, levelID).Scan(&value)
		if e != nil {
			return e
		}
	}
	if value < required {
		return ErrCriterion
	}
	_, e = tx.Exec(ctx, `INSERT INTO user_skill_level_progress(user_id,skill_level_id,status,progress_value,completed_at) VALUES($1::uuid,$2::uuid,'completed',$3,NOW()) ON CONFLICT(user_id,skill_level_id) DO UPDATE SET status='completed',progress_value=GREATEST(user_skill_level_progress.progress_value,EXCLUDED.progress_value),completed_at=COALESCE(user_skill_level_progress.completed_at,NOW()),updated_at=NOW()`, user, levelID, value)
	if e != nil {
		return e
	}
	_, e = tx.Exec(ctx, `INSERT INTO user_skill_progress(user_id,skill_id,current_level,status,started_at) VALUES($1::uuid,$2::uuid,$3,'in_progress',NOW()) ON CONFLICT(user_id,skill_id) DO UPDATE SET current_level=GREATEST(user_skill_progress.current_level,$3),status=CASE WHEN user_skill_progress.status='mastered' THEN 'mastered' ELSE 'in_progress' END,updated_at=NOW()`, user, skillID, levelNumber+1)
	if e != nil {
		return e
	}
	return tx.Commit(ctx)
}
func (s *Service) Master(ctx context.Context, user, skillID string, value int32) (Mastery, error) {
	tx, e := s.pool.Begin(ctx)
	if e != nil {
		return Mastery{}, e
	}
	defer tx.Rollback(ctx)
	var code, criterion string
	var required, reward, total, done int32
	e = tx.QueryRow(ctx, `SELECT code,final_criterion_type,final_criterion_value,xp_reward,(SELECT COUNT(*)::int FROM skill_levels WHERE skill_id=s.id),(SELECT COUNT(*)::int FROM user_skill_level_progress up JOIN skill_levels sl ON sl.id=up.skill_level_id WHERE up.user_id=$1::uuid AND sl.skill_id=s.id AND up.status='completed') FROM skills s WHERE id=$2::uuid`, user, skillID).Scan(&code, &criterion, &required, &reward, &total, &done)
	if errors.Is(e, pgx.ErrNoRows) {
		return Mastery{}, ErrNotFound
	}
	if e != nil {
		return Mastery{}, e
	}
	_ = criterion
	if done < total || value < required {
		return Mastery{}, ErrCriterion
	}
	var status string
	e = tx.QueryRow(ctx, `INSERT INTO user_skill_progress(user_id,skill_id,current_level,status,started_at) VALUES($1::uuid,$2::uuid,$3,'in_progress',NOW()) ON CONFLICT(user_id,skill_id) DO UPDATE SET updated_at=NOW() RETURNING status`, user, skillID, total).Scan(&status)
	if e != nil {
		return Mastery{}, e
	}
	if status == "mastered" {
		return Mastery{SkillID: skillID, Status: status, AlreadyMastered: true}, tx.Commit(ctx)
	}
	_, e = tx.Exec(ctx, `UPDATE user_skill_progress SET status='mastered',current_level=$3,completed_at=NOW(),updated_at=NOW() WHERE user_id=$1::uuid AND skill_id=$2::uuid`, user, skillID, total)
	if e != nil {
		return Mastery{}, e
	}
	_, e = tx.Exec(ctx, `UPDATE profiles SET xp=xp+$2,level=(SELECT level FROM levels WHERE min_xp<=xp+$2 ORDER BY min_xp DESC LIMIT 1) WHERE user_id=$1::uuid`, user, reward)
	if e != nil {
		return Mastery{}, e
	}
	achievement := code + "_MASTERED"
	tag, e := tx.Exec(ctx, `INSERT INTO user_achievements(user_id,achievement_id) SELECT $1::uuid,id FROM achievements WHERE code=$2 ON CONFLICT DO NOTHING`, user, achievement)
	if e != nil {
		return Mastery{}, e
	}
	if e = tx.Commit(ctx); e != nil {
		return Mastery{}, e
	}
	out := Mastery{SkillID: skillID, Status: "mastered", XPEarned: reward}
	if tag.RowsAffected() > 0 {
		out.Achievement = achievement
	}
	return out, nil
}
