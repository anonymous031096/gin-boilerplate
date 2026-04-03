DELETE FROM users WHERE email = 'admin@example.com';
DELETE FROM role_permissions WHERE role_id IN (SELECT id FROM roles WHERE name IN ('admin', 'viewer'));
DELETE FROM roles WHERE name IN ('admin', 'user');
DELETE FROM permissions WHERE name IN (
    'user:create', 'user:read', 'user:update', 'user:delete',
    'role:create', 'role:read', 'role:update', 'role:delete'
);
