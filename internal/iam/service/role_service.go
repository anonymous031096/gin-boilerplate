package service

import (
	"context"
	"database/sql"

	"gin-boilerplate/internal/iam/dto"
	"gin-boilerplate/pkg/middleware"
	"gin-boilerplate/pkg/response"

	"github.com/redis/go-redis/v9"
)

type RoleService struct {
	db    *sql.DB
	redis *redis.Client
}

func NewRoleService(db *sql.DB, redis *redis.Client) *RoleService {
	return &RoleService{db: db, redis: redis}
}

func (s *RoleService) GetByID(ctx context.Context, id string) (dto.RoleDetailResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT r.id, r.name, r.is_system, r.is_superadmin, r.is_default,
		        p.id::text AS perm_id, p.name AS perm_name
		 FROM roles r
		 LEFT JOIN LATERAL (
		     SELECT p2.id, p2.name FROM permissions p2
		     WHERE CASE
		         WHEN r.is_superadmin THEN true
		         ELSE p2.id IN (SELECT rp.permission_id FROM role_permissions rp WHERE rp.role_id = r.id)
		     END
		     ORDER BY p2.name
		 ) p ON true
		 WHERE r.id = $1`,
		id,
	)
	if err != nil {
		return dto.RoleDetailResponse{}, err
	}
	defer rows.Close()

	var role dto.RoleDetailResponse
	role.Permissions = make([]dto.PermissionResponse, 0)
	var found bool
	for rows.Next() {
		var permID, permName string
		if err := rows.Scan(&role.ID, &role.Name, &role.IsSystem, &role.IsSuperadmin, &role.IsDefault, &permID, &permName); err != nil {
			return dto.RoleDetailResponse{}, err
		}
		found = true
		if permID != "" {
			role.Permissions = append(role.Permissions, dto.PermissionResponse{ID: permID, Name: permName})
		}
	}

	if !found {
		return dto.RoleDetailResponse{}, sql.ErrNoRows
	}

	return role, nil
}

func (s *RoleService) List(ctx context.Context) ([]dto.RoleResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, is_system, is_superadmin, is_default FROM roles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]dto.RoleResponse, 0)
	for rows.Next() {
		var role dto.RoleResponse
		if err := rows.Scan(&role.ID, &role.Name, &role.IsSystem, &role.IsSuperadmin, &role.IsDefault); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (s *RoleService) Create(ctx context.Context, req dto.CreateRoleRequest, createdBy string) (dto.RoleDetailResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return dto.RoleDetailResponse{}, err
	}
	defer tx.Rollback()

	var roleID string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO roles (name, created_by, updated_by) VALUES ($1, $2, $2) RETURNING id`,
		req.Name, createdBy,
	).Scan(&roleID)
	if err != nil {
		return dto.RoleDetailResponse{}, err
	}

	for _, permID := range req.PermissionIDs {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`,
			roleID, permID,
		)
		if err != nil {
			return dto.RoleDetailResponse{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return dto.RoleDetailResponse{}, err
	}

	return s.GetByID(ctx, roleID)
}

func (s *RoleService) Update(ctx context.Context, id string, req dto.UpdateRoleRequest, updatedBy string) (dto.RoleDetailResponse, error) {
	var isSystem bool
	if err := s.db.QueryRowContext(ctx, `SELECT is_system FROM roles WHERE id = $1`, id).Scan(&isSystem); err != nil {
		return dto.RoleDetailResponse{}, err
	}
	if isSystem {
		return dto.RoleDetailResponse{}, response.NewFieldErr("role", "system role cannot be modified")
	}

	// Check if current user is using this role (superadmin excluded)
	var currentUserUsing bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM user_roles ur
			WHERE ur.role_id = $1 AND ur.user_id = $2
			AND ur.user_id NOT IN (
				SELECT ur2.user_id FROM user_roles ur2
				JOIN roles r ON r.id = ur2.role_id
				WHERE r.is_superadmin = true
			)
		)`, id, updatedBy,
	).Scan(&currentUserUsing)
	if err != nil {
		return dto.RoleDetailResponse{}, err
	}
	if currentUserUsing {
		return dto.RoleDetailResponse{}, response.NewFieldErr("role", "cannot modify a role you are currently using")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return dto.RoleDetailResponse{}, err
	}
	defer tx.Rollback()

	if req.Name != "" {
		_, err = tx.ExecContext(ctx,
			`UPDATE roles SET name = $1, updated_by = $2 WHERE id = $3`,
			req.Name, updatedBy, id,
		)
		if err != nil {
			return dto.RoleDetailResponse{}, err
		}
	}

	if len(req.PermissionIDs) > 0 {
		_, err = tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, id)
		if err != nil {
			return dto.RoleDetailResponse{}, err
		}

		for _, permID := range req.PermissionIDs {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`,
				id, permID,
			)
			if err != nil {
				return dto.RoleDetailResponse{}, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return dto.RoleDetailResponse{}, err
	}

	// Revoke tokens for users using this role (exclude superadmin)
	s.revokeRoleUsers(ctx, id)

	return s.GetByID(ctx, id)
}

func (s *RoleService) Delete(ctx context.Context, id string) error {
	var isSystem bool
	if err := s.db.QueryRowContext(ctx, `SELECT is_system FROM roles WHERE id = $1`, id).Scan(&isSystem); err != nil {
		return err
	}
	if isSystem {
		return response.NewFieldErr("role", "system role cannot be deleted")
	}

	var userCount int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM user_roles WHERE role_id = $1`, id,
	).Scan(&userCount); err != nil {
		return err
	}
	if userCount > 0 {
		return response.NewFieldErr("role", "cannot delete role that is assigned to users")
	}

	_, err := s.db.ExecContext(ctx, `DELETE FROM roles WHERE id = $1`, id)
	return err
}

func (s *RoleService) revokeRoleUsers(ctx context.Context, roleID string) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT DISTINCT ur.user_id FROM user_roles ur
		 WHERE ur.role_id = $1
		 AND ur.user_id NOT IN (
			SELECT ur2.user_id FROM user_roles ur2
			JOIN roles r ON r.id = ur2.role_id
			WHERE r.is_superadmin = true
		 )`, roleID,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			continue
		}
		middleware.RevokeAllTokens(s.redis, userID)
	}
}

func (s *RoleService) getRolePermissions(ctx context.Context, roleID string) ([]dto.PermissionResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT p.id, p.name
		 FROM permissions p
		 JOIN role_permissions rp ON rp.permission_id = p.id
		 WHERE rp.role_id = $1
		 ORDER BY p.name`,
		roleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := make([]dto.PermissionResponse, 0)
	for rows.Next() {
		var perm dto.PermissionResponse
		if err := rows.Scan(&perm.ID, &perm.Name); err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}
	return permissions, nil
}
