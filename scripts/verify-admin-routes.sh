#!/bin/bash
# Verify admin routes are registered

echo "Checking if backend service is running..."
if systemctl is-active --quiet arabella-api; then
    echo "✅ Backend service is running"
else
    echo "❌ Backend service is not running"
    exit 1
fi

echo ""
echo "Checking route registration in code..."
if grep -q "adminTemplateRoutes.PUT" /var/www/arabella/backend/cmd/api/main.go; then
    echo "✅ PUT route found in code"
else
    echo "❌ PUT route NOT found in code"
    exit 1
fi

echo ""
echo "To fix 404 error, restart the backend service:"
echo "  sudo systemctl restart arabella-api"
echo ""
echo "Then check logs:"
echo "  sudo journalctl -u arabella-api -f"

