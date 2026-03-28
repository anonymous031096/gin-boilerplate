DELETE FROM role_permissions
WHERE role_id IN (SELECT id FROM roles WHERE code IN ('USER', 'ADMIN'));

DELETE FROM roles WHERE code IN ('USER', 'ADMIN');

DELETE FROM permissions
WHERE code IN (
    'users:read',
    'users:read-me',
    'users:update',
    'roles:create',
    'roles:read',
    'roles:update',
    'roles:delete',
    'permissions:read',
    'products:create',
    'products:read',
    'products:update',
    'products:delete'
);
