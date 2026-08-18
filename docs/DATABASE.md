# Database

## Overview

PostgreSQL is the source of truth for Calisthenics Coach. The database is a normalized schema for one modular monolith and is changed only through `golang-migrate` migration files in `backend/internal/db/migrations`.

The Phase 2 schema contains 16 tables: `users`, `profiles`, `lesson_categories`, `lessons`, `user_lesson_progress`, `exercises`, `programs`, `workouts`, `workout_exercises`, `workout_sessions`, `exercise_sets`, `user_progress`, `user_exercise_stats`, `achievements`, `user_achievements`, and `admin_users`.

## Identifier and time strategy

Every entity identifier and foreign key is a PostgreSQL `uuid`. UUIDs are generated with `gen_random_uuid()` from `pgcrypto`; this retains a single ID representation across API, application, and database boundaries without relying on sequence allocation.

All event and audit times are `timestamptz` in UTC. `profiles.timezone` stores a user IANA timezone such as `Asia/Novosibirsk`; it is required later to calculate a local calendar-day streak correctly.

## Relationships

- A user has exactly one profile and one aggregate `user_progress` row once an authenticated user is provisioned.
- Lessons belong to a category; user lesson progress is one row per user/lesson.
- Workouts belong to a program; workout exercises attach catalog exercises to a workout.
- Sessions belong to a user and workout; recorded sets belong to a session and exercise.
- Per-exercise stats and unlocked achievements are one row per user/exercise and user/achievement respectively.
- An admin row grants either `admin` or `super_admin` role to exactly one user.

Foreign keys use `ON DELETE RESTRICT`. Content and historical user records are protected from accidental cascading deletion; future admin deletion flows must explicitly account for references.

## Arrays and data representation

`exercises.muscle_groups` and `exercises.equipment` are PostgreSQL `text[]`. They are short, display-oriented filter values with no standalone metadata or admin taxonomy in MVP. Arrays avoid premature join tables while allowing `ANY(...)` filtering later. If those values acquire descriptions, aliases, translations, or independently managed lifecycle, they should be migrated to catalog tables then.

## Constraints and indexes

The schema uses UUID primary keys, foreign keys, meaningful uniqueness, and validation checks:

- unique Telegram ID, slugs, achievement codes, one program day, lesson progress, set number, exercise position, user achievement, and admin user;
- non-negative XP, streaks, times, weights, counts, and ordering values;
- percentage between 0 and 100;
- controlled difficulty, session status, achievement condition type, and admin role values;
- exactly one planned load target (repetitions or duration) for a workout exercise and one recorded result type for a set;
- completed lessons require 100% progress and a completion time; completed sessions require a completion time.

There are 11 explicit secondary indexes: published lessons, lesson category, both lesson-progress lookup paths, workout program, workout exercises, three workout-session history/lookups, exercise sets by session, and user achievements. Unique/primary-key constraints additionally provide the indexes needed for their lookup paths. `users.telegram_id` is indexed by its unique constraint.

## Stored and derived progress

`profiles` is authoritative for XP, level, and streak values. `user_progress` stores only materialized counters that are expensive or inconvenient to aggregate on every home/progress request: total workouts, completed exercises, and training seconds. `user_exercise_stats` similarly stores per-exercise aggregate counters.

These materialized aggregates are not client writable. Their reconciliation source is completed workout sessions and completed exercise sets. Phase 7/8 services will update them transactionally.

## Transaction boundary

The future `Complete workout` service must perform one PostgreSQL transaction:

```text
BEGIN
  complete workout session
  update aggregate progress and per-exercise stats
  calculate and store XP; update profile XP and level
  update streak using profile timezone
  unlock eligible achievements
COMMIT
```

Failure rolls back the complete operation. The unique constraints and a status-guarded session update support idempotency, but the business service is intentionally deferred beyond Phase 2.

## Migrations and seed workflow

`000001_create_schema` creates the schema. `000002_seed_demo_content` adds deterministic Russian demo content: 4 lesson categories, 10 lessons, 10 exercises, 2 programs, 5 workouts, 16 workout-exercise assignments, and 10 achievements. `000003_progress_gamification` adds the configurable level thresholds and the last local workout date used by streak calculation. The seed uses fixed UUIDs only for reference content; production rows use generated UUIDs.

Run migrations against the Docker PostgreSQL service:

```powershell
docker-compose --profile tools run --rm migrate
```

For a rollback, override the final migrate command with `down 1`; then run `up` again. The down migration for seed removes only reference content. The schema down migration removes the tables in dependency order.

Generate typed SQLC code after query or schema changes:

```powershell
Set-Location backend
sqlc generate
```

Generated files in `backend/internal/db/sqlc` are never edited manually.
