package service

import (
	"context"
	"database/sql"

	"gin-boilerplate/internal/iam/dto"
)

type RoleService struct {
	db *sql.DB
}

func NewRoleService(db *sql.DB) *RoleService {
	return &RoleService{db: db}
}

func (s *RoleService) GetByID(ctx context.Context, id string) (dto.RoleResponse, error) {
	var role dto.RoleResponse
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name FROM roles WHERE id = $1`,
		id,
	).Scan(&role.ID, &role.Name)
	if err != nil {
		return dto.RoleResponse{}, err
	}

	permissions, err := s.getRolePermissions(ctx, id)
	if err != nil {
		return dto.RoleResponse{}, err
	}
	role.Permissions = permissions

	return role, nil
}

func (s *RoleService) List(ctx context.Context) ([]dto.RoleResponse, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name FROM roles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []dto.RoleResponse
	for rows.Next() {
		var role dto.RoleResponse
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			return nil, err
		}

		permissions, err := s.getRolePermissions(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		role.Permissions = permissions

		roles = append(roles, role)
	}
	return roles, nil
}

func (s *RoleService) Create(ctx context.Context, req dto.CreateRoleRequest, createdBy string) (dto.RoleResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return dto.RoleResponse{}, err
	}
	defer tx.Rollback()

	var roleID string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO roles (name, created_by, updated_by) VALUES ($1, $2, $2) RETURNING id`,
		req.Name, createdBy,
	).Scan(&roleID)
	if err != nil {
		return dto.RoleResponse{}, err
	}

	for _, permID := range req.PermissionIDs {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`,
			roleID, permID,
		)
		if err != nil {
			return dto.RoleResponse{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return dto.RoleResponse{}, err
	}

	return s.GetByID(ctx, roleID)
}

func (s *RoleService) Update(ctx context.Context, id string, req dto.UpdateRoleRequest, updatedBy string) (dto.RoleResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return dto.RoleResponse{}, err
	}
	defer tx.Rollback()

	if req.Name != "" {
		_, err = tx.ExecContext(ctx,
			`UPDATE roles SET name = $1, updated_by = $2 WHERE id = $3`,
			req.Name, updatedBy, id,
		)
		if err != nil {
			return dto.RoleResponse{}, err
		}
	}

	if len(req.PermissionIDs) > 0 {
		_, err = tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, id)
		if err != nil {
			return dto.RoleResponse{}, err
		}

		for _, permID := range req.PermissionIDs {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`,
				id, permID,
			)
			if err != nil {
				return dto.RoleResponse{}, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return dto.RoleResponse{}, err
	}

	return s.GetByID(ctx, id)
}

func (s *RoleService) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM roles WHERE id = $1`, id)
	return err
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

	var permissions []dto.PermissionResponse
	for rows.Next() {
		var perm dto.PermissionResponse
		if err := rows.Scan(&perm.ID, &perm.Name); err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}
	return permissions, nil
}
