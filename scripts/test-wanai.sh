#!/bin/bash
# Test script for Wan AI integration

set -e

echo "🧪 Testing Wan AI Integration"
echo ""

# Check if API key is set
if [ -z "$WANAI_API_KEY" ]; then
    WANAI_API_KEY=$(grep WANAI_API_KEY .env | cut -d '=' -f2)
fi

if [ -z "$WANAI_API_KEY" ]; then
    echo "❌ WANAI_API_KEY not found in .env"
    exit 1
fi

echo "✅ API Key found: ${WANAI_API_KEY:0:10}..."
echo ""

# Test API connection
echo "🔍 Testing Wan AI API connection..."
response=$(curl -s -w "\n%{http_code}" \
    -H "Authorization: Bearer $WANAI_API_KEY" \
    -H "Content-Type: application/json" \
    "https://api.wanai.dev/v1/videos?limit=1")

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

echo "HTTP Status: $http_code"
echo "Response: $body"
echo ""

if [ "$http_code" = "200" ] || [ "$http_code" = "401" ]; then
    echo "✅ API is reachable"
else
    echo "⚠️  Unexpected response code: $http_code"
fi

echo ""
echo "📋 Next steps:"
echo "1. Restart the backend service: sudo systemctl restart arabella-api"
echo "2. Test video generation through the API"
echo "3. Check logs: sudo journalctl -u arabella-api -f"







