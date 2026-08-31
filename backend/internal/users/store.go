package users

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/auth"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID                  string   `json:"id"`
	TelegramID          int64    `json:"telegram_id"`
	Username            string   `json:"username,omitempty"`
	FirstName           string   `json:"first_name"`
	LastName            string   `json:"last_name,omitempty"`
	PhotoURL            string   `json:"photo_url,omitempty"`
	DisplayName         string   `json:"display_name"`
	Level               int32    `json:"level"`
	XP                  int32    `json:"xp"`
	CurrentStreak       int32    `json:"current_streak"`
	LongestStreak       int32    `json:"longest_streak"`
	PreferredDifficulty string   `json:"preferred_difficulty"`
	Timezone            string   `json:"timezone"`
	Role                string   `json:"role"`
	AvailableModes      []string `json:"available_modes"`
	CurrentTenant       *Tenant  `json:"current_tenant,omitempty"`
	Tenants             []Tenant `json:"tenants"`
}

type Tenant struct {
	ID                string `json:"id"`
	Slug              string `json:"slug"`
	Name              string `json:"name"`
	Role              string `json:"role"`
	Description       string `json:"description,omitempty"`
	Status            string `json:"status,omitempty"`
	OwnerUserID       string `json:"owner_user_id,omitempty"`
	OwnerName         string `json:"owner_name,omitempty"`
	CreatedAt         string `json:"created_at,omitempty"`
	Students          int    `json:"students,omitempty"`
	AvatarMediaID     string `json:"avatar_media_id,omitempty"`
	AvatarURL         string `json:"avatar_url,omitempty"`
	TelegramAvatarURL string `json:"telegram_avatar_url,omitempty"`
	AvatarInitials    string `json:"avatar_initials"`
}

func finishTenant(t *Tenant) {
	words := strings.Fields(t.Name)
	for i, word := range words {
		if i == 2 {
			break
		}
		r := []rune(word)
		if len(r) > 0 {
			t.AvatarInitials += strings.ToUpper(string(r[0]))
		}
	}
}

func AvailableModes(role string) []string {
	if role == "super_admin" {
		return []string{"student", "coach", "admin"}
	}
	if role == "coach" || role == "admin" { // coach retained for pre-v16 rolling compatibility
		return []string{"student", "coach"}
	}
	return []string{"student"}
}

type Store struct{ pool *pgxpool.Pool }

var tenantSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var reservedTenantSlugs = map[string]bool{"admin": true, "api": true, "app": true, "coach": true, "root": true, "start": true, "support": true, "system": true}

func validateTenantSlug(slug string) error {
	if len(slug) < 2 || len(slug) > 63 || !tenantSlugPattern.MatchString(slug) {
		return fmt.Errorf("invalid tenant slug")
	}
	if reservedTenantSlugs[slug] {
		return fmt.Errorf("reserved tenant slug")
	}
	return nil
}

func normalizeTenantSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(value, "@")))
	value = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(value, "-")
	return strings.Trim(value, "-")
}

