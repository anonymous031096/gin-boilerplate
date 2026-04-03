DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE name IN (
        'todo:create', 'todo:read', 'todo:update', 'todo:delete'
    )
);
DELETE FROM permissions WHERE name IN (
    'todo:create', 'todo:read', 'todo:update', 'todo:delete'
);
