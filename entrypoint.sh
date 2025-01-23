#!/bin/sh

# Wait for PostgreSQL
echo "Waiting for PostgreSQL at ${DB_HOST}:${DB_PORT}..."
# check on port 5434 and $DB_PORT
until pg_isready -h "$DB_HOST" -p 5434 -U "$DB_USER" -d "$DB_NAME" || pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME"; do
  sleep 2
done

# Start the application
echo "Starting application with PUBLIC_SITE_URL=${PUBLIC_SITE_URL}, ENABLE_DOCS=${ENABLE_DOCS}..."
exec ./go-link-shortener
