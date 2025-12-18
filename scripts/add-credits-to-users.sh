#!/bin/bash
# Script to add 10000 credits to all users for testing

set -e

echo "💰 Adding 10000 credits to all users..."
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
    echo "  UPDATE users SET credits = 10000;"
    echo ""
    exit 1
fi

# Export password for psql
export PGPASSWORD="$DB_PASSWORD"

echo "📊 Current user credits:"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT id, email, name, credits, tier FROM users;" 2>&1 || {
    echo "❌ Failed to connect to database"
    echo ""
    echo "Please run this SQL manually:"
    echo "  UPDATE users SET credits = 10000;"
    exit 1
}

echo ""
echo "🔄 Updating all users to have 10000 credits..."
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "UPDATE users SET credits = 10000;" 2>&1

echo ""
echo "✅ Updated user credits:"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT id, email, name, credits, tier FROM users;" 2>&1

echo ""
echo "✅ All users now have 10000 credits. You can test Wan AI video generation!"







