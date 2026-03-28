INSERT INTO permissions (id, code, name)
VALUES
    (gen_random_uuid(), 'users:read', 'Read users'),
    (gen_random_uuid(), 'users:read-me', 'Read self profile'),
    (gen_random_uuid(), 'users:update', 'Update users'),
    (gen_random_uuid(), 'roles:create', 'Create roles'),
    (gen_random_uuid(), 'roles:read', 'Read roles'),
    (gen_random_uuid(), 'roles:update', 'Update roles'),
    (gen_random_uuid(), 'roles:delete', 'Delete roles'),
    (gen_random_uuid(), 'permissions:read', 'Read permissions'),
    (gen_random_uuid(), 'products:create', 'Create products'),
    (gen_random_uuid(), 'products:read', 'Read products'),
    (gen_random_uuid(), 'products:update', 'Update products'),
    (gen_random_uuid(), 'products:delete', 'Delete products')
ON CONFLICT (code) DO NOTHING;

INSERT INTO roles (id, code, name)
VALUES
    (gen_random_uuid(), 'USER', 'User'),
    (gen_random_uuid(), 'ADMIN', 'Administrator')
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.code = 'users:read-me'
WHERE r.code = 'USER'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'ADMIN'
ON CONFLICT DO NOTHING;
