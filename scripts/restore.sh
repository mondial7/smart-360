#!/usr/bin/env bash
#
# Restore a Smart 360 PostgreSQL dump produced by scripts/backup.sh.
#
# Usage:
#   DATABASE_URL=postgres://user:pass@host:5432/smart360 \
#     ./scripts/restore.sh backups/smart360-YYYYMMDDTHHMMSSZ.dump
#
# WARNING: this drops and recreates the objects in the target database. Point it
# at a fresh/empty database for disaster recovery, or knowingly overwrite.
set -euo pipefail

: "${DATABASE_URL:?set DATABASE_URL to the target Postgres connection string}"
dump="${1:?usage: restore.sh <dump-file>}"

[ -f "$dump" ] || { echo "dump not found: $dump" >&2; exit 1; }
command -v pg_restore >/dev/null 2>&1 || {
  echo "pg_restore not found — install the PostgreSQL client tools." >&2
  exit 1
}

echo "About to restore ${dump}"
echo "  into ${DATABASE_URL%%\?*}"
echo "  (existing objects will be dropped and recreated). Ctrl-C within 5s to abort."
sleep 5

pg_restore --clean --if-exists --no-owner --no-privileges --dbname="$DATABASE_URL" "$dump"
echo "Restore complete. The app re-applies any pending migrations on next start."
