-- Remove seeded templates
DELETE FROM templates WHERE id IN (
    'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    'a1b2c3d4-e5f6-7890-abcd-ef1234567891',
    'b2c3d4e5-f6a7-8901-bcde-f12345678901',
    'b2c3d4e5-f6a7-8901-bcde-f12345678902',
    'c3d4e5f6-a7b8-9012-cdef-123456789012',
    'c3d4e5f6-a7b8-9012-cdef-123456789013',
    'd4e5f6a7-b8c9-0123-defa-234567890123',
    'd4e5f6a7-b8c9-0123-defa-234567890124',
    'e5f6a7b8-c9d0-1234-efab-345678901234',
    'e5f6a7b8-c9d0-1234-efab-345678901235',
    'f6a7b8c9-d0e1-2345-fabc-456789012345',
    'f6a7b8c9-d0e1-2345-fabc-456789012346'
);

