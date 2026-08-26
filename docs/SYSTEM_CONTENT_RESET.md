# System content reset and exercise library foundation

## Classification

`owner_user_id IS NULL` is system content; a non-null owner is coach content. Ownership exists on lessons, exercises, workouts, programs, skills, and media assets. Users, profiles, progress, sessions, sets, schedules, planned workouts, achievements earned by users, and skill/lesson progress are user/history data and are never reset.

Target system content is published universal exercises only. System lesson categories, levels, achievements, and role records remain as service/reference data. `CALISTHENICS_BASE` is not required for server startup, but deleting it would destroy skill criteria and user mastery history through existing cascading skill FKs; the reset therefore archives/hides it like every other system skill.

## Dependency graph

```text
lesson_categories -> lessons <- user_lesson_progress
media_assets <- lessons | exercises | workouts | programs | skills
programs -> program_levels <- workouts
programs <- workouts -> workout_exercises -> exercises
workouts <- workout_sessions -> exercise_sets -> exercises
workouts <- user_training_schedules -> schedule_days
workouts <- user_planned_workouts <- workout_sessions
skills -> skill_levels -> user_skill_level_progress
skills -> skill_requirements -> skills
skills -> skill_criteria -> user_skill_criteria
skills -> user_skill_progress
achievements -> user_achievements <- users
users -> profiles | user_progress | user_exercise_stats | all history/progress tables
```

Physical deletion is unsafe: most workout, lesson, exercise, calendar, achievement, and session history references use `ON DELETE RESTRICT`; some skill relations use `CASCADE` and would silently erase learner progression. The supported reset is therefore a reversible lifecycle archive, not DELETE.

## Reset command

Build or run `./backend/cmd/content-reset`. It requires `DATABASE_URL` and always requires `--system-only`.

```sh
go run ./cmd/content-reset --system-only --dry-run
go run ./cmd/content-reset --system-only --confirm
```

Dry-run prints counts and conflicts. Confirm archives system lessons/workouts/programs/skills, hides system skills, clears the default-warmup marker, and ensures system exercises remain published. It never updates rows with a non-null owner and never touches user/history tables. Take a database backup first. Do not run against production as part of deploy automation.

## Exercise model

The effective exercise record contains identity (`id`, stable `slug`), instructional text, difficulty, `movement_type`, muscle groups, equipment, legacy image/video URLs, optional cover media, coach tips, ownership, and lifecycle/publishing fields. System exercises are `owner_user_id=NULL,status=published`; coach UI may read/use them but ordinary coach updates are owner-guarded. Coach exercises are created as owned drafts and support publish/archive/duplicate.

Current hard enums and recommended vocabulary are in `EXERCISE_ENUMS.json`. Migration `000012_exercise_library_metadata` adds the missing family/skill `tags` array and GIN/filter indexes; it is reversible and is not applied automatically. Existing selector scope already includes system plus the current coach's non-archived content; published workouts are validated before publishing.

## Validate-only standard catalog

The input format has no UUIDs. `key` is a stable lowercase kebab-case identity which maps naturally to the persisted slug during a future importer. No write importer exists yet.

```sh
go run ./cmd/standard-exercises --validate-only --file ../docs/standard-exercises.example.json
```

Validation rejects unknown JSON fields, duplicate keys, invalid enums, missing instructional fields, empty/duplicate muscle groups, duplicate equipment/tags, invalid tag keys, and non-HTTPS media URLs.

For generation of 200 exercises, provide an AI exactly these files:

1. `docs/standard-exercises.schema.json`
2. `docs/EXERCISE_ENUMS.json`
3. `docs/standard-exercises.example.json`
4. `docs/exercise-library.schema.json`
5. `docs/SYSTEM_CONTENT_RESET.md`

Generated output must be a single catalog matching `standard-exercises.schema.json`; validate it locally before any import implementation or database write.
