#!/bin/bash
# Update backend binary and restart service

echo "🛑 Stopping backend service..."
sudo systemctl stop arabella-api

echo "📦 Copying new binary..."
cp /var/www/arabella/backend/arabella-api /var/www/arabella/backend/bin/api

echo "🔄 Restarting backend service..."
sudo systemctl start arabella-api

sleep 3

echo ""
echo "📊 Service Status:"
sudo systemctl status arabella-api --no-pager | head -10

echo ""
echo "🧪 Testing proxy endpoint..."
curl -s http://localhost:8112/api/v1/proxy/image?url=https%3A%2F%2Fnanobanana.uz%2Fapi%2Fuploads%2Fimages%2Fc85e3ffd-cffe-4483-b78a-2de212908e94.png -I | head -3

echo ""
echo "✅ Done!"

