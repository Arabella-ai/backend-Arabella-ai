#!/bin/bash
# Script to rebuild and restart the backend with DashScope integration

set -e

echo "🔄 Rebuilding and restarting backend with DashScope integration..."
echo ""

cd /var/www/arabella/backend

echo "📦 Building backend..."
go build -o /tmp/arabella-api ./cmd/api

if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo "✅ Build successful"
echo ""

echo "🔄 Stopping backend service..."
sudo systemctl stop arabella-api

echo "📋 Copying new binary..."
sudo cp /tmp/arabella-api /usr/local/bin/arabella-api || sudo cp /tmp/arabella-api /opt/arabella/arabella-api || {
    echo "⚠️  Could not determine binary location. Please copy manually:"
    echo "   sudo cp /tmp/arabella-api <path-to-binary>"
    echo ""
    echo "Then restart: sudo systemctl start arabella-api"
    exit 1
}

echo "🔄 Starting backend service..."
sudo systemctl start arabella-api

echo ""
echo "✅ Backend restarted!"
echo ""
echo "📋 Check status:"
echo "   sudo systemctl status arabella-api"
echo ""
echo "📋 View logs:"
echo "   sudo journalctl -u arabella-api -f"
echo ""
echo "🔍 Verify DashScope endpoint is being used:"
echo "   sudo journalctl -u arabella-api | grep 'DashScope' | tail -5"






