# Coach Content Studio

Coach Studio is available at `/coach` and its API under `/api/v1/coach`. Access is decided by `admin_users.role`: `coach` manages rows whose `owner_user_id` is their user ID, while `admin` and `super_admin` can manage all content. A normal user receives `403`. Existing seed rows have a null owner and remain system content.

## Authorization role and application mode

The PostgreSQL `admin_users.role` is the security authority. `GET /api/v1/me` returns that role plus server-derived `available_modes`. The frontend application mode (`student` or `coach`) only selects navigation and presentation and is stored in `localStorage` as `calisthenics_app_mode`. It never grants API permissions. Every `/api/v1/coach/*` request still checks the database-backed role and returns `403` for a normal user, even if local storage is manually changed.

Authorized coaches and administrators can switch modes from `/profile`. A saved coach mode is restored only when the current `/me` response still permits it; otherwise it falls back to student mode.

## Mobile authoring

`/coach/content` is the mobile content hub for lessons, exercises, workouts, programs, and skill progressions. Editors use server-provided selector options instead of exposing UUIDs. PostgreSQL generates entity IDs, and the service derives collision-safe slugs for entities that require them. Existing owned content opens through a validated UUID route and is prefilled from the Coach detail endpoint. Seed content is visible to coaches as read-only system content; a coach can inspect it or create an owned draft copy. Administrators can edit system content directly. Dashboard and list counts use this same `system + owned` scope.

Lesson JSON blocks are presented with human labels and simple up/down ordering. Workout exercises, program stages, and skill prerequisites use catalog selectors. Forms disable saves in flight and warn about unsaved changes through browser navigation and the Telegram BackButton.

## Relationships and independent lifecycle

The authoring dependency direction is `Exercise → Workout ← Program → Skill`. A workout is standalone and stores its own difficulty plus ordered exercise rows (sets, repetitions XOR duration, rest, notes); its editor does not select a program, level, or day. A program owns composition: its editor selects and orders existing workouts, and the service maintains nullable legacy `program_id`, `program_level_id`, and `day_number` fields for compatibility. The current schema supports one program assignment per workout. Removing a workout from a program makes it standalone without deleting it. A skill level can link to a program level, and a skill can require other skills.

A workout can offer a reusable standard warmup with `warmup_enabled` (true by default outside the `warmup` category). Warmup exercises are never copied into the main workout. The published `category='warmup'` row marked `is_default_warmup` is returned as runtime preview metadata. Completing a warmup records exercise/activity time, but does not increment full-workout count, streak, workout achievements, progression criteria, or XP.

The optional Voice Coach wraps browser Speech Synthesis rather than calling it throughout player components. It uses `ru-RU`, follows the same visual `remaining` state for final 5-second timed/rest announcements, cancels obsolete speech on player actions, and falls back silently to visual countdown plus Telegram haptics. The user preference is local under `calisthenics_voice_coach_enabled` and defaults on; the player provides a test action initiated by user gesture.

Skills use the existing `skills.sort_order` as an explicit display position. The Coach editor expresses it as “После какого навыка?”, while prerequisites remain a separate dependency concept. Student queries order by `sort_order`, then name and ID for deterministic ties.

Every lesson, exercise, workout, program, and skill has its own `draft`, `published`, and `archived` lifecycle. Publishing never cascades:

- a published workout may contain only published exercises;
- a published program may contain only published workouts;
- a published skill may reference only published programs and prerequisite skills;
- restoring archived content always returns it to `draft`.

Draft parents may use draft dependencies. Archived entities are hidden from new selector choices. Student APIs expose only published content; historical sessions keep their foreign-key references. Archiving is rejected when it would break published dependent content, with a dependency count in the validation message.

Duplicate creates a new coach-owned UUID in `draft`. Lesson blocks, workout exercise rows, program levels, and skill levels are copied, while learner progress, sessions, and analytics are never copied. System content is never modified by a coach through duplication.

Selector options return readable names plus lifecycle status, ownership, and parent IDs where relevant. Editors show the status next to dependency names so draft dependencies are visible before publishing. Options exclude archived content and provide explicit loading, error, and retry states.

Legacy `published` flags on lessons and programs are synchronized for compatibility. Publishing records `published_at` and `published_by`.

Lessons retain legacy `content` and add ordered JSONB blocks. Allowed blocks are `heading`, `text`, `image`, `video`, `tip`, `warning`, `checklist`, and `divider`. The backend validates block shape. The mobile editor supports adding, removing, and reordering blocks. Other content lists expose search, status filtering, lifecycle actions, and provide the foundation for specialized workout/program/skill builders.

Analytics is aggregate-only: user activity, lesson/workout completions, popular workouts, skill progress, and achievements. There is no coach-to-student tenancy relation yet, so coach analytics currently describes the global learner population; individual CRM and payments are intentionally excluded.

## Student training taxonomy and readiness gate

New workout writes use one of four canonical categories: `warmup`, `morning`, `strength`, or `skill`. The Coach selector presents these as «Разминка», «Зарядка», «Развитие силы», and «Тренировка навыков». Existing program taxonomy remains compatible: legacy values are mapped at migration/read time and are not shown as additional student workout categories.

The student `/workouts` page is the training home: it combines the four category filters with small published-program and skill previews. Full program routes are `/programs` and `/programs/:id`; programs are ordered workout collections, while `/skills` remains the separate progression graph.

`CALISTHENICS_BASE` is system-owned and is shown first in the skill graph. Its eight readiness checks are rows in `skill_criteria`, with current-user confirmations in `user_skill_criteria`; clients never submit a user ID. Confirming a criterion through `POST /api/v1/skills/{id}/criteria/{criterionID}/confirm` is idempotent. The eighth confirmation masters the base, awards XP/achievement once, and unlocks skills through ordinary `skill_requirements`. `PULL_UP_BASE` and `DIP_BASE` remain preserved as hidden technical records rather than visible graph nodes. Like all system content, the base is read-only for a regular coach and manageable only through an administrator-authorized system-content workflow.
