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
	rows, err := s.db.QueryContext(ctx,
		`SELECT u.id, u.email, u.name, u.created_at, u.updated_at,
		        COALESCE(r.id::text, '') AS role_id, COALESCE(r.name, '') AS role_name,
		        COALESCE(p.id::text, '') AS perm_id, COALESCE(p.name, '') AS perm_name
		 FROM users u
		 LEFT JOIN user_roles ur ON ur.user_id = u.id
		 LEFT JOIN roles r ON r.id = ur.role_id
		 LEFT JOIN user_permissions up ON up.user_id = u.id
		 LEFT JOIN permissions p ON p.id = up.permission_id
		 WHERE u.id = $1`,
		id,
	)
	if err != nil {
		return dto.UserDetailResponse{}, err
	}
	defer rows.Close()

	var user dto.UserDetailResponse
	user.Roles = make([]dto.UserRoleItem, 0)
	user.Permissions = make([]dto.UserPermissionItem, 0)
	var found bool
	seenRoles := make(map[string]bool)
	seenPerms := make(map[string]bool)

	for rows.Next() {
		var roleID, roleName, permID, permName string
		if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt,
			&roleID, &roleName, &permID, &permName); err != nil {
			return dto.UserDetailResponse{}, err
		}
		found = true
		if roleID != "" && !seenRoles[roleID] {
			seenRoles[roleID] = true
			user.Roles = append(user.Roles, dto.UserRoleItem{ID: roleID, Name: roleName})
		}
		if permID != "" && !seenPerms[permID] {
			seenPerms[permID] = true
			user.Permissions = append(user.Permissions, dto.UserPermissionItem{ID: permID, Name: permName})
		}
	}

	if !found {
		return dto.UserDetailResponse{}, sql.ErrNoRows
	}

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
