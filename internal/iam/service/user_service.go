package service

import (
	"context"
	"database/sql"

	"gin-boilerplate/internal/iam/dto"
	"gin-boilerplate/pkg/auth"

	"golang.org/x/sync/singleflight"
)

type UserService struct {
	db *sql.DB
	sf singleflight.Group
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetByID(ctx context.Context, id string) (dto.UserResponse, error) {
	result, err, _ := s.sf.Do("user:"+id, func() (any, error) {
		return s.getByIDFromDB(ctx, id)
	})
	if err != nil {
		return dto.UserResponse{}, err
	}
	return result.(dto.UserResponse), nil
}

func (s *UserService) getByIDFromDB(ctx context.Context, id string) (dto.UserResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT u.id, u.email, u.name, u.created_at, u.updated_at,
		        COALESCE(r.id::text, '') as role_id, COALESCE(r.name, '') as role_name
		 FROM users u
		 LEFT JOIN user_roles ur ON ur.user_id = u.id
		 LEFT JOIN roles r ON r.id = ur.role_id
		 WHERE u.id = $1`,
		id,
	)
	if err != nil {
		return dto.UserResponse{}, err
	}
	defer rows.Close()

	var user dto.UserResponse
	var found bool
	for rows.Next() {
		var roleID, roleName string
		if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt, &roleID, &roleName); err != nil {
			return dto.UserResponse{}, err
		}
		found = true
		if roleID != "" {
			user.Roles = append(user.Roles, dto.UserRoleItem{ID: roleID, Name: roleName})
		}
	}

	if !found {
		return dto.UserResponse{}, sql.ErrNoRows
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

	var users []dto.UserResponse
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

func (s *UserService) Create(ctx context.Context, req dto.CreateUserRequest, createdBy string) (dto.UserResponse, error) {
	hashed, err := auth.HashPassword(req.Password)
	if err != nil {
		return dto.UserResponse{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return dto.UserResponse{}, err
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
		return dto.UserResponse{}, err
	}

	for _, roleID := range req.RoleIDs {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`,
			id, roleID,
		)
		if err != nil {
			return dto.UserResponse{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return dto.UserResponse{}, err
	}

	return s.GetByID(ctx, id)
}

func (s *UserService) Update(ctx context.Context, id string, req dto.UpdateUserRequest, updatedBy string) (dto.UserResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return dto.UserResponse{}, err
	}
	defer tx.Rollback()

	if req.Name != "" {
		_, err = tx.ExecContext(ctx,
			`UPDATE users SET name = $1, updated_by = $2 WHERE id = $3`,
			req.Name, updatedBy, id,
		)
		if err != nil {
			return dto.UserResponse{}, err
		}
	}

	if len(req.RoleIDs) > 0 {
		_, err = tx.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id = $1`, id)
		if err != nil {
			return dto.UserResponse{}, err
		}

		for _, roleID := range req.RoleIDs {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`,
				id, roleID,
			)
			if err != nil {
				return dto.UserResponse{}, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return dto.UserResponse{}, err
	}

	return s.GetByID(ctx, id)
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
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

	var roles []dto.UserRoleItem
	for rows.Next() {
		var role dto.UserRoleItem
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}
