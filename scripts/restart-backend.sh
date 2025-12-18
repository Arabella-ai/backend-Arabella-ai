#!/bin/bash
# Quick script to rebuild and restart the backend service

echo "🔨 Rebuilding backend..."
cd /var/www/arabella/backend

# Build the application
go build -o arabella-api ./cmd/api/main.go

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo ""
    echo "🔄 Restarting backend service..."
    sudo systemctl restart arabella-api
    
    sleep 2
    
    echo ""
    echo "📊 Service Status:"
    sudo systemctl status arabella-api --no-pager | head -10
    
    echo ""
    echo "✅ Backend restarted!"
    echo ""
    echo "💡 Test the proxy endpoint:"
    echo "   curl 'https://arabella.uz/api/v1/proxy/image?url=https%3A%2F%2Fnanobanana.uz%2Fapi%2Fuploads%2Fimages%2Fc85e3ffd-cffe-4483-b78a-2de212908e94.png' -I"
else
    echo "❌ Build failed! Check errors above."
    exit 1
fi

