-- Permissions: todo
INSERT INTO permissions (name) VALUES
    ('todo:create'),
    ('todo:read'),
    ('todo:update'),
    ('todo:delete');

-- Assign todo permissions to admin
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name IN (
    'todo:create', 'todo:read', 'todo:update', 'todo:delete'
);

-- Assign all todo permissions to user
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'user' AND p.name IN (
    'todo:create', 'todo:read', 'todo:update', 'todo:delete'
);
