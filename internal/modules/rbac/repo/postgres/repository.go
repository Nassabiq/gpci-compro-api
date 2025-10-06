package postgres

import (
	"context"
	"database/sql"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/rbac/domain"
)

type RBACRepository struct{ DB *sql.DB }

func (r *RBACRepository) CreateRole(ctx context.Context, name string) (int64, error) {
	var id int64
	err := r.DB.QueryRowContext(ctx, `INSERT INTO roles(name) VALUES($1) RETURNING id`, name).Scan(&id)
	return id, err
}

func (r *RBACRepository) CreatePermission(ctx context.Context, key, description string) (int64, error) {
	var id int64
	err := r.DB.QueryRowContext(ctx, `INSERT INTO permissions(key, description) VALUES($1,$2) RETURNING id`, key, description).Scan(&id)
	return id, err
}

func (r *RBACRepository) ListRoles(ctx context.Context) ([]domain.Role, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, name, created_at, updated_at
		FROM roles
		WHERE deleted_at IS NULL
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *RBACRepository) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, key, description, created_at, updated_at
		FROM permissions
		WHERE deleted_at IS NULL
		ORDER BY key
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.Key, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *RBACRepository) AssignPermissionToRole(ctx context.Context, roleName, permKey string) error {
	var roleID, permID int64
	if err := r.DB.QueryRowContext(ctx, `SELECT id FROM roles WHERE name=$1 AND deleted_at IS NULL`, roleName).Scan(&roleID); err != nil {
		return err
	}
	if err := r.DB.QueryRowContext(ctx, `SELECT id FROM permissions WHERE key=$1 AND deleted_at IS NULL`, permKey).Scan(&permID); err != nil {
		return err
	}
	_, err := r.DB.ExecContext(ctx, `INSERT INTO role_permissions(role_id, permission_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, roleID, permID)
	return err
}

func (r *RBACRepository) AssignRoleToUserByXID(ctx context.Context, userXID, roleName string) error {
	var userID, roleID int64
	if err := r.DB.QueryRowContext(ctx, `SELECT id FROM users WHERE xid=$1 AND deleted_at IS NULL`, userXID).Scan(&userID); err != nil {
		return err
	}
	if err := r.DB.QueryRowContext(ctx, `SELECT id FROM roles WHERE name=$1 AND deleted_at IS NULL`, roleName).Scan(&roleID); err != nil {
		return err
	}
	_, err := r.DB.ExecContext(ctx, `INSERT INTO user_roles(user_id, role_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, userID, roleID)
	return err
}

func (r *RBACRepository) UserHasPermissionByXID(ctx context.Context, userXID, permKey string) (bool, error) {
	var ok bool
	// Admin role shortcut OR explicit permission via role
	err := r.DB.QueryRowContext(ctx, `
        SELECT EXISTS (
            SELECT 1
            FROM users u
            JOIN user_roles ur ON ur.user_id = u.id
            JOIN roles r ON r.id = ur.role_id AND r.deleted_at IS NULL
            LEFT JOIN role_permissions rp ON rp.role_id = r.id
            LEFT JOIN permissions p ON p.id = rp.permission_id AND p.deleted_at IS NULL
            WHERE u.xid = $1 AND u.deleted_at IS NULL
              AND (r.name = 'admin' OR p.key = $2)
        )
    `, userXID, permKey).Scan(&ok)
	return ok, err
}
