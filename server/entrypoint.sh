#!/bin/sh
set -e

echo "==> Waiting for PostgreSQL..."
until pg_isready -h postgres -U mathlib -d mathlib > /dev/null 2>&1; do
  echo "    PostgreSQL is unavailable - sleeping"
  sleep 2
done
echo "    PostgreSQL is up!"

echo "==> Running database migrations..."
go run github.com/pressly/goose/v3/cmd/goose@v3.26.0 \
  -dir migrations \
  postgres "$DATABASE_URL" \
  up

if [ "$AUTO_SEED" = "true" ]; then
  echo "==> Seeding database..."
  go run ./cmd/mathlib seed
fi

echo "==> Starting MathLib server..."
exec go run ./cmd/mathlib serve
