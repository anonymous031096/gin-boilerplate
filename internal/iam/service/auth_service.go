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

func (s *AuthService) ChangePassword(ctx context.Context, userID string, deviceID string, req dto.ChangePasswordRequest) (*dto.TokenResponse, error) {
	var hashedPassword string
	err := s.db.QueryRowContext(ctx,
		`SELECT password FROM users WHERE id = $1`,
		userID,
	).Scan(&hashedPassword)
	if err != nil {
		return nil, err
	}

	if !auth.CheckPassword(req.OldPassword, hashedPassword) {
		return nil, response.NewFieldErr("oldPassword", "old password is incorrect")
	}

	hashed, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		return nil, err
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE users SET password = $1, updated_by = $2 WHERE id = $2`,
		hashed, userID,
	)
	if err != nil {
		return nil, err
	}

	if !req.LogoutOtherDevices {
		return nil, nil
	}

	// Revoke all devices
	middleware.RevokeAllTokens(s.redis, userID)

	// Generate new tokens for current device
	roles, perms, err := s.getUserRolesAndPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	cfg := configs.Get()
	accessToken, err := auth.GenerateAccessToken(cfg.JWT.AccessSecret, userID, deviceID, roles, perms, cfg.JWT.AccessTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := auth.GenerateRefreshToken(cfg.JWT.RefreshSecret, userID, deviceID, cfg.JWT.RefreshTTL)
	if err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
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

	// Revoke old tokens
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

func (s *AuthService) getUserRolesAndPermissions(ctx context.Context, userID string) ([]auth.RoleClaim, []string, error) {
	// Get roles with is_superadmin flag
	roleRows, err := s.db.QueryContext(ctx,
		`SELECT r.id, r.name, r.is_superadmin FROM roles r
		 JOIN user_roles ur ON ur.role_id = r.id
		 WHERE ur.user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, nil, err
	}
	defer roleRows.Close()

	type roleWithFlag struct {
		auth.RoleClaim
		isSuperadmin bool
	}

	var rolesWithFlag []roleWithFlag
	var hasSuperadmin bool
	for roleRows.Next() {
		var r roleWithFlag
		if err := roleRows.Scan(&r.ID, &r.Name, &r.isSuperadmin); err != nil {
			return nil, nil, err
		}
		if r.isSuperadmin {
			hasSuperadmin = true
		}
		rolesWithFlag = append(rolesWithFlag, r)
	}

	// Get all permissions once if superadmin
	var allPerms []string
	if hasSuperadmin {
		allPerms, err = s.getAllPermissions(ctx)
		if err != nil {
			return nil, nil, err
		}
	}

	// Build role claims with permissions
	roles := make([]auth.RoleClaim, len(rolesWithFlag))
	for i, r := range rolesWithFlag {
		roles[i] = r.RoleClaim
		if r.isSuperadmin {
			roles[i].Permissions = allPerms
		} else {
			perms, err := s.getRolePermissions(ctx, r.ID)
			if err != nil {
				return nil, nil, err
			}
			roles[i].Permissions = perms
		}
	}

	// Get direct user permissions
	directPerms, err := s.getDirectPermissions(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	return roles, directPerms, nil
}

func (s *AuthService) getAllPermissions(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT name FROM permissions ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	perms := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		perms = append(perms, name)
	}
	return perms, nil
}

func (s *AuthService) getRolePermissions(ctx context.Context, roleID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT p.name FROM permissions p
		 JOIN role_permissions rp ON rp.permission_id = p.id
		 WHERE rp.role_id = $1`,
		roleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	perms := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		perms = append(perms, name)
	}
	return perms, nil
}

func (s *AuthService) getDirectPermissions(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT p.name FROM permissions p
		 JOIN user_permissions up ON up.permission_id = p.id
		 WHERE up.user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	perms := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		perms = append(perms, name)
	}
	return perms, nil
}
