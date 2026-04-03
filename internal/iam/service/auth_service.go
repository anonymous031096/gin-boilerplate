package service

import (
	"context"
	"database/sql"
	"fmt"

	"gin-boilerplate/configs"
	"gin-boilerplate/internal/iam/dto"
	"gin-boilerplate/pkg/auth"
	"gin-boilerplate/pkg/middleware"
	"gin-boilerplate/pkg/response"

	"github.com/redis/go-redis/v9"
)

type AuthService struct {
	db    *sql.DB
	redis *redis.Client
}

func NewAuthService(db *sql.DB, redis *redis.Client) *AuthService {
	return &AuthService{db: db, redis: redis}
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest, deviceID string) (dto.TokenResponse, error) {
	var userID, hashedPassword string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, password FROM users WHERE email = $1`,
		req.Email,
	).Scan(&userID, &hashedPassword)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	if !auth.CheckPassword(req.Password, hashedPassword) {
		return dto.TokenResponse{}, sql.ErrNoRows
	}

	// Revoke old tokens for this user+device
	middleware.RevokeTokens(s.redis, userID, deviceID)

	roles, perms, err := s.getUserRolesAndPermissions(ctx, userID)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	accessToken, err := auth.GenerateAccessToken(configs.Get().JWT.AccessSecret, userID, deviceID, roles, perms, configs.Get().JWT.AccessTTL)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	refreshToken, err := auth.GenerateRefreshToken(configs.Get().JWT.RefreshSecret, userID, deviceID, configs.Get().JWT.RefreshTTL)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	return dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) error {
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`,
		req.Email,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return response.NewFieldErr("email", "email already exists")
	}

	hashed, err := auth.HashPassword(req.Password)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var userID string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO users (email, password, name)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		req.Email, hashed, req.Name,
	).Scan(&userID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO user_roles (user_id, role_id)
		 SELECT $1, id FROM roles WHERE name = 'user'`,
		userID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *AuthService) RefreshToken(ctx context.Context, req dto.RefreshTokenRequest, deviceID string) (dto.TokenResponse, error) {
	claims, err := auth.ParseRefreshToken(configs.Get().JWT.RefreshSecret, req.RefreshToken)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	userID := claims.GetUserID()
	tokenDeviceID := claims.GetDeviceID()

	// Check device mismatch
	if tokenDeviceID != deviceID {
		return dto.TokenResponse{}, fmt.Errorf("device mismatch")
	}

	// Check revocation
	revokedAt, err := s.redis.Get(ctx, fmt.Sprintf("revoke:%s:%s", userID, deviceID)).Int64()
	if err == nil && claims.IssuedAt != nil {
		if claims.IssuedAt.Time.Unix() < revokedAt {
			return dto.TokenResponse{}, fmt.Errorf("token has been revoked")
		}
	}

	roles, perms, err := s.getUserRolesAndPermissions(ctx, userID)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	accessToken, err := auth.GenerateAccessToken(configs.Get().JWT.AccessSecret, userID, deviceID, roles, perms, configs.Get().JWT.AccessTTL)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	refreshToken, err := auth.GenerateRefreshToken(configs.Get().JWT.RefreshSecret, userID, deviceID, configs.Get().JWT.RefreshTTL)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	return dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) getUserRolesAndPermissions(ctx context.Context, userID string) ([]auth.RoleClaim, []string, error) {
	// Get roles
	roleRows, err := s.db.QueryContext(ctx,
		`SELECT r.id, r.name FROM roles r
		 JOIN user_roles ur ON ur.role_id = r.id
		 WHERE ur.user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, nil, err
	}
	defer roleRows.Close()

	var roles []auth.RoleClaim
	for roleRows.Next() {
		var role auth.RoleClaim
		if err := roleRows.Scan(&role.ID, &role.Name); err != nil {
			return nil, nil, err
		}
		roles = append(roles, role)
	}

	// Get permissions per role
	for i, role := range roles {
		permRows, err := s.db.QueryContext(ctx,
			`SELECT p.name FROM permissions p
			 JOIN role_permissions rp ON rp.permission_id = p.id
			 WHERE rp.role_id = $1`,
			role.ID,
		)
		if err != nil {
			return nil, nil, err
		}

		var perms []string
		for permRows.Next() {
			var name string
			if err := permRows.Scan(&name); err != nil {
				permRows.Close()
				return nil, nil, err
			}
			perms = append(perms, name)
		}
		permRows.Close()

		roles[i].Permissions = perms
	}

	// Get direct user permissions
	directRows, err := s.db.QueryContext(ctx,
		`SELECT p.name FROM permissions p
		 JOIN user_permissions up ON up.permission_id = p.id
		 WHERE up.user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, nil, err
	}
	defer directRows.Close()

	var directPerms []string
	for directRows.Next() {
		var name string
		if err := directRows.Scan(&name); err != nil {
			return nil, nil, err
		}
		directPerms = append(directPerms, name)
	}

	return roles, directPerms, nil
}
