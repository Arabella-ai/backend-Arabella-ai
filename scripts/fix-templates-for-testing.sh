#!/bin/bash
# Script to make templates non-premium for testing Wan AI

set -e

echo "🔧 Making templates non-premium for testing..."
echo ""

# Get database credentials from .env
DB_HOST=$(grep "^DB_HOST=" .env | cut -d '=' -f2 || echo "localhost")
DB_PORT=$(grep "^DB_PORT=" .env | cut -d '=' -f2 || echo "5432")
DB_USER=$(grep "^DB_USER=" .env | cut -d '=' -f2 || echo "postgres")
DB_NAME=$(grep "^DB_NAME=" .env | cut -d '=' -f2 || echo "arabella")
DB_PASSWORD=$(grep "^DB_PASSWORD=" .env | cut -d '=' -f2 || echo "")

if [ -z "$DB_PASSWORD" ]; then
    echo "⚠️  DB_PASSWORD not found in .env, you may need to enter it manually"
    echo ""
    echo "Run this SQL command manually:"
    echo "  UPDATE templates SET is_premium = false WHERE is_premium = true;"
    echo ""
    exit 1
fi

# Export password for psql
export PGPASSWORD="$DB_PASSWORD"

echo "📊 Current templates:"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT id, name, is_premium, is_active FROM templates;" 2>&1 || {
    echo "❌ Failed to connect to database"
    echo ""
    echo "Please run this SQL manually:"
    echo "  UPDATE templates SET is_premium = false WHERE is_premium = true;"
    exit 1
}

echo ""
echo "🔄 Updating templates to be non-premium..."
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "UPDATE templates SET is_premium = false WHERE is_premium = true;" 2>&1

echo ""
echo "✅ Updated templates:"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT id, name, is_premium, is_active FROM templates;" 2>&1

echo ""
echo "✅ Templates are now non-premium. You can test Wan AI video generation!"







