#!/bin/sh
# Daily PostgreSQL backup with optional S3/MinIO upload.
# Cron example: 0 2 * * * /opt/optistock/scripts/backup.sh >> /var/log/optistock-backup.log 2>&1

set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
ENV_FILE="${ENV_FILE:-$ROOT_DIR/.env.prod}"

if [ -f "$ENV_FILE" ]; then
  set -a
  # shellcheck disable=SC1090
  . "$ENV_FILE"
  set +a
fi

TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
BACKUP_FILE="optistock_${TIMESTAMP}.sql.gz"
TMP_FILE="/tmp/${BACKUP_FILE}"

PGHOST="${POSTGRES_HOST:-postgres}"
PGPORT="${POSTGRES_PORT:-5432}"
PGUSER="${POSTGRES_USER:-optistock}"
PGDATABASE="${POSTGRES_DB:-optistock}"
export PGPASSWORD="${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"

echo "[$(date -Iseconds)] Starting backup ${BACKUP_FILE}"

pg_dump -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" | gzip > "$TMP_FILE"

if [ "${S3_UPLOAD_ENABLED:-false}" = "true" ]; then
  OBJECT_KEY="backups/$(date +%Y-%m-%d)/${BACKUP_FILE}"
  URL="${S3_ENDPOINT%/}/${S3_BUCKET}/${OBJECT_KEY}"
  curl -sfS -X PUT \
    -u "${S3_ACCESS_KEY}:${S3_SECRET_KEY}" \
    -H "Content-Type: application/gzip" \
    --data-binary "@${TMP_FILE}" \
    "$URL"
  echo "[$(date -Iseconds)] Uploaded to ${URL}"
else
  LOCAL_DIR="${BACKUP_LOCAL_DIR:-$ROOT_DIR/backups}"
  mkdir -p "$LOCAL_DIR"
  mv "$TMP_FILE" "${LOCAL_DIR}/${BACKUP_FILE}"
  echo "[$(date -Iseconds)] Saved to ${LOCAL_DIR}/${BACKUP_FILE}"
  exit 0
fi

rm -f "$TMP_FILE"
echo "[$(date -Iseconds)] Backup complete"
