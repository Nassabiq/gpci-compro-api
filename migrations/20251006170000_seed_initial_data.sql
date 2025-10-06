-- +goose Up
WITH seed_users AS (
    SELECT
        LEFT(MD5(random()::text || clock_timestamp()::text || 'admin'), 12) AS xid,
        'Admin User'::TEXT AS name,
        'admin@example.com'::TEXT AS email,
        '$2a$10$L3hRRoLbNBXD8xfjXlKntOAww6ztlggdc79GaHuy5IXnGM/p0yyWq'::TEXT AS password,
        TRUE AS is_active,
        NOW() AS email_verified_at
    UNION ALL
    SELECT
        LEFT(MD5(random()::text || clock_timestamp()::text || 'editor'), 12),
        'Editor User',
        'editor@example.com',
        '$2a$10$L3hRRoLbNBXD8xfjXlKntOAww6ztlggdc79GaHuy5IXnGM/p0yyWq',
        TRUE,
        NOW()
    UNION ALL
    SELECT
        LEFT(MD5(random()::text || clock_timestamp()::text || 'disabled'), 12),
        'Disabled User',
        'disabled@example.com',
        '$2a$10$L3hRRoLbNBXD8xfjXlKntOAww6ztlggdc79GaHuy5IXnGM/p0yyWq',
        FALSE,
        NULL
)
INSERT INTO users (xid, name, email, password, is_active, email_verified_at)
SELECT xid, name, email, password, is_active, email_verified_at
FROM seed_users
ON CONFLICT (email) DO NOTHING;

INSERT INTO roles (name)
VALUES ('admin'), ('editor')
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (key, description)
VALUES
    ('rbac.roles.read', 'List roles'),
    ('rbac.roles.write', 'Create or update roles'),
    ('rbac.permissions.read', 'List permissions'),
    ('rbac.permissions.write', 'Create or update permissions'),
    ('rbac.roles.assign', 'Assign permissions to a role'),
    ('rbac.users.assign_role', 'Assign roles to a user')
ON CONFLICT (key) DO NOTHING;

-- Ensure admin role has all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.key IN (
    'rbac.roles.read',
    'rbac.roles.write',
    'rbac.permissions.read',
    'rbac.permissions.write',
    'rbac.roles.assign',
    'rbac.users.assign_role'
)
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

-- Assign editor a subset of permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.key IN ('rbac.roles.read', 'rbac.permissions.read', 'rbac.roles.assign', 'rbac.users.assign_role')
WHERE r.name = 'editor'
ON CONFLICT DO NOTHING;

-- Map users to roles
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
JOIN roles r ON r.name = 'admin'
WHERE u.email = 'admin@example.com'
ON CONFLICT DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
JOIN roles r ON r.name = 'editor'
WHERE u.email = 'editor@example.com'
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM user_roles
WHERE user_id IN (SELECT id FROM users WHERE email IN ('admin@example.com', 'editor@example.com', 'disabled@example.com'));

DELETE FROM role_permissions
WHERE role_id IN (SELECT id FROM roles WHERE name IN ('admin', 'editor'))
  AND permission_id IN (
      SELECT id FROM permissions
      WHERE key IN (
          'rbac.roles.read',
          'rbac.roles.write',
          'rbac.permissions.read',
          'rbac.permissions.write',
          'rbac.roles.assign',
          'rbac.users.assign_role'
      )
  );

DELETE FROM users WHERE email IN ('admin@example.com', 'editor@example.com', 'disabled@example.com');
DELETE FROM permissions WHERE key IN ('rbac.roles.read', 'rbac.roles.write', 'rbac.permissions.read', 'rbac.permissions.write', 'rbac.roles.assign', 'rbac.users.assign_role');
DELETE FROM roles WHERE name IN ('admin', 'editor');
