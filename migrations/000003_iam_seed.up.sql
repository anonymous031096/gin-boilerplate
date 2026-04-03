-- Permissions: user
INSERT INTO permissions (name) VALUES
    ('user:create'),
    ('user:read'),
    ('user:update'),
    ('user:delete');

-- Permissions: role
INSERT INTO permissions (name) VALUES
    ('role:create'),
    ('role:read'),
    ('role:update'),
    ('role:delete');

-- Role: admin
INSERT INTO roles (name) VALUES ('admin');

-- Assign all permissions to admin
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name IN (
    'user:create', 'user:read', 'user:update', 'user:delete',
    'role:create', 'role:read', 'role:update', 'role:delete'
);

-- Role: user
INSERT INTO roles (name) VALUES ('user');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'user' AND p.name IN ('user:read', 'role:read');

-- Admin user (password: Abc@1234)
INSERT INTO users (email, password, name) VALUES
    ('admin@example.com',
     '$2a$10$DsJ.NluXLiZ8gH0/k6cnROu6j2bZ5eenuLgJXw4MjcNZl9uQLMuDa',
     'Admin');

-- Assign admin role to admin user
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id FROM users u, roles r
WHERE u.email = 'admin@example.com' AND r.name = 'admin';
