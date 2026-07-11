#!/usr/bin/env bash
#
# Back up the Smart 360 PostgreSQL database to a timestamped, compressed
# custom-format dump, then prune old dumps.
#
# Usage:
#   DATABASE_URL=postgres://user:pass@host:5432/smart360 ./scripts/backup.sh
#
# Env:
#   DATABASE_URL     required — the Postgres connection string
#   BACKUP_DIR       output directory (default: ./backups)
#   RETENTION_DAYS   prune dumps older than this many days (default: 14)
#
# Restore a dump with scripts/restore.sh. Ship dumps off-host (they contain all
# feedback data) and test a restore periodically — an untested backup is a guess.
set -euo pipefail

: "${DATABASE_URL:?set DATABASE_URL to the Postgres connection string}"
BACKUP_DIR="${BACKUP_DIR:-./backups}"
RETENTION_DAYS="${RETENTION_DAYS:-14}"

command -v pg_dump >/dev/null 2>&1 || {
  echo "pg_dump not found — install the PostgreSQL client tools." >&2
  exit 1
}

mkdir -p "$BACKUP_DIR"
ts="$(date -u +%Y%m%dT%H%M%SZ)"
out="$BACKUP_DIR/smart360-${ts}.dump"

echo "Backing up to ${out}"
pg_dump --format=custom --no-owner --no-privileges --file="$out" "$DATABASE_URL"
echo "Wrote ${out} ($(du -h "$out" | cut -f1))"

# Prune old dumps (best-effort).
find "$BACKUP_DIR" -maxdepth 1 -name 'smart360-*.dump' -type f -mtime "+${RETENTION_DAYS}" -print -delete \
  | sed 's/^/Pruned: /' || true
