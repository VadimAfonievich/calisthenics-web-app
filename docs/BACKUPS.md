# Backup and restore

Back up PostgreSQL daily and before every migration. Redis holds rate-limit state and can be rebuilt; PostgreSQL is the source of truth.

Create a compressed logical backup from the production host:

```sh
docker compose -f docker-compose.production.yml exec -T postgres pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" | gzip > "calisthenics-$(date +%F).sql.gz"
```

Encrypt backups at rest, copy them to separate storage, retain at least 30 daily copies, and test restoration regularly. To restore, stop application traffic, create an empty target database, then run `gunzip -c backup.sql.gz | docker compose -f docker-compose.production.yml exec -T postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"`. Run migrations afterward and verify `/healthz`.
