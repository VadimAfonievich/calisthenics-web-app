# Architecture

## Phase 0 discovery record

This document records the decisions for the Calisthenics Telegram Mini App MVP. It is derived from `CALISTHENICS_TELEGRAM_MINI_APP_MASTER_SPEC.md` and the actual repository state inspected on 2026-08-13.

### Repository state

- The repository has no commits and its only project file is the master specification.
- `README.md`, application code, package manifests, Docker configuration, migrations, CI configuration, and tests are absent.
- There are no existing applications, services, entry points, dependencies, or reusable implementation assets.
- The implementation will therefore start from the empty-repository layout required by the specification. The master specification itself remains the only reusable project asset and the source of requirements.

## Architectural decisions

The MVP is a modular monolith, not a set of microservices. A single Go API process owns all business rules and data writes; a single React Mini App consumes its REST API.

```text
Telegram
  ├─ Bot (/start, Web App button)
  └─ Mini App (React + TypeScript)
          │ Telegram initData; Bearer JWT
          ▼
   Go API — /api/v1 (Chi)
  ├─ auth and Telegram signature validation
  ├─ domain services and HTTP handlers
  ├─ SQLC repositories
  ├─ PostgreSQL (source of truth)
  └─ Redis (cache and rate limiting only)
```

| Area | Decision | Rationale |
| --- | --- | --- |
| Backend HTTP | Go with `chi` | Lightweight, idiomatic router suitable for a modular monolith. |
| API | REST under `/api/v1`; OpenAPI is the API contract | Versioning and a testable, documented frontend/backend boundary. |
| Persistence | PostgreSQL, UUID primary keys, `timestamptz` timestamps | One consistent identifier strategy, safe distributed creation, and time-zone-aware records. |
| Data access | `sqlc` and parameterized SQL | Typed queries and no manually assembled SQL. |
| Schema changes | `golang-migrate` migrations only | Reproducible database evolution; no runtime schema mutation. |
| Cache/rate limits | Redis | It is never a source of truth or a workflow store. |
| Authentication | Backend verification of Telegram `initData`, then short-lived JWT access token | The client never supplies a trusted Telegram identity. |
| Frontend | React + TypeScript + Vite + React Router + TanStack Query | Matches the specification and separates routing from server-state caching. |
| Client state | React state by default; Zustand only for session/preferences | Server data has one owner: TanStack Query. |
| Styling | Tailwind CSS | Mobile-first, compact implementation that can follow Telegram theme variables. |
| Runtime | Docker Compose with frontend, backend, PostgreSQL, and Redis | Identical local startup for the four required services. |
| Logging | Structured JSON logs with request ID | Supports diagnosis while excluding secrets, JWTs, and raw Telegram initData. |

## Repository layout

The following layout is the target of Phase 1 and later phases. Directories are not created during Phase 0 except for this documentation directory.

```text
backend/
  cmd/server/main.go
  internal/
    auth/ users/ lessons/ exercises/ workouts/
    progress/ achievements/ stats/ admin/
    handler/ service/ repository/ middleware/ telegram/ config/
    db/migrations/ db/queries/ db/sqlc/
  docs/
frontend/
  src/app/ src/pages/ src/components/ src/api/ src/hooks/
  src/store/ src/types/ src/utils/
  public/
docker/
docs/
docker-compose.yml
.env.example
README.md
```

`cmd/server/main.go` will be the backend executable entry point. The Vite entry point will be `frontend/src/main.tsx`. The bot is part of the Go process and is configured by environment variables; it is not a separate microservice.

## Backend boundaries

Each domain owns its handler, application service, and repository interface/queries while sharing cross-cutting infrastructure. HTTP handlers only validate and translate requests; business rules belong in services; repositories only perform persistence.

- `auth`: validates Telegram initData, creates or updates a user, issues JWT, and supplies authentication middleware.
- `users`: exposes `/me` and profile data, including the user's IANA timezone.
- `lessons` and `exercises`: catalog and detail reads; lesson completion is idempotent.
- `workouts`: programs, today's workout, sessions, set recording, and idempotent completion.
- `progress`: XP ledger/calculation, levels, streaks, statistics, and history. The backend alone calculates all of them.
- `achievements`: evaluates conditions and persists one unlock per user/achievement.
- `admin`: role-protected content management endpoints.

Critical multi-entity operations, particularly workout completion, run in one PostgreSQL transaction. The transaction changes session state, progress, XP, streak, and achievement unlocks atomically. Database uniqueness constraints provide the second line of defense against duplicate lesson completion, duplicate XP, and duplicate achievement rewards.

## Data model and invariants

Core tables required by the specification are `users`, `profiles`, `lesson_categories`, `lessons`, `user_lesson_progress`, `exercises`, `programs`, `workouts`, `workout_exercises`, `workout_sessions`, `exercise_sets`, `user_progress`, `user_exercise_stats`, `achievements`, `user_achievements`, and `admin_users`.

The detailed schema will be designed in Phase 2, with foreign keys, indexes, checks, and unique constraints. At minimum, it will enforce a unique Telegram ID, one profile per user, one lesson-progress row per user/lesson, one achievement unlock per user/achievement, and ownership of every workout session and set. Exercise and lesson content is data in PostgreSQL, never hardcoded into the application.

XP thresholds are configuration or database data, not frontend constants. Streak calculations use the timezone persisted in the profile and record a maximum of one streak update per local calendar day.

## API, security, and client flow

The Mini App initializes the Telegram WebApp SDK and sends its raw `initData` only to `POST /api/v1/auth/telegram`. The backend validates Telegram's HMAC signature with the bot token, finds or creates the user, and returns a JWT. Protected endpoints read the user exclusively from the validated JWT.

