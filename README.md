# Calisthenics Coach

Telegram Mini App for learning calisthenics, completing workouts, and tracking progress. The project is built as a Go modular monolith with a React mobile-first client.

## Stack

- Go 1.24, Chi, PostgreSQL 16, Redis 7
- React, TypeScript, Vite, Tailwind CSS
- Docker Compose, OpenAPI

## Prerequisites

- Docker Engine with the Docker Compose plugin
- Go 1.24+ for local backend development
- Node.js 20.19+ (Node 24 is supported) and npm for local frontend development

## Configuration

Copy `.env.example` to `.env` and replace development values where appropriate. Do not commit `.env` or real secrets.

Required runtime variables are `DATABASE_URL`, `REDIS_URL`, and `JWT_SECRET`. To enable Telegram sign-in, set `TELEGRAM_BOT_TOKEN` to the bot token; the backend never accepts a Telegram identity that is not cryptographically verified in `initData`.

## Start with Docker

```powershell
docker-compose up --build
```

If your Docker installation exposes the Compose plugin, `docker compose up --build` is equivalent.

After startup:

- Frontend: `http://localhost:3000`
- Backend health: `http://localhost:8080/healthz`
- OpenAPI foundation: `http://localhost:8080/openapi.yaml`

The compose stack starts PostgreSQL, Redis, backend, and frontend. The backend starts only after PostgreSQL and Redis pass their health checks.

## Local development

Start PostgreSQL and Redis with Docker Compose, then set the variables from `.env.example` in your shell.

```powershell
Set-Location backend
go run ./cmd/server
```

In a second shell:

```powershell
Set-Location frontend
npm install
npm run dev
```

## Production

Use the separate production Compose file and a protected environment file; it keeps PostgreSQL and Redis off public ports and binds app services to loopback for an HTTPS reverse proxy.

```sh
cp .env.production.example .env.production
# Edit all placeholder values, then:
docker compose --env-file .env.production -f docker-compose.production.yml up -d postgres redis
docker compose --env-file .env.production -f docker-compose.production.yml --profile tools run --rm migrate
docker compose --env-file .env.production -f docker-compose.production.yml up -d --build backend frontend
```

See [deployment](docs/DEPLOYMENT.md), [backups](docs/BACKUPS.md), and [Telegram setup](docs/TELEGRAM.md) before exposing the app publicly.

## Database, seed data, and Telegram

The backend never creates schema implicitly. Apply the schema and demo seed data through the migration tool profile:

```powershell
docker-compose --profile tools run --rm migrate
```

The seed includes Russian lessons, exercises, programs, workouts, and achievements. Regenerate typed database queries with `Set-Location backend; sqlc generate` after changing schema or queries. Telegram authentication and bot setup are scheduled for a later phase.

## Checks

```powershell
Set-Location backend
go test ./...

Set-Location ../frontend
npm run build

# Docker smoke test (starts required services and applies migrations)
Set-Location ..
docker compose up -d postgres redis
docker compose --profile tools run --rm migrate
docker compose up -d --build backend frontend
Invoke-WebRequest http://localhost:8080/healthz
```

## Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Database design and workflow](docs/DATABASE.md)
- [Deployment](docs/DEPLOYMENT.md)
- [Backup and restore](docs/BACKUPS.md)
- [Telegram setup](docs/TELEGRAM.md)
- [OpenAPI foundation](backend/docs/openapi.yaml)
- [Master specification](CALISTHENICS_TELEGRAM_MINI_APP_MASTER_SPEC.md)
