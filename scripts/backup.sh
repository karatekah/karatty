#!/usr/bin/env bash
set -euo pipefail

: "${DATABASE_URL:?set DATABASE_URL}"
: "${BACKUP_RECIPIENT:?set the age public recipient}"
: "${BACKUP_REMOTE:?set the rclone remote ending in /db-backups}"

if [[ "$BACKUP_RECIPIENT" != age1* || "$BACKUP_REMOTE" != *'/db-backups' ]]; then
  echo 'invalid backup recipient or remote' >&2
  exit 2
fi
for program in pg_dump age rclone; do
  command -v "$program" >/dev/null || { echo "missing $program" >&2; exit 2; }
done

umask 077
encrypted="$(mktemp "${BACKUP_TMPDIR:-${TMPDIR:-/tmp}}/termlink-XXXXXX.age")"
trap 'rm -f -- "$encrypted"' EXIT

pg_dump --format=custom --no-owner --no-acl --dbname="$DATABASE_URL" |
  age --recipient "$BACKUP_RECIPIENT" > "$encrypted"
test -s "$encrypted"

object="$(date -u +%Y%m%dT%H%M%SZ)-$$.dump.age"
rclone copyto "$encrypted" "$BACKUP_REMOTE/$object"

# A three-day window leaves margin below the seven-day maximum disclosed to users.
rclone delete "$BACKUP_REMOTE" --min-age 3d --include '*.dump.age'
stale="$(rclone lsf "$BACKUP_REMOTE" --min-age 4d --include '*.dump.age')"
if [[ -n "$stale" ]]; then
  echo 'old backup objects remain; check remote retention' >&2
  exit 1
fi
printf 'encrypted backup uploaded: %s/%s\n' "$BACKUP_REMOTE" "$object"