The API uses the common error envelope:

```json
{"error":{"code":"WORKOUT_NOT_FOUND","message":"Workout not found"}}
```

All public input is validated. Admin endpoints require an admin role; CORS is explicit; rate limiting is Redis-backed; secure headers are set; logs exclude credentials, raw initData, and sensitive personal data. OpenAPI documents all public API endpoints under `/api/v1`.

## Frontend composition

The Mini App is mobile-first and implements the specified routes: home, lesson list/detail, exercise list/detail, today's workout/workout detail, progress, achievements, profile, and the isolated admin routes. It provides Telegram theme, safe-area, viewport, loading, empty, and error states. TanStack Query owns API responses and invalidation after mutations; the UI never computes or assigns XP, level, streak, or completion results.

## Concrete implementation plan

1. **Phase 1 — Foundation:** create the prescribed backend/frontend/docker/docs layout; add Go/Chi server, Vite React TypeScript app, Docker Compose services, typed environment configuration, health endpoint, CORS, structured request logging, and OpenAPI foundation. Add `.env.example` and an operational `README.md`.
2. **Phase 2 — Database:** add golang-migrate migrations for the normalized schema and constraints, SQLC configuration/queries, database connection lifecycle, and deterministic seed data (10 lessons, 10 exercises, 2 programs, 5 workouts, 10 achievements). Verify migration up/down and queries.
3. **Phase 3 — Telegram authentication:** implement Telegram signature validation, user upsert, JWT issuance and middleware, `/me` and profile endpoints, and auth-validation tests.
4. **Phase 4 — Frontend foundation:** initialize Telegram SDK, authentication bootstrap, router, API client, Query provider, minimal Zustand session store, responsive layout/navigation, and common state handling.
5. **Phase 5 — Lessons:** implement category/list/detail/completion APIs and pages, with idempotent completion and backend XP award.
6. **Phase 6 — Exercises:** implement database-backed catalog, filters, detail endpoints, and mobile pages.
7. **Phase 7 — Training engine:** implement program/workout reads, sessions, set recording, timers, history, ownership controls, transactional and idempotent completion, and the completion UI.
8. **Phase 8 — Progress and gamification:** implement XP service, configurable levels, timezone-aware streak service, statistics/history, achievements service, and associated UI.
9. **Phase 9 — Admin:** add admin authorization and minimal content CRUD/publish controls for lessons, exercises, programs, workouts, and workout exercises.
10. **Phases 10–12 — Hardening, testing, production:** add rate-limit/security/transaction/race-condition coverage; run backend unit/integration/API tests plus frontend component and smoke/E2E flow; finalize production Docker, deployment, backup, monitoring, and Telegram setup documentation.

Each phase ends with its specified checks and documentation update before moving forward. No code from Phase 1 or later is introduced by this Phase 0 change.

## Phase 1 delivery record

Phase 1 created the documented foundation: the Go/Chi server, PostgreSQL and Redis readiness checks, JSON structured request logging, CORS and security headers, the `/healthz` endpoint, OpenAPI foundation, React/Vite/Tailwind client, Dockerfiles, Docker Compose stack, `.env.example`, and `README.md`.

The backend only starts after it can connect to PostgreSQL and Redis. The health endpoint checks both dependencies on every request and returns the standard error envelope if either is unavailable. Runtime migration and domain data are deliberately deferred to Phase 2.

Validated locally on 2026-08-13:

- `go test ./...` passed, including health, CORS, and error-envelope tests.
- `npm run build` passed.
- `docker-compose config --quiet` passed.

Container-runtime validation completed on 2026-08-13 after Docker Desktop was started:

- `docker-compose up -d` started PostgreSQL, Redis, backend, and frontend.
- PostgreSQL, Redis, and backend reported healthy status.
- `GET http://localhost:8080/healthz` returned `200 {"status":"ok"}`.
- `GET http://localhost:8080/openapi.yaml` returned the OpenAPI 3.0.3 foundation document.
- `GET http://localhost:3000/` returned the frontend shell with HTTP 200.

The initial container build exposed an unreliable connection from Docker to `proxy.golang.org`. Go dependencies are now vendored and the backend Docker build uses `-mod=vendor`, making the backend image build independent of the module proxy. Docker build contexts now exclude generated frontend dependencies and build output.

## Phase 3 delivery record

Phase 3 adds `POST /api/v1/auth/telegram` and the protected `GET /api/v1/me` endpoint. The backend validates the Telegram WebApp `initData` signature server-side, rejects stale payloads, provisions a user, profile, and progress row transactionally, and issues a 24-hour HS256 JWT. Protected routes only obtain the user ID from a verified Bearer token.

The Telegram bot token remains intentionally unset in `.env.example`; without it, the authentication endpoint safely reports `AUTH_UNAVAILABLE` rather than accepting an unverified client identity. Unit tests cover valid, expired, and tampered Telegram initData, JWT verification, and the protected `/me` route.

## Phase 4 delivery record

Phase 4 establishes the React Mini App client flow: Telegram WebApp initialization, application of Telegram theme values and safe-area support, a Telegram authentication bootstrap, Zustand session state, a typed API client with the common error envelope, and TanStack Query as the server-state provider. The mobile-first application shell includes the specified MVP routes, loading/error/demo states, and bottom navigation.

When opened outside Telegram, the app intentionally enters a clearly labeled demo mode rather than attempting to create a trusted user identity. `npm run build` passed on 2026-08-18.
