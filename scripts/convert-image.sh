#!/bin/bash
# Convert image.png to proper thumbnail format

IMAGE_FILE="/var/www/arabella/backend/static/thumbnails/image.png"
THUMBNAILS_DIR="/var/www/arabella/backend/static/thumbnails"

if [ ! -f "$IMAGE_FILE" ]; then
    echo "❌ image.png not found at $IMAGE_FILE"
    exit 1
fi

echo "📸 Converting image.png to thumbnails..."
echo ""

# Check if ImageMagick is available
if command -v convert &> /dev/null; then
    echo "Using ImageMagick to convert..."
    
    # List of required thumbnails
    thumbnails=(
        "cyber-runner.jpg"
        "neon-city-intro.jpg"
        "minimal-product-spin.jpg"
        "luxury-reveal.jpg"
        "morning-coffee.jpg"
        "sunset-drive.jpg"
        "tech-review-setup.jpg"
        "device-unboxing.jpg"
        "golden-hour-walk.jpg"
        "liquid-metal.jpg"
        "mountain-sunrise.jpg"
        "ocean-waves.jpg"
        "particle-universe.jpg"
        "tech-desk-setup.jpg"
    )
    
    for thumb in "${thumbnails[@]}"; do
        output="$THUMBNAILS_DIR/$thumb"
        # Resize to 400x600 and convert to JPEG
        convert "$IMAGE_FILE" -resize 400x600^ -gravity center -extent 400x600 -quality 85 "$output"
        echo "✅ Created: $thumb"
    done
    
elif command -v ffmpeg &> /dev/null; then
    echo "Using ffmpeg to convert..."
    # FFmpeg can also convert images
    for thumb in "${thumbnails[@]}"; do
        output="$THUMBNAILS_DIR/$thumb"
        ffmpeg -i "$IMAGE_FILE" -vf "scale=400:600:force_original_aspect_ratio=increase,crop=400:600" -q:v 2 "$output" -y 2>/dev/null
        echo "✅ Created: $thumb"
    done
else
    echo "⚠️  No image conversion tool found (ImageMagick or ffmpeg)"
    echo ""
    echo "To convert manually, you can:"
    echo "  1. Install ImageMagick: sudo apt install imagemagick"
    echo "  2. Or use online tools to convert PNG to JPG and resize to 400x600"
    echo "  3. Then copy to: $THUMBNAILS_DIR/"
    exit 1
fi

echo ""
echo "✅ Conversion complete!"
echo ""
echo "📋 Verify images:"
echo "  ls -lh $THUMBNAILS_DIR/*.jpg"


