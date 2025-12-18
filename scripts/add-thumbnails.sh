#!/bin/bash
# Helper script to add thumbnail images

THUMBNAILS_DIR="/var/www/arabella/backend/static/thumbnails"

echo "📸 Thumbnail Image Manager"
echo "=========================="
echo ""

# Check if directory exists
if [ ! -d "$THUMBNAILS_DIR" ]; then
    echo "Creating directory: $THUMBNAILS_DIR"
    mkdir -p "$THUMBNAILS_DIR"
fi

echo "Current thumbnails:"
ls -lh "$THUMBNAILS_DIR" 2>/dev/null | grep -v "^total" || echo "  (empty)"

echo ""
echo "To add a thumbnail image:"
echo "  1. Copy your image file to: $THUMBNAILS_DIR"
echo "  2. Name it according to the template (e.g., cyber-runner.jpg)"
echo "  3. Ensure it's 400x600 pixels and JPEG format"
echo ""
echo "Example:"
echo "  cp /path/to/image.jpg $THUMBNAILS_DIR/cyber-runner.jpg"
echo "  chmod 644 $THUMBNAILS_DIR/cyber-runner.jpg"
echo "  chown pro:pro $THUMBNAILS_DIR/cyber-runner.jpg"
echo ""
echo "Required files:"
echo "  - cyber-runner.jpg"
echo "  - neon-city-intro.jpg"
echo "  - minimal-product-spin.jpg"
echo "  - luxury-reveal.jpg"
echo "  - morning-coffee.jpg"
echo "  - sunset-drive.jpg"
echo "  - tech-review-setup.jpg"
echo "  - device-unboxing.jpg"
echo "  - golden-hour-walk.jpg"
echo "  - liquid-metal.jpg"
echo "  - mountain-sunrise.jpg"
echo "  - ocean-waves.jpg"
echo "  - particle-universe.jpg"
echo "  - tech-desk-setup.jpg"


