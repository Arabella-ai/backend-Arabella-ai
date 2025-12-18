#!/bin/bash
# Setup SSL for api.arabella.uz using Let's Encrypt

set -e

echo "🔒 Setting up SSL for api.arabella.uz..."

# Check if certbot is installed
if ! command -v certbot &> /dev/null; then
    echo "📦 Installing certbot..."
    sudo apt-get update
    sudo apt-get install -y certbot python3-certbot-nginx
fi

# Request SSL certificate
echo "📜 Requesting SSL certificate for api.arabella.uz..."
sudo certbot --nginx -d api.arabella.uz --non-interactive --agree-tos --email admin@arabella.uz --redirect

# Test certificate renewal
echo "🧪 Testing certificate renewal..."
sudo certbot renew --dry-run

echo "✅ SSL setup complete!"
echo ""
echo "📋 Next steps:"
echo "  1. Check Nginx config: sudo nginx -t"
echo "  2. Reload Nginx: sudo systemctl reload nginx"
echo "  3. Test SSL: curl -I https://api.arabella.uz/health"



