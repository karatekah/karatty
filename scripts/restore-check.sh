#!/usr/bin/env bash
set -euo pipefail

: "${BACKUP_OBJECT:?set the exact encrypted object path}"
: "${AGE_IDENTITY_FILE:?set the offline age identity file}"
: "${RESTORE_DATABASE_URL:?set an empty scratch PostgreSQL database URL}"

for program in psql pg_restore age rclone; do
  command -v "$program" >/dev/null || { echo "missing $program" >&2; exit 2; }
done
test -r "$AGE_IDENTITY_FILE"

database="$(psql "$RESTORE_DATABASE_URL" -Atqc 'SELECT current_database()')"
if [[ "$database" != termlink_restore_* ]]; then
  echo 'refusing to restore outside a termlink_restore_* scratch database' >&2
  exit 2
fi
tables="$(psql "$RESTORE_DATABASE_URL" -Atqc "SELECT count(*) FROM pg_tables WHERE schemaname='public'")"
if [[ "$tables" != 0 ]]; then
  echo 'scratch database is not empty' >&2
  exit 2
fi

rclone cat "$BACKUP_OBJECT" |
  age --decrypt --identity "$AGE_IDENTITY_FILE" |
  pg_restore --dbname="$RESTORE_DATABASE_URL" --no-owner --no-acl --single-transaction --exit-on-error

psql "$RESTORE_DATABASE_URL" -Atqc \
  "SELECT 'users='||count(*) FROM users; SELECT 'messages='||count(*) FROM messages;"
echo 'restore verified in scratch database'