type RoleUser struct {
	ID          string `json:"id"`
	TelegramID  int64  `json:"telegram_id"`
	Username    string `json:"username,omitempty"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) UpsertTelegramUser(ctx context.Context, telegramUser auth.TelegramUser) (User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return User{}, fmt.Errorf("begin user upsert: %w", err)
	}
	defer tx.Rollback(ctx) // no-op after Commit
	var id pgtype.UUID
	err = tx.QueryRow(ctx, `INSERT INTO users (telegram_id, username, first_name, last_name, photo_url)
VALUES ($1, NULLIF($2, ''), $3, NULLIF($4, ''), NULLIF($5, ''))
ON CONFLICT (telegram_id) DO UPDATE SET username = EXCLUDED.username, first_name = EXCLUDED.first_name, last_name = EXCLUDED.last_name, photo_url = EXCLUDED.photo_url
RETURNING id`, telegramUser.ID, telegramUser.Username, telegramUser.FirstName, telegramUser.LastName, telegramUser.PhotoURL).Scan(&id)
	if err != nil {
		return User{}, fmt.Errorf("upsert user: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO profiles (user_id, display_name) VALUES ($1, $2) ON CONFLICT (user_id) DO NOTHING`, id, telegramUser.FirstName)
	if err != nil {
		return User{}, fmt.Errorf("create profile: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO user_progress (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING`, id)
	if err != nil {
		return User{}, fmt.Errorf("create user progress: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return User{}, fmt.Errorf("commit user upsert: %w", err)
	}
	return s.GetByID(ctx, id.String())
}

func (s *Store) GetByID(ctx context.Context, id string) (User, error) {
	var user User
	err := s.pool.QueryRow(ctx, `SELECT u.id::text, u.telegram_id, COALESCE(u.username, ''), u.first_name, COALESCE(u.last_name, ''), COALESCE(u.photo_url, ''), p.display_name, p.level, p.xp, p.current_streak, p.longest_streak, p.preferred_difficulty, p.timezone, COALESCE(a.role, 'user')
FROM users u JOIN profiles p ON p.user_id = u.id LEFT JOIN admin_users a ON a.user_id=u.id WHERE u.id = $1::uuid`, id).Scan(&user.ID, &user.TelegramID, &user.Username, &user.FirstName, &user.LastName, &user.PhotoURL, &user.DisplayName, &user.Level, &user.XP, &user.CurrentStreak, &user.LongestStreak, &user.PreferredDifficulty, &user.Timezone, &user.Role)
	if err != nil {
		return User{}, err
	}
	rows, tenantErr := s.pool.Query(ctx, `SELECT t.id::text,t.slug,t.name,m.role,t.description,COALESCE(t.avatar_media_id::text,''),COALESCE(am.url,NULLIF(t.avatar_url,''),u.photo_url,''),COALESCE(u.photo_url,'') FROM tenant_memberships m JOIN tenants t ON t.id=m.tenant_id JOIN users u ON u.id=t.owner_user_id LEFT JOIN media_assets am ON am.id=t.avatar_media_id WHERE m.user_id=$1::uuid AND m.status='active' AND t.status='active' ORDER BY t.name`, id)
	if tenantErr == nil {
		defer rows.Close()
		for rows.Next() {
			var t Tenant
			if err = rows.Scan(&t.ID, &t.Slug, &t.Name, &t.Role, &t.Description, &t.AvatarMediaID, &t.AvatarURL, &t.TelegramAvatarURL); err != nil {
				return User{}, err
			}
			finishTenant(&t)
			user.Tenants = append(user.Tenants, t)
		}
		if err = rows.Err(); err != nil {
			return User{}, err
		}
	}
	user.AvailableModes = AvailableModes(user.Role)
	for _, tenant := range user.Tenants {
		if tenant.Role == "coach" && !contains(user.AvailableModes, "coach") {
			user.AvailableModes = append(user.AvailableModes, "coach")
		}
	}
	return user, nil
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

// BootstrapTenant is called only after Telegram initData (including
// start_param) has been cryptographically verified. It is the only path that
// may create a student membership from a public slug.
func (s *Store) BootstrapTenant(ctx context.Context, userID, slug string) (*Tenant, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var t Tenant
	if slug != "" {
		err = tx.QueryRow(ctx, `SELECT t.id::text,t.slug,t.name,t.description,COALESCE(t.avatar_media_id::text,''),COALESCE(am.url,NULLIF(t.avatar_url,''),u.photo_url,''),COALESCE(u.photo_url,'') FROM tenants t JOIN users u ON u.id=t.owner_user_id LEFT JOIN media_assets am ON am.id=t.avatar_media_id WHERE t.slug=$1 AND t.status='active'`, slug).Scan(&t.ID, &t.Slug, &t.Name, &t.Description, &t.AvatarMediaID, &t.AvatarURL, &t.TelegramAvatarURL)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, `INSERT INTO tenant_memberships(tenant_id,user_id,role) VALUES($1::uuid,$2::uuid,'student') ON CONFLICT(tenant_id,user_id) DO UPDATE SET status=CASE WHEN tenant_memberships.status='blocked' THEN 'blocked' ELSE 'active' END,updated_at=NOW()`, t.ID, userID)
		if err != nil {
			return nil, err
		}
	} else {
		err = tx.QueryRow(ctx, `SELECT t.id::text,t.slug,t.name,t.description,COALESCE(t.avatar_media_id::text,''),COALESCE(am.url,NULLIF(t.avatar_url,''),u.photo_url,''),COALESCE(u.photo_url,'') FROM tenant_memberships m JOIN tenants t ON t.id=m.tenant_id JOIN users u ON u.id=t.owner_user_id LEFT JOIN media_assets am ON am.id=t.avatar_media_id WHERE m.user_id=$1::uuid AND m.status='active' AND t.status='active' ORDER BY m.updated_at DESC LIMIT 1`, userID).Scan(&t.ID, &t.Slug, &t.Name, &t.Description, &t.AvatarMediaID, &t.AvatarURL, &t.TelegramAvatarURL)
		if err != nil {
			return nil, nil
		}
	}
	if err = tx.QueryRow(ctx, `SELECT role FROM tenant_memberships WHERE tenant_id=$1::uuid AND user_id=$2::uuid AND status='active'`, t.ID, userID).Scan(&t.Role); err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	finishTenant(&t)
	return &t, nil
}

// ResolveTenant selects an existing active membership. Unlike
// BootstrapTenant it never grants access, so X-Tenant-Slug is only a selector.
func (s *Store) ResolveTenant(ctx context.Context, userID, slug string) (*Tenant, error) {
	if slug == "" {
		return nil, nil
	}
	var t Tenant
	err := s.pool.QueryRow(ctx, `SELECT t.id::text,t.slug,t.name,m.role,t.description,COALESCE(t.avatar_media_id::text,''),COALESCE(am.url,NULLIF(t.avatar_url,''),u.photo_url,''),COALESCE(u.photo_url,'') FROM tenant_memberships m JOIN tenants t ON t.id=m.tenant_id JOIN users u ON u.id=t.owner_user_id LEFT JOIN media_assets am ON am.id=t.avatar_media_id WHERE m.user_id=$1::uuid AND t.slug=$2 AND m.status='active' AND t.status='active'`, userID, slug).Scan(&t.ID, &t.Slug, &t.Name, &t.Role, &t.Description, &t.AvatarMediaID, &t.AvatarURL, &t.TelegramAvatarURL)
	if err != nil {
		return nil, err
	}
	finishTenant(&t)
	return &t, nil
}

func (s *Store) ListTenants(ctx context.Context) ([]Tenant, error) {
	rows, e := s.pool.Query(ctx, `SELECT t.id::text,t.slug,t.name,t.description,t.status,t.owner_user_id::text,p.display_name,t.created_at::text,count(m.user_id) FILTER(WHERE m.role='student' AND m.status='active')::int FROM tenants t JOIN profiles p ON p.user_id=t.owner_user_id LEFT JOIN tenant_memberships m ON m.tenant_id=t.id GROUP BY t.id,p.display_name ORDER BY t.created_at DESC`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Tenant{}
	for rows.Next() {
		var t Tenant
		if e = rows.Scan(&t.ID, &t.Slug, &t.Name, &t.Description, &t.Status, &t.OwnerUserID, &t.OwnerName, &t.CreatedAt, &t.Students); e != nil {
			return nil, e
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func (s *Store) GetTenant(ctx context.Context, id string) (Tenant, error) {
	var t Tenant
	e := s.pool.QueryRow(ctx, `SELECT t.id::text,t.slug,t.name,t.description,t.status,t.owner_user_id::text,p.display_name,t.created_at::text,count(m.user_id) FILTER(WHERE m.role='student' AND m.status='active')::int,COALESCE(t.avatar_media_id::text,''),COALESCE(am.url,NULLIF(t.avatar_url,''),u.photo_url,''),COALESCE(u.photo_url,'') FROM tenants t JOIN profiles p ON p.user_id=t.owner_user_id JOIN users u ON u.id=t.owner_user_id LEFT JOIN media_assets am ON am.id=t.avatar_media_id LEFT JOIN tenant_memberships m ON m.tenant_id=t.id WHERE t.id=$1::uuid GROUP BY t.id,p.display_name,am.url,u.photo_url`, id).Scan(&t.ID, &t.Slug, &t.Name, &t.Description, &t.Status, &t.OwnerUserID, &t.OwnerName, &t.CreatedAt, &t.Students, &t.AvatarMediaID, &t.AvatarURL, &t.TelegramAvatarURL)
	finishTenant(&t)
	return t, e
}

func (s *Store) UpdateOwnTenantAvatar(ctx context.Context, user, tenant string, mediaID *string) (Tenant, error) {
	tag, err := s.pool.Exec(ctx, `UPDATE tenants t SET avatar_media_id=$3::uuid,updated_at=NOW() WHERE t.id=$2::uuid AND t.owner_user_id=$1::uuid AND EXISTS(SELECT 1 FROM tenant_memberships m WHERE m.tenant_id=t.id AND m.user_id=$1::uuid AND m.role='coach' AND m.status='active')`, user, tenant, mediaID)
	if err != nil {
		return Tenant{}, err
	}
	if tag.RowsAffected() == 0 {
		return Tenant{}, fmt.Errorf("tenant forbidden")
	}
	return s.GetTenant(ctx, tenant)
}
func (s *Store) UpdateOwnTenant(ctx context.Context, user, tenant, name, description string) (Tenant, error) {
	if len(strings.TrimSpace(name)) < 2 || len(name) > 120 || len(description) > 1000 {
		return Tenant{}, fmt.Errorf("invalid tenant settings")
	}
	tag, e := s.pool.Exec(ctx, `UPDATE tenants t SET name=$3,description=$4 WHERE id=$2::uuid AND owner_user_id=$1::uuid AND EXISTS(SELECT 1 FROM tenant_memberships m WHERE m.tenant_id=t.id AND m.user_id=$1::uuid AND m.role='coach' AND m.status='active')`, user, tenant, strings.TrimSpace(name), strings.TrimSpace(description))
	if e != nil {
		return Tenant{}, e
	}
	if tag.RowsAffected() == 0 {
		return Tenant{}, fmt.Errorf("tenant forbidden")
	}
	return s.GetTenant(ctx, tenant)
}

func (s *Store) UpdateOwnTenantSlug(ctx context.Context, user, tenant, slug string) (Tenant, error) {
	slug = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(slug, "@")))
	if err := validateTenantSlug(slug); err != nil {
		return Tenant{}, err
	}
	tag, err := s.pool.Exec(ctx, `UPDATE tenants t SET slug=$3,updated_at=NOW() WHERE id=$2::uuid AND owner_user_id=$1::uuid AND EXISTS(SELECT 1 FROM tenant_memberships m WHERE m.tenant_id=t.id AND m.user_id=$1::uuid AND m.role='coach' AND m.status='active')`, user, tenant, slug)
	if err != nil {
		if strings.Contains(err.Error(), "tenants_slug_key") || strings.Contains(err.Error(), "duplicate key") {
			return Tenant{}, fmt.Errorf("tenant slug already exists")
		}
		return Tenant{}, err
	}
	if tag.RowsAffected() == 0 {
		return Tenant{}, fmt.Errorf("tenant forbidden")
	}
	return s.GetTenant(ctx, tenant)
}

func (s *Store) SearchRoleUsers(ctx context.Context, query string) ([]RoleUser, error) {
	rows, err := s.pool.Query(ctx, `SELECT u.id::text,u.telegram_id,COALESCE(u.username,''),p.display_name,COALESCE(a.role,CASE WHEN EXISTS(SELECT 1 FROM tenant_memberships m WHERE m.user_id=u.id AND m.role='coach' AND m.status='active') THEN 'coach' ELSE 'user' END) FROM users u JOIN profiles p ON p.user_id=u.id LEFT JOIN admin_users a ON a.user_id=u.id WHERE $1='' OR COALESCE(u.username,'') ILIKE '%'||$1||'%' OR p.display_name ILIKE '%'||$1||'%' OR u.telegram_id::text=$1 ORDER BY p.display_name LIMIT 30`, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []RoleUser{}
	for rows.Next() {
		var x RoleUser
		if err = rows.Scan(&x.ID, &x.TelegramID, &x.Username, &x.DisplayName, &x.Role); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *Store) SetCoachRole(ctx context.Context, actor, target, role string) error {
	return s.SetCoachSpace(ctx, actor, target, role, "", "")
}

func (s *Store) SetCoachSpace(ctx context.Context, actor, target, role, name, slug string) error {
	if actor == target || (role != "user" && role != "coach") {
		return fmt.Errorf("unsafe role change")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var actorRole, oldRole string
	if err = tx.QueryRow(ctx, `SELECT role FROM admin_users WHERE user_id=$1::uuid`, actor).Scan(&actorRole); err != nil || actorRole != "super_admin" {
		return fmt.Errorf("super admin required")
	}
	if err = tx.QueryRow(ctx, `SELECT COALESCE((SELECT role FROM admin_users WHERE user_id=$1::uuid),'user')`, target).Scan(&oldRole); err != nil {
		return err
	}
	if oldRole == "admin" || oldRole == "super_admin" {
		return fmt.Errorf("protected role")
	}
	if role == "coach" {
		if name == "" {
			if err = tx.QueryRow(ctx, `SELECT display_name||' · Coach Space' FROM profiles WHERE user_id=$1::uuid`, target).Scan(&name); err != nil {
				return err
			}
		}
		if slug == "" {
			var existingSlug string
			if existingErr := tx.QueryRow(ctx, `SELECT slug FROM tenants WHERE owner_user_id=$1::uuid AND status='active'`, target).Scan(&existingSlug); existingErr == nil {
				slug = existingSlug // never rename an existing space implicitly
			} else {
				var username string
				_ = tx.QueryRow(ctx, `SELECT COALESCE(username,'') FROM users WHERE id=$1::uuid`, target).Scan(&username)
				candidates := []string{normalizeTenantSlug(username), normalizeTenantSlug(name), "coach-" + strings.ReplaceAll(target, "-", "")}
				for _, candidate := range candidates {
					if validateTenantSlug(candidate) != nil {
						continue
					}
					var available bool
					if queryErr := tx.QueryRow(ctx, `SELECT NOT EXISTS(SELECT 1 FROM tenants WHERE slug=$1)`, candidate).Scan(&available); queryErr == nil && available {
						slug = candidate
						break
					}
				}
			}
		}
		slug = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(slug, "@")))
		if validationErr := validateTenantSlug(slug); validationErr != nil {
			return validationErr
		}
		var tenantID string
		err = tx.QueryRow(ctx, `INSERT INTO tenants(name,slug,owner_user_id) VALUES($1,$2,$3::uuid) ON CONFLICT(owner_user_id) WHERE status='active' DO UPDATE SET name=EXCLUDED.name,slug=EXCLUDED.slug,updated_at=NOW() RETURNING id::text`, name, slug, target).Scan(&tenantID)
		if err != nil && (strings.Contains(err.Error(), "tenants_slug_key") || strings.Contains(err.Error(), "duplicate key")) {
			return fmt.Errorf("tenant slug already exists")
		}
		if err == nil {
			_, err = tx.Exec(ctx, `INSERT INTO tenant_memberships(tenant_id,user_id,role,status) VALUES($1::uuid,$2::uuid,'coach','active') ON CONFLICT(tenant_id,user_id) DO UPDATE SET role='coach',status='active',updated_at=NOW()`, tenantID, target)
		}
	} else {
		_, err = tx.Exec(ctx, `UPDATE tenant_memberships SET status='inactive',updated_at=NOW() WHERE user_id=$1::uuid AND role='coach'`, target)
	}
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO role_change_audit(actor_user_id,target_user_id,old_role,new_role) VALUES($1::uuid,$2::uuid,$3,$4)`, actor, target, oldRole, role)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
