#!/bin/bash
# Script to delete specific templates from the database

# Database connection (adjust if needed)
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-arabella}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"

# Templates to delete
TEMPLATES=(
  "Minimal Product Spin"
  "Luxury Reveal"
  "Golden Hour Walk"
  "Tech Desk Setup"
  "Morning Coffee"
  "Device Unboxing"
  "Ocean Waves"
  "Neon City Intro"
  "Mountain Sunrise"
  "Liquid Metal"
)

echo "🗑️  Deleting templates from database..."
echo ""

# Export password for psql
export PGPASSWORD="$DB_PASSWORD"

for template_name in "${TEMPLATES[@]}"; do
  echo "Deleting: $template_name"
  
  # Use soft delete (set is_active = false) - matches backend behavior
  psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c \
    "UPDATE templates SET is_active = false, updated_at = NOW() WHERE name = '$template_name';" 2>&1
  
  if [ $? -eq 0 ]; then
    echo "  ✅ Deleted: $template_name"
  else
    echo "  ❌ Failed to delete: $template_name"
  fi
done

# Or use hard delete (completely remove from database)
# Uncomment the following if you want to permanently delete instead of soft delete:
# echo ""
# echo "⚠️  Performing HARD DELETE (permanent removal)..."
# for template_name in "${TEMPLATES[@]}"; do
#   echo "Hard deleting: $template_name"
#   psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c \
#     "DELETE FROM templates WHERE name = '$template_name';" 2>&1
#   if [ $? -eq 0 ]; then
#     echo "  ✅ Hard deleted: $template_name"
#   else
#     echo "  ❌ Failed to hard delete: $template_name"
#   fi
# done

echo ""
echo "✅ Template deletion complete!"
echo ""
echo "📋 Verify deletion:"
echo "  psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c \"SELECT name, is_active FROM templates WHERE name IN ('Minimal Product Spin', 'Luxury Reveal', 'Golden Hour Walk', 'Tech Desk Setup', 'Morning Coffee', 'Device Unboxing', 'Ocean Waves', 'Neon City Intro', 'Mountain Sunrise', 'Liquid Metal');\""

