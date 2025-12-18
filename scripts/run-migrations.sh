#!/bin/bash
# Run database migrations

set -e

cd /var/www/arabella/backend

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Default values
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-arabella}
DB_PASSWORD=${DB_PASSWORD:-arabella_secret}
DB_NAME=${DB_NAME:-arabella}

# Check if migrate is installed
if ! command -v migrate &> /dev/null; then
    echo "❌ migrate tool not found. Installing..."
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    export PATH=$PATH:$(go env GOPATH)/bin
fi

# Database URL
DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

echo "🔄 Running database migrations..."
echo "Database: ${DB_HOST}:${DB_PORT}/${DB_NAME}"
echo ""

# Run migrations
migrate -path ../database/migrations -database "$DATABASE_URL" up

echo ""
echo "✅ Migrations completed successfully!"


