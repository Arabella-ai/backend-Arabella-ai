#!/bin/bash
# Deploy DashScope integration - rebuild and restart backend

set -e

echo "🚀 Deploying DashScope integration..."
echo ""

cd /var/www/arabella/backend

echo "📦 Building backend with DashScope integration..."
go build -o bin/api ./cmd/api

if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo "✅ Build successful"
echo ""

echo "📋 Binary location: /var/www/arabella/backend/bin/api"
ls -lh bin/api
echo ""

echo "🔄 Restarting backend service..."
sudo systemctl restart arabella-api

echo ""
echo "⏳ Waiting 2 seconds for service to start..."
sleep 2

echo ""
echo "📊 Service status:"
sudo systemctl status arabella-api --no-pager | head -15

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📋 To verify DashScope is being used, check logs:"
echo "   sudo journalctl -u arabella-api -f | grep -i dashscope"






