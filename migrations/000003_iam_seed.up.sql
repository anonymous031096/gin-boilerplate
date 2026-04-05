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

-- Permissions: permission
INSERT INTO permissions (name) VALUES
    ('permission:read');

-- Role: admin (is_system + is_superadmin, no need to seed permissions)
INSERT INTO roles (name, is_system, is_superadmin) VALUES ('admin', true, true);

-- Role: user (is_system + is_default)
INSERT INTO roles (name, is_system, is_default) VALUES ('user', true, true);

-- Assign basic permissions to default role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.is_default = true AND p.name IN ('user:read', 'role:read');

-- Admin user (password: Abc@1234)
INSERT INTO users (email, password, name) VALUES
    ('admin@init.com',
     '$2a$10$DsJ.NluXLiZ8gH0/k6cnROu6j2bZ5eenuLgJXw4MjcNZl9uQLMuDa',
     'Admin');

-- Assign superadmin role to admin user
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id FROM users u, roles r
WHERE u.email = 'admin@init.com' AND r.is_superadmin = true;
