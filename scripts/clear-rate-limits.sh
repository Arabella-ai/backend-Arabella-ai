#!/bin/bash
# Script to clear rate limits and restart backend

set -e

echo "🔄 Clearing rate limits and restarting backend..."
echo ""

# Get Redis credentials from .env
REDIS_HOST=$(grep "^REDIS_HOST=" .env | cut -d '=' -f2 || echo "localhost")
REDIS_PORT=$(grep "^REDIS_PORT=" .env | cut -d '=' -f2 || echo "6379")
REDIS_PASSWORD=$(grep "^REDIS_PASSWORD=" .env | cut -d '=' -f2 || echo "")

if [ -z "$REDIS_PASSWORD" ]; then
    REDIS_CMD="redis-cli -h $REDIS_HOST -p $REDIS_PORT"
else
    REDIS_CMD="redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD"
fi

echo "🗑️  Clearing rate limit keys from Redis..."
$REDIS_CMD --scan --pattern "ratelimit:*" | xargs -r $REDIS_CMD DEL 2>/dev/null || echo "No rate limit keys found or Redis not accessible"

echo ""
echo "✅ Rate limits cleared"
echo ""
echo "🔄 Restarting backend service..."
sudo systemctl restart arabella-api

echo ""
echo "✅ Backend restarted"
echo ""
echo "📋 Check status:"
echo "  sudo systemctl status arabella-api"
echo ""
echo "📋 View logs:"
echo "  sudo journalctl -u arabella-api -f"







