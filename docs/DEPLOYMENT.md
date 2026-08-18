# Production deployment

Deploy on a host with Docker Engine, Docker Compose, a firewall, and a reverse proxy that terminates TLS. Keep ports 8080 and 3000 bound to loopback; expose only the reverse proxy on ports 80 and 443.

1. Copy `.env.example` to a protected production environment file and set strong, unique `POSTGRES_PASSWORD` and `JWT_SECRET` values. Use `https://` values for `TELEGRAM_WEBAPP_URL` and `CORS_ORIGINS`.
2. Set `DATABASE_URL` to the internal PostgreSQL service and `REDIS_URL` to the internal Redis service. Do not publish PostgreSQL or Redis ports.
3. Build and start infrastructure: `docker compose -f docker-compose.production.yml up -d postgres redis`.
4. Apply migrations once per deployment: `docker compose -f docker-compose.production.yml --profile tools run --rm migrate`.
5. Start application services: `docker compose -f docker-compose.production.yml up -d --build backend frontend`.
6. Configure the reverse proxy to serve the frontend and proxy `/api/` to `http://127.0.0.1:8080`. Redirect HTTP to HTTPS and forward `Host` and `X-Forwarded-Proto`.

Verify with `curl -fsS http://127.0.0.1:8080/healthz` and inspect JSON logs using `docker compose -f docker-compose.production.yml logs -f backend`.

The health endpoint checks PostgreSQL and Redis. Alert on non-200 health responses, repeated container restarts, database-volume capacity, and elevated 5xx responses in the reverse-proxy logs.
