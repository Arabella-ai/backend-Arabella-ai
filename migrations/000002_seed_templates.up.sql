-- Seed sample templates
INSERT INTO templates (id, name, category, description, thumbnail_url, base_prompt, default_params, credit_cost, estimated_time_seconds, is_premium, tags)
VALUES
    -- Cyberpunk Category
    (
        'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
        'Neon City Intro',
        'cyberpunk_intro',
        'A stunning cyberpunk cityscape with neon lights, flying cars, and futuristic architecture. Perfect for tech and gaming content.',
        'https://nanobanana.uz/api/uploads/images/c85e3ffd-cffe-4483-b78a-2de212908e94.png',
        'Cinematic cyberpunk city at night with vibrant neon signs, holographic advertisements, and flying vehicles. Rain-soaked streets reflect the colorful lights. Ultra detailed, 4K quality.',
        '{"duration": 15, "resolution": "1080p", "aspect_ratio": "16:9", "fps": 30}',
        2,
        90,
        FALSE,
        ARRAY['cyberpunk', 'neon', 'city', 'intro', 'gaming']
    ),
    (
        'a1b2c3d4-e5f6-7890-abcd-ef1234567891',
        'Cyber Runner',
        'cyberpunk_intro',
        'A dynamic chase sequence through neon-lit alleyways with a cyber-enhanced protagonist.',
        'https://nanobanana.uz/api/uploads/images/c85e3ffd-cffe-4483-b78a-2de212908e94.png',
        'First-person perspective running through narrow cyberpunk alleyways, neon signs blurring past, holographic advertisements flickering. Intense action sequence.',
        '{"duration": 10, "resolution": "1080p", "aspect_ratio": "16:9", "fps": 60}',
        3,
        120,
        TRUE,
        ARRAY['cyberpunk', 'action', 'chase', 'first-person']
    ),

    -- Product Showcase Category
    (
        'b2c3d4e5-f6a7-8901-bcde-f12345678901',
        'Minimal Product Spin',
        'product_showcase',
        'Clean, minimal 360-degree product rotation with soft lighting and neutral background.',
        'https://nanobanana.uz/api/uploads/images/c85e3ffd-cffe-4483-b78a-2de212908e94.png',
        'Professional product photography style, clean white background, soft studio lighting, smooth 360-degree rotation, subtle reflections on surface. Minimal and elegant.',
        '{"duration": 8, "resolution": "1080p", "aspect_ratio": "1:1", "fps": 30}',
        1,
        45,
        FALSE,
        ARRAY['product', 'minimal', 'ecommerce', 'showcase']
    ),
    (
        'b2c3d4e5-f6a7-8901-bcde-f12345678902',
        'Luxury Reveal',
        'product_showcase',
        'Premium product reveal with dramatic lighting, particle effects, and cinematic camera movements.',
        'https://nanobanana.uz/api/uploads/images/c85e3ffd-cffe-4483-b78a-2de212908e94.png',
        'Luxury product reveal with dramatic spotlight, golden particle effects, slow-motion unveiling, premium black velvet background. Cinematic camera sweep.',
        '{"duration": 12, "resolution": "4k", "aspect_ratio": "16:9", "fps": 30}',
        4,
        150,
        TRUE,
        ARRAY['product', 'luxury', 'reveal', 'premium', 'cinematic']
    ),

    -- Daily Vlog Category
    (
        'c3d4e5f6-a7b8-9012-cdef-123456789012',
        'Morning Coffee',
        'daily_vlog',
        'Cozy morning atmosphere with steam rising from a coffee cup, soft natural lighting.',
        'https://nanobanana.uz/api/uploads/images/c85e3ffd-cffe-4483-b78a-2de212908e94.png',
        'Cozy morning scene, steam rising from a fresh cup of coffee, soft golden hour sunlight streaming through window, peaceful and warm atmosphere. Lifestyle aesthetic.',
        '{"duration": 10, "resolution": "1080p", "aspect_ratio": "9:16", "fps": 30}',
        1,
        60,
        FALSE,
        ARRAY['vlog', 'lifestyle', 'morning', 'cozy', 'coffee']
    ),
    (
        'c3d4e5f6-a7b8-9012-cdef-123456789013',
        'Golden Hour Walk',
        'daily_vlog',
        'Serene walking footage during golden hour with lens flares and dreamy atmosphere.',
        'https://nanobanana.uz/api/uploads/images/c85e3ffd-cffe-4483-b78a-2de212908e94.png',
        'Point-of-view walking through nature during golden hour, lens flares, dreamy atmosphere, tall grass swaying gently. Peaceful and contemplative mood.',
        '{"duration": 15, "resolution": "1080p", "aspect_ratio": "9:16", "fps": 30}',
        1,
        60,
        FALSE,
        ARRAY['vlog', 'nature', 'golden-hour', 'walking', 'peaceful']
    ),

    -- Tech Review Category
    (
        'd4e5f6a7-b8c9-0123-defa-234567890123',
        'Tech Desk Setup',
        'tech_review',
        'Modern desk setup with ambient RGB lighting, multiple monitors, and tech gadgets.',
        'https://nanobanana.uz/api/uploads/images/c85e3ffd-cffe-4483-b78a-2de212908e94.png',
        'Modern tech workstation with multiple monitors, RGB ambient lighting, mechanical keyboard, gaming peripherals. Clean cable management, minimalist aesthetic.',
        '{"duration": 12, "resolution": "1080p", "aspect_ratio": "16:9", "fps": 30}',
        2,
        75,
        FALSE,
        ARRAY['tech', 'setup', 'desk', 'gaming', 'workspace']
    ),
    (
        'd4e5f6a7-b8c9-0123-defa-234567890124',
        'Device Unboxing',
        'tech_review',
        'Premium unboxing experience with ASMR-style close-ups and satisfying reveals.',
        'https://nanobanana.uz/api/uploads/images/c85e3ffd-cffe-4483-b78a-2de212908e94.png',
        'Premium tech unboxing experience, close-up shots of packaging being opened, hands revealing new device, satisfying textures and sounds, soft lighting.',
        '{"duration": 20, "resolution": "4k", "aspect_ratio": "16:9", "fps": 30}',
        3,
        120,
        TRUE,
        ARRAY['tech', 'unboxing', 'review', 'ASMR', 'premium']
    ),

    -- Nature Category
    (
        'e5f6a7b8-c9d0-1234-efab-345678901234',
        'Ocean Waves',
        'nature',
        'Calming ocean waves crashing on a sandy beach during sunset.',
        'https://nanobanana.uz/api/uploads/images/c85e3ffd-cffe-4483-b78a-2de212908e94.png',
        'Peaceful ocean waves gently crashing on pristine sandy beach, golden sunset colors reflecting on water, seagulls in distance. Relaxing and meditative.',
        '{"duration": 30, "resolution": "1080p", "aspect_ratio": "16:9", "fps": 30}',
        1,
        90,
        FALSE,
        ARRAY['nature', 'ocean', 'beach', 'sunset', 'relaxing']
    ),
    (
        'e5f6a7b8-c9d0-1234-efab-345678901235',
        'Mountain Sunrise',
        'nature',
        'Breathtaking time-lapse of sunrise over mountain peaks with clouds rolling through valleys.',
        'https://nanobanana.uz/api/uploads/images/c85e3ffd-cffe-4483-b78a-2de212908e94.png',
        'Epic mountain sunrise time-lapse, golden light illuminating snow-capped peaks, clouds flowing through valleys like rivers, dramatic sky colors. Cinematic and awe-inspiring.',
        '{"duration": 15, "resolution": "4k", "aspect_ratio": "16:9", "fps": 30}',
        3,
        180,
        TRUE,
        ARRAY['nature', 'mountain', 'sunrise', 'timelapse', 'epic']
    ),

    -- Abstract Category
    (
        'f6a7b8c9-d0e1-2345-fabc-456789012345',
        'Liquid Metal',
        'abstract',
        'Mesmerizing fluid simulation with metallic chrome-like surface and rainbow reflections.',
        'https://nanobanana.uz/api/uploads/images/c85e3ffd-cffe-4483-b78a-2de212908e94.png',
        'Abstract liquid metal simulation, chrome-like reflective surface, iridescent rainbow colors, smooth flowing motion, hypnotic and mesmerizing. Perfect loop.',
        '{"duration": 10, "resolution": "1080p", "aspect_ratio": "1:1", "fps": 60}',
        2,
        60,
        FALSE,
        ARRAY['abstract', 'liquid', 'metal', 'chrome', 'hypnotic']
    ),
    (
        'f6a7b8c9-d0e1-2345-fabc-456789012346',
        'Particle Universe',
        'abstract',
        'Stunning particle system forming cosmic patterns and nebula-like structures.',
        'https://nanobanana.uz/api/uploads/images/c85e3ffd-cffe-4483-b78a-2de212908e94.png',
        'Abstract particle simulation forming cosmic nebula patterns, deep space colors, billions of glowing particles, formation and dissolution of star systems. Ethereal and cosmic.',
        '{"duration": 20, "resolution": "4k", "aspect_ratio": "16:9", "fps": 30}',
        4,
        240,
        TRUE,
        ARRAY['abstract', 'particles', 'cosmic', 'nebula', 'space']
    );

