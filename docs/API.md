# Calendar and scheduling API

All endpoints require a Bearer JWT and enforce current-user ownership. Dates use `YYYY-MM-DD`, times use `HH:MM`, and JSON uses `snake_case`.

- `GET /api/v1/calendar?from=YYYY-MM-DD&to=YYYY-MM-DD` — bounded recurrence projection with persisted status overrides.
- `GET /api/v1/calendar/today` — today's events in profile timezone, timed entries first.
- `GET|POST /api/v1/training-schedules` and `PUT|DELETE /api/v1/training-schedules/{id}` — recurring schedules; DELETE disables.
- `POST /api/v1/planned-workouts`, `GET|PUT|DELETE /api/v1/planned-workouts/{id}` — one-off events and occurrence materialization.
- `POST /api/v1/planned-workouts/{id}/skip` — explicit skip.
- `POST /api/v1/workouts/{id}/start` optionally accepts `{ "planned_workout_id": "uuid" }`; completion updates that owned planned event in the same transaction.
# Coach Studio API

All `/api/v1/coach/*` endpoints require JWT plus `coach`, `admin`, or `super_admin`. Dashboard and analytics are exposed at `/coach/dashboard` and `/coach/analytics`. Content collections support search/status filters; lesson creation/update validates JSON blocks. Lifecycle actions use `/coach/{kind}/{id}/publish`, `/unpublish`, `/archive`, and `/duplicate`. `/coach/media` manages external metadata and reference-safe deletion; binary upload reports a controlled unavailable response until object storage is configured.
