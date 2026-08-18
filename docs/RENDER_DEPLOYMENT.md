# Render deployment

This guide prepares the accepted MVP release candidate for a manual Render deployment. It does not replace the release acceptance process and does not authorize an automatic deployment.

## Blueprint architecture

`render.yaml` creates these resources:

- `calisthenics-web`: a React/Vite Render Static Site built with `npm ci && npm run build` and published from `frontend/dist`;
- `calisthenics-api`: a Docker Render Web Service built from `backend/Dockerfile`, with `/healthz` as its health check;
- `calisthenics-db`: managed Render Postgres, available to the backend through its private connection string;
- `calisthenics-key-value`: Render Key Value with no public IP allow list, available through its private connection string.

The backend, Postgres, and Key Value resources are pinned to `frankfurt`. If another region is required, change all three together before creating the Blueprint. Static sites are served through Render's global CDN and do not have a region setting.

The Go server binds to `:$PORT`, which is equivalent to `0.0.0.0:$PORT`; Render supplies `PORT` automatically. Do not add a fixed `PORT` value in the Dashboard.

## First deployment through the Dashboard

1. Sign in to Render and connect the GitHub account that can access `VadimAfonievich/calisthenics-web-app`.
2. Select **New > Blueprint**, select that repository, and use the root `render.yaml` file.
3. Review the Blueprint diff. Keep `calisthenics-api`, `calisthenics-db`, and `calisthenics-key-value` in the same region.
4. During initial Blueprint creation, enter the values requested for variables marked `sync: false`:
   - `JWT_SECRET`: a new cryptographically random production secret;
   - `TELEGRAM_BOT_TOKEN`: the token issued by BotFather;
   - `TELEGRAM_BOT_USERNAME`: the bot username without a repository-stored credential.
5. Confirm the non-secret backend configuration:
   - `APP_ENV=production`;
   - `TELEGRAM_WEBAPP_URL=https://calisthenics-web.onrender.com`;
   - `CORS_ORIGINS=https://calisthenics-web.onrender.com`;
   - `LOG_LEVEL=INFO`.
6. Confirm the frontend build variable `VITE_API_URL=https://calisthenics-api.onrender.com/api/v1`. Vite reads this at build time, so redeploy the static site after changing it.
7. Do not manually replace `DATABASE_URL` or `REDIS_URL`. The Blueprint injects the private Postgres and Key Value connection strings through Render resource references.

If Render assigns different service hostnames because either requested name is unavailable, update `VITE_API_URL`, `TELEGRAM_WEBAPP_URL`, and `CORS_ORIGINS` to the actual HTTPS URLs, then rebuild the static site and redeploy the backend configuration.

## Migrations and MVP seed

The backend image includes the existing `golang-migrate` executable and migration files solely for manual jobs. Its default command remains `/server`; migrations never run on backend restart, and a Render one-off job can safely override the default command.

After Postgres is ready and the backend image has built, open an ephemeral shell or start a one-off job based on `calisthenics-api`. Run:

```sh
/usr/local/bin/migrate -path=/migrations -database "$DATABASE_URL" up
```

Migration `000002_seed_demo_content` is the MVP seed. Consequently, the command above applies both schema migrations and the idempotent, version-tracked seed exactly once. There is no separate ad-hoc seed command. To stage the first empty database explicitly, the equivalent commands are:

```sh
/usr/local/bin/migrate -path=/migrations -database "$DATABASE_URL" up 1  # schema
/usr/local/bin/migrate -path=/migrations -database "$DATABASE_URL" up 1  # MVP seed
/usr/local/bin/migrate -path=/migrations -database "$DATABASE_URL" up    # remaining migrations
```

Do not force a migration version, run `down`, or rerun the seed SQL manually in production. Before later migrations, create a managed Postgres backup and follow `docs/BACKUPS.md`.

Render ephemeral shells and one-off jobs require an eligible service plan. If the selected plan does not provide them, use Render's Shell for the web service when available, or temporarily run the same `migrate/migrate:v4.18.3` command from an authorized machine using the database's external URL and a narrowly scoped temporary IP allow-list. Remove that allow-list immediately afterward; do not commit the external URL.

## Telegram configuration

1. In BotFather, set the bot Menu Button or Mini App URL to the final frontend HTTPS URL.
2. Ensure `TELEGRAM_WEBAPP_URL` exactly matches that URL.
3. Ensure `CORS_ORIGINS` contains only the frontend origin, without a path or trailing wildcard.
4. Open the app from Telegram and verify signed `initData`; a direct desktop browser does not provide production Telegram authentication.

## Verification

Verify the public endpoints after migrations complete:

```text
https://calisthenics-api.onrender.com/healthz
https://calisthenics-api.onrender.com/openapi.yaml
https://calisthenics-web.onrender.com
```

### Deployment verification checklist

Backend:

- [ ] Deploy completed successfully.
- [ ] `GET /healthz` returns 200.
- [ ] `GET /openapi.yaml` returns 200.
- [ ] PostgreSQL connectivity is healthy.
- [ ] Render Key Value connectivity is healthy.

Frontend:

- [ ] Production build completed successfully.
- [ ] The index page loads.
- [ ] Direct navigation and refresh on SPA routes load `index.html`.
- [ ] API requests reach `calisthenics-api` without CORS errors.

Database:

- [ ] All migrations are applied and the migration state is not dirty.
- [ ] MVP seed migration is applied.
- [ ] Lessons: 10 or more.
- [ ] Exercises: 10 or more.
- [ ] Programs: 2 or more.
- [ ] Workouts: 5 or more.

Telegram:

- [ ] Bot `/start` works.
- [ ] The Open App button works.
- [ ] The Mini App opens at the production frontend URL.
- [ ] Real Telegram `initData` is accepted.
- [ ] A user is created.
- [ ] Authenticated `GET /api/v1/me` works.

User flow:

- [ ] Programs load.
- [ ] Lessons load and can be completed.
- [ ] Today's workout loads.
- [ ] A workout session starts.
- [ ] Exercise sets can be recorded.
- [ ] Workout completion succeeds.
- [ ] XP is awarded once.
- [ ] Streak is updated.
- [ ] Achievements are evaluated.
- [ ] Progress and history update.

## Operational limitations

- Render's service filesystem is ephemeral. The MVP has no user-upload endpoint; lesson and exercise media are stored only as external `video_url` and `image_url` values. Do not add local uploads without object storage.
- Key Value stores rate-limit state and is rebuildable; PostgreSQL is the system of record.
- Database backups, retention, restore tests, monitoring, custom domains, and TLS/DNS ownership remain manual operational responsibilities.
- Render plan availability, one-off job access, sleep behavior, retention, and pricing depend on the selected account and plan; review them in the Dashboard before creation.
