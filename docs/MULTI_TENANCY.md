# Multi-tenancy

Phase 19A introduces logical tenant isolation in the existing PostgreSQL database. It does not create a database, schema, frontend, or bot per coach.

## Identity and roles

`tenants` is the stable Coach Space record. Its lowercase URL-safe `slug` is public identification, never authorization. `tenant_memberships` is the many-to-many boundary between users and spaces; `(tenant_id,user_id)` is unique and membership has an independent `coach` or `student` role and access status.

Platform roles remain in `admin_users` and are restricted to `admin` and `super_admin`. A tenant coach is authorized through an active coach membership, not a platform role. This permits one Telegram user to join multiple spaces and leaves room for a future tenant subscription/entitlement check.

## Tenant resolution and Telegram links

Coach links use `https://t.me/<bot_username>?startapp=<tenant_slug>`. Telegram includes the value as `start_param` in signed init data. The backend validates the entire init-data signature before parsing the slug, validates the slug syntax, resolves only an active tenant, and idempotently creates/activates a student membership without replacing a coach membership. The authentication response contains `current_tenant` and all active `tenants` available to the user.

Subsequent requests send the stable slug in `X-Tenant-Slug`. Authentication middleware resolves it server-side and adds verified tenant ID and membership role to request context. UUIDs or slugs supplied by clients are never treated as privileges. A direct open without a slug selects the most recently active membership; with no membership it returns no tenant context, so tenant-owned content cannot be aggregated across spaces.

## Content ownership

Lessons, workouts, programs, skills, media, and coach exercises carry `tenant_id`. New owner-authored content is assigned to an active space by a database trigger, and its owner must be an active coach member of that space.

`owner_user_id` remains author/audit metadata and is not a visibility boundary. Every tenant-owned read and mutation is authorized by `tenant_id` plus the verified membership from request context.

Standard exercises are the sole global content type: `tenant_id IS NULL`, `owner_user_id IS NULL`, and `standard_key IS NOT NULL`. Coach-created exercises require both tenant and owner. Database triggers reject a workout using another tenant's exercise, a workout attached to another tenant's program, and cross-tenant skill prerequisites. Global exercises remain valid workout dependencies.

Media storage keys should use `tenants/<tenant_id>/media/...`; existing external/data-URL media retains its provider key during migration. Platform standard-exercise media remains global.

## Progress, calendar, and analytics

Workout sessions, lesson/program/skill progress, training schedules, planned workouts, and exercise statistics carry tenant context. Program progress is keyed by `(user_id,tenant_id,program_id)`. APIs must verify both the requested content and progress/session tenant; a matching UUID alone is insufficient.

Coach analytics resolve the coach's active space and aggregate its active student memberships and tenant-scoped activity. Platform totals belong only in Administration mode. The former implementation queried `users` and all activity directly, which exposed global metrics to every coach.

`access.Service` remains the central entitlement seam. A future subscription can add `(user_id,tenant_id,status)` checks there without scattering payment logic through handlers. PostgreSQL RLS is intentionally deferred until all repository calls set transaction-local tenant identity; enabling it prematurely would either break jobs/migrations or create bypasses.

## Migration and rollout

Migration `000016_multi_tenancy` is reversible and uses the next schema version. It creates one deterministic space for each distinct existing owner (including legacy coach-role users), assigns all of that owner's content, makes the owner a coach member, derives student memberships from existing activity, and leaves system content global. This avoids assuming that production has exactly one original coach. Product owners must review generated names/slugs and may rename them through the administration flow after migration.

Before production rollout:

1. Back up PostgreSQL and restore it into staging.
2. Run up, down, and up migrations against a representative production-data copy.
3. Inspect all global exercises for non-null `standard_key` and all owned content for a tenant.
4. Run PostgreSQL isolation E2E fixtures for two coaches/two students, direct UUID attacks, analytics counts, deep-link resolution, and same-tenant composition constraints.

## Operational semantics

Coach promotion is one database transaction: tenant creation/update, active coach membership, and audit record either all commit or all roll back. Repeating promotion reuses the owner's single active tenant; it does not create another tenant. Demotion only marks coach memberships inactive. It deliberately preserves the tenant, students, content, progress, and history for later administrative recovery; repeating demotion is safe.

Tenant slugs are immutable in the MVP UI. This avoids invalidating previously shared Telegram links. A future slug-change workflow must explicitly handle old links and uniqueness/reserved-name validation.

