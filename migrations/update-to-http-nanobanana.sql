-- Update all template thumbnails to use HTTP instead of HTTPS for nanobanana.uz
-- This avoids SSL certificate issues, but may cause mixed content warnings if site is HTTPS
UPDATE templates 
SET thumbnail_url = REPLACE(thumbnail_url, 'https://nanobanana.uz', 'http://nanobanana.uz')
WHERE thumbnail_url LIKE 'https://nanobanana.uz%';








