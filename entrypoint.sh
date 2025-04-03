#!/bin/sh

# exit immediately if a command exits with a non-zero status.
set -e

echo "Entrypoint: Verifying PostgreSQL connection to ${DB_HOST}:${DB_PORT}..."

# timeout
WAIT_TIMEOUT=60
SECONDS_WAITED=0

# loop until pg_isready returns success (exit code 0)
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -q; do
  if [ $SECONDS_WAITED -ge $WAIT_TIMEOUT ]; then
    echo "Error: Timed out after ${WAIT_TIMEOUT}s waiting for PostgreSQL at ${DB_HOST}:${DB_PORT}." >&2
    exit 1
  fi
  echo "Entrypoint: PostgreSQL not ready yet. Waiting 2 seconds..."
  sleep 2
  SECONDS_WAITED=$((SECONDS_WAITED + 2))
done

echo "Entrypoint: PostgreSQL is ready!"

# start application
echo "Entrypoint: Starting application (go-link-shortener)..."
echo "  PUBLIC_SITE_URL=${PUBLIC_SITE_URL}"
echo "  ENABLE_DOCS=${ENABLE_DOCS}"
echo "  SERVER_PORT=${SERVER_PORT}"

exec ./go-link-shortener