Migration 16 down is structurally safe but not a perfect semantic rollback: it removes tenant and membership records and all tenant columns, so tenant selection, membership roles/statuses, tenant descriptions, and tenant-scoped distinctions cannot be reconstructed from the downgraded schema. It does not intentionally delete legacy users, content, media, or progress rows.

## Tenant SQL Boundary Audit

| Repository/service | Entity and purpose | Classification | Enforcement |
|---|---|---|---|
| `middleware`, `users.Store` | Resolve current space and memberships | TENANT_SCOPED | Signed bootstrap or `X-Tenant-Slug` may select only an active membership; UUID is resolved server-side. |
| `lessons.Service` | Student list/detail/completion | TENANT_SCOPED | Published lesson and progress operations require the current `tenant_id`; media is reachable only through that lesson. |
| `exercises.Service` | Student library/detail/search | TENANT_SCOPED + GLOBAL SYSTEM CONTENT | Current tenant exercises plus rows with `tenant_id IS NULL`, no owner, and a `standard_key`. |
| `workouts.Service` | Catalog, start/resume/set/finish | TENANT_SCOPED | Workout, session, planned item and set writes require current tenant. Exercise statistics are keyed by user, tenant and exercise. |
| `programs.Service` | Catalog/detail/start/progress refresh | TENANT_SCOPED | Program, workouts, sessions and `user_program_progress` are constrained to current tenant. |
| `skills.Service` | Map/detail/criteria/levels/mastery | TENANT_SCOPED | Parent skill is verified first; progress rows carry the same tenant. Child IDs are reachable only from the verified parent. |
| `calendar.Service` | Schedules, planned workouts, calendar/today | TENANT_SCOPED | Every direct ID read/mutation checks user and current tenant; referenced workout must share the tenant. |
| `progress.Service` | Summary/history/statistics | TENANT_SCOPED | Session and set aggregations filter by current tenant. Profile XP/level and account achievements remain intentional user-global gamification state. |
| `coach.Service` | Studio CRUD, lifecycle, duplicate, options, media and analytics | TENANT_SCOPED | Current coach membership is required by routing; every entity read/mutation uses current `tenant_id`. `owner_user_id` is an additional author check, never the sole boundary. |
| `users.Store` role/tenant administration | User search, tenant list/detail, promotion/demotion | SUPER_ADMIN_GLOBAL | Handler verifies platform `super_admin`; coach promotion is transactional. |
| `admin.Service` | Legacy platform content administration | SUPER_ADMIN_GLOBAL | Mounted only behind the platform-admin guard; it is not used by Coach mode. |
| levels, achievements, lesson categories | Shared taxonomy and gamification definitions | GLOBAL INTENTIONAL | Definitions contain no coach/student private data. |
| standard exercises | Shared exercise catalog | GLOBAL SYSTEM CONTENT | Only ownerless rows with a stable `standard_key`; DB constraints prevent owned rows from becoming global. |
| generated `internal/db/sqlc` legacy queries | Historical generated accessors | GLOBAL INTENTIONAL (inactive) | No runtime package imports or callers remain; active HTTP services use the tenant-aware repositories above. |

Direct-ID routes were reviewed for lessons, workouts, sessions, programs, skills, criteria/levels, schedules and planned workouts. The tenant check occurs before mutation. Database triggers additionally reject mismatched session/workout, calendar/workout, exercise-set/exercise, content/media, program/workout and skill dependency relationships.

The old platform `coach` role is removed by migration 16 and is not accepted by the `admin_users` constraint. Coach authorization comes from an active `tenant_memberships(role='coach')` row for the selected tenant. Platform roles cannot bypass Coach Studio tenant checks.
5. Deploy backend before frontend in a controlled maintenance window, verify generated spaces/share links, then enable traffic.

No migration or deployment is performed automatically by this phase.
# Global system content

Platform content now has two explicit global classes: standard exercises and system workouts. Both use `tenant_id IS NULL`, `owner_user_id IS NULL`, a stable `standard_key`, and published status. System workouts are read-only, visible in every active tenant, and may reference only global standard exercises. A coach can duplicate one into the current school; the copy receives the current `tenant_id` and coach `owner_user_id` and is fully editable. Tenant workouts may combine global standard exercises with exercises owned by that same tenant.

Production rollout is deliberately separate from application deployment: back up the database, apply migration 17, confirm the 247 standard-exercise catalog is present, then run `DATABASE_URL=... go run ./cmd/system-workouts` from `backend`. The importer is idempotent and stops with the missing `standard_key` if the catalog is incomplete. Verify `schema_migrations.dirty=false` and the four published workout keys before deploying the application. Rollback must be planned together with seeded data and any student sessions already referencing system workouts.
