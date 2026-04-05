package service

import (
	"context"
	"database/sql"

	"gin-boilerplate/internal/iam/dto"
	"gin-boilerplate/pkg/auth"
	"gin-boilerplate/pkg/middleware"
	"gin-boilerplate/pkg/response"

	"github.com/redis/go-redis/v9"

	"golang.org/x/sync/singleflight"
)

type UserService struct {
	db    *sql.DB
	redis *redis.Client
	sf    singleflight.Group
}

func NewUserService(db *sql.DB, redis *redis.Client) *UserService {
	return &UserService{db: db, redis: redis}
}

func (s *UserService) GetByID(ctx context.Context, id string) (dto.UserDetailResponse, error) {
	result, err, _ := s.sf.Do("user:"+id, func() (any, error) {
		return s.getByIDFromDB(ctx, id)
	})
	if err != nil {
		return dto.UserDetailResponse{}, err
	}
	return result.(dto.UserDetailResponse), nil
}

func (s *UserService) getByIDFromDB(ctx context.Context, id string) (dto.UserDetailResponse, error) {
	var user dto.UserDetailResponse
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return dto.UserDetailResponse{}, err
	}

	roles, err := s.getUserRoles(ctx, id)
	if err != nil {
		return dto.UserDetailResponse{}, err
	}
	user.Roles = roles

	permissions, err := s.getUserPermissions(ctx, id)
	if err != nil {
		return dto.UserDetailResponse{}, err
	}
	user.Permissions = permissions

	return user, nil
}

func (s *UserService) List(ctx context.Context, limit, offset int) ([]dto.UserResponse, int, error) {
	var total int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, email, name, created_at, updated_at
		 FROM users
		 ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]dto.UserResponse, 0)
	for rows.Next() {
		var user dto.UserResponse
		if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, 0, err
		}

		roles, err := s.getUserRoles(ctx, user.ID)
		if err != nil {
			return nil, 0, err
		}
		user.Roles = roles

		users = append(users, user)
	}
	return users, total, nil
}

func (s *UserService) Create(ctx context.Context, req dto.CreateUserRequest, createdBy string) (dto.UserDetailResponse, error) {
	hashed, err := auth.HashPassword(req.Password)
	if err != nil {
		return dto.UserDetailResponse{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return dto.UserDetailResponse{}, err
	}
	defer tx.Rollback()

	var id string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO users (email, password, name, created_by, updated_by)
		 VALUES ($1, $2, $3, $4, $4)
		 RETURNING id`,
		req.Email, hashed, req.Name, createdBy,
	).Scan(&id)
	if err != nil {
		return dto.UserDetailResponse{}, err
	}

	for _, roleID := range req.RoleIDs {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`,
			id, roleID,
		)
		if err != nil {
			return dto.UserDetailResponse{}, err
		}
	}

	for _, permID := range req.PermissionIDs {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO user_permissions (user_id, permission_id) VALUES ($1, $2)`,
			id, permID,
		)
		if err != nil {
			return dto.UserDetailResponse{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return dto.UserDetailResponse{}, err
	}

	return s.GetByID(ctx, id)
}

func (s *UserService) Update(ctx context.Context, id string, req dto.UpdateUserRequest, updatedBy string) (dto.UserDetailResponse, error) {
	// Cannot modify own roles/permissions (unless superadmin)
	if updatedBy == id && (len(req.RoleIDs) > 0 || req.PermissionIDs != nil) {
		var isSuperadmin bool
		s.db.QueryRowContext(ctx,
			`SELECT EXISTS(
				SELECT 1 FROM user_roles ur
				JOIN roles r ON r.id = ur.role_id
				WHERE ur.user_id = $1 AND r.is_superadmin = true
			)`, updatedBy,
		).Scan(&isSuperadmin)

		if !isSuperadmin {
			return dto.UserDetailResponse{}, response.NewFieldErr("user", "cannot modify your own roles or permissions")
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return dto.UserDetailResponse{}, err
	}
	defer tx.Rollback()

	if req.Name != "" {
		_, err = tx.ExecContext(ctx,
			`UPDATE users SET name = $1, updated_by = $2 WHERE id = $3`,
			req.Name, updatedBy, id,
		)
		if err != nil {
			return dto.UserDetailResponse{}, err
		}
	}

	if len(req.RoleIDs) > 0 {
		_, err = tx.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id = $1`, id)
		if err != nil {
			return dto.UserDetailResponse{}, err
		}

		for _, roleID := range req.RoleIDs {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`,
				id, roleID,
			)
			if err != nil {
				return dto.UserDetailResponse{}, err
			}
		}
	}

	if req.PermissionIDs != nil {
		_, err = tx.ExecContext(ctx, `DELETE FROM user_permissions WHERE user_id = $1`, id)
		if err != nil {
			return dto.UserDetailResponse{}, err
		}

		for _, permID := range req.PermissionIDs {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO user_permissions (user_id, permission_id) VALUES ($1, $2)`,
				id, permID,
			)
			if err != nil {
				return dto.UserDetailResponse{}, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return dto.UserDetailResponse{}, err
	}

	return s.GetByID(ctx, id)
}

func (s *UserService) Delete(ctx context.Context, id string, deletedBy string) error {
	if id == deletedBy {
		return response.NewFieldErr("user", "cannot delete yourself")
	}

	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}

	middleware.RevokeAllTokens(s.redis, id)
	return nil
}

func (s *UserService) getUserRoles(ctx context.Context, userID string) ([]dto.UserRoleItem, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT r.id, r.name FROM roles r
		 JOIN user_roles ur ON ur.role_id = r.id
		 WHERE ur.user_id = $1
		 ORDER BY r.name`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]dto.UserRoleItem, 0)
	for rows.Next() {
		var role dto.UserRoleItem
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (s *UserService) getUserPermissions(ctx context.Context, userID string) ([]dto.UserPermissionItem, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT p.id, p.name FROM permissions p
		 JOIN user_permissions up ON up.permission_id = p.id
		 WHERE up.user_id = $1
		 ORDER BY p.name`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	perms := make([]dto.UserPermissionItem, 0)
	for rows.Next() {
		var perm dto.UserPermissionItem
		if err := rows.Scan(&perm.ID, &perm.Name); err != nil {
			return nil, err
		}
		perms = append(perms, perm)
	}
	return perms, nil
}
