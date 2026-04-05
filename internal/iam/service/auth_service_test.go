package service

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"gin-boilerplate/configs"
	"gin-boilerplate/internal/iam/dto"
	"gin-boilerplate/pkg/auth"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

var testRedis *miniredis.Miniredis

func TestMain(m *testing.M) {
	// Create a temporary .env so godotenv.Load() does not panic.
	// The real values come from os.Setenv calls below.
	_ = os.WriteFile(".env", []byte(""), 0644)
	defer os.Remove(".env")

	os.Setenv("PORT", "8080")
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_PORT", "5432")
	os.Setenv("POSTGRES_USER", "test")
	os.Setenv("POSTGRES_PASSWORD", "test")
	os.Setenv("POSTGRES_DB", "test")
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("JWT_ACCESS_SECRET", "test-access-secret")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret")
	os.Setenv("JWT_ACCESS_TTL", "15m")
	os.Setenv("JWT_REFRESH_TTL", "168h")
	configs.Load()

	testRedis = miniredis.RunT(&testing.T{})

	code := m.Run()
	testRedis.Close()
	os.Exit(code)
}

func newTestRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: testRedis.Addr(),
	})
}

// helper: hash a password for mock rows
func hashPassword(t *testing.T, plain string) string {
	t.Helper()
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	assert.NoError(t, err)
	return string(hashed)
}

// helper: mock the getUserRolesAndPermissions queries (non-superadmin role)
func expectRolesAndPermissions(mock sqlmock.Sqlmock, userID, roleID, roleName string, rolePerms, directPerms []string) {
	// roles query
	mock.ExpectQuery(`SELECT r\.id, r\.name, r\.is_superadmin FROM roles r`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "is_superadmin"}).
			AddRow(roleID, roleName, false))

	// role permissions query
	permRows := sqlmock.NewRows([]string{"name"})
	for _, p := range rolePerms {
		permRows.AddRow(p)
	}
	mock.ExpectQuery(`SELECT p\.name FROM permissions p`).
		WithArgs(roleID).
		WillReturnRows(permRows)

	// direct permissions query
	directRows := sqlmock.NewRows([]string{"name"})
	for _, p := range directPerms {
		directRows.AddRow(p)
	}
	mock.ExpectQuery(`SELECT p\.name FROM permissions p`).
		WithArgs(userID).
		WillReturnRows(directRows)
}

// ---------------------------------------------------------------------------
// Login
// ---------------------------------------------------------------------------

func TestLogin_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	hashed := hashPassword(t, "password123")

	mock.ExpectQuery(`SELECT id, password FROM users WHERE email = \$1`).
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "password"}).
			AddRow("user-uuid-1", hashed))

	expectRolesAndPermissions(mock, "user-uuid-1", "role-uuid-1", "user", []string{"user:read", "user:create"}, nil)

	rdb := newTestRedisClient()
	svc := NewAuthService(db, rdb)
	result, err := svc.Login(context.Background(), dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}, "device-1")

	assert.NoError(t, err)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)

	// Verify the access token contains expected claims
	claims, err := auth.ParseAccessToken(configs.Get().JWT.AccessSecret, result.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, "user-uuid-1", claims.GetUserID())
	assert.Equal(t, "device-1", claims.GetDeviceID())
	assert.Len(t, claims.Roles, 1)
	assert.Equal(t, "user", claims.Roles[0].Name)
	assert.ElementsMatch(t, []string{"user:read", "user:create"}, claims.Roles[0].Permissions)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogin_WrongPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	hashed := hashPassword(t, "password123")

	mock.ExpectQuery(`SELECT id, password FROM users WHERE email = \$1`).
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "password"}).
			AddRow("user-uuid-1", hashed))

	svc := NewAuthService(db, newTestRedisClient())
	_, err = svc.Login(context.Background(), dto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}, "device-1")

	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogin_UserNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT id, password FROM users WHERE email = \$1`).
		WithArgs("nonexistent@example.com").
		WillReturnError(sql.ErrNoRows)

	svc := NewAuthService(db, newTestRedisClient())
	_, err = svc.Login(context.Background(), dto.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}, "device-1")

	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// Register
// ---------------------------------------------------------------------------

func TestRegister_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Email existence check
	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("new@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Begin transaction
	mock.ExpectBegin()

	// Insert user
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("new@example.com", sqlmock.AnyArg(), "New User").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("new-user-uuid"))

	// Assign default role
	mock.ExpectExec(`INSERT INTO user_roles`).
		WithArgs("new-user-uuid").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Commit
	mock.ExpectCommit()

	svc := NewAuthService(db, nil)
	err = svc.Register(context.Background(), dto.RegisterRequest{
		Email:    "new@example.com",
		Password: "Password1!",
		Name:     "New User",
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("existing@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	svc := NewAuthService(db, nil)
	err = svc.Register(context.Background(), dto.RegisterRequest{
		Email:    "existing@example.com",
		Password: "Password1!",
		Name:     "Existing User",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email already exists")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// ChangePassword
// ---------------------------------------------------------------------------

func TestChangePassword_SuccessWithoutLogout(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	hashed := hashPassword(t, "oldpassword")

	// Fetch current password
	mock.ExpectQuery(`SELECT password FROM users WHERE id = \$1`).
		WithArgs("user-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"password"}).AddRow(hashed))

	// Update password
	mock.ExpectExec(`UPDATE users SET password`).
		WithArgs(sqlmock.AnyArg(), "user-uuid-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	svc := NewAuthService(db, newTestRedisClient())
	result, err := svc.ChangePassword(context.Background(), "user-uuid-1", "device-1", dto.ChangePasswordRequest{
		OldPassword:        "oldpassword",
		NewPassword:        "Newpassword1!",
		LogoutOtherDevices: false,
	})

	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChangePassword_SuccessWithLogout(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	hashed := hashPassword(t, "oldpassword")

	// Fetch current password
	mock.ExpectQuery(`SELECT password FROM users WHERE id = \$1`).
		WithArgs("user-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"password"}).AddRow(hashed))

	// Update password
	mock.ExpectExec(`UPDATE users SET password`).
		WithArgs(sqlmock.AnyArg(), "user-uuid-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// getUserRolesAndPermissions for new token generation
	expectRolesAndPermissions(mock, "user-uuid-1", "role-uuid-1", "user", []string{"user:read"}, nil)

	rdb := newTestRedisClient()
	svc := NewAuthService(db, rdb)
	result, err := svc.ChangePassword(context.Background(), "user-uuid-1", "device-1", dto.ChangePasswordRequest{
		OldPassword:        "oldpassword",
		NewPassword:        "Newpassword1!",
		LogoutOtherDevices: true,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)

	// Verify new token contains correct user and device
	claims, err := auth.ParseAccessToken(configs.Get().JWT.AccessSecret, result.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, "user-uuid-1", claims.GetUserID())
	assert.Equal(t, "device-1", claims.GetDeviceID())

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChangePassword_WrongOldPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	hashed := hashPassword(t, "correctpassword")

	mock.ExpectQuery(`SELECT password FROM users WHERE id = \$1`).
		WithArgs("user-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"password"}).AddRow(hashed))

	svc := NewAuthService(db, newTestRedisClient())
	result, err := svc.ChangePassword(context.Background(), "user-uuid-1", "device-1", dto.ChangePasswordRequest{
		OldPassword:        "wrongpassword",
		NewPassword:        "Newpassword1!",
		LogoutOtherDevices: false,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "old password is incorrect")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// RefreshToken
// ---------------------------------------------------------------------------

func TestRefreshToken_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	cfg := configs.Get()

	// Generate a valid refresh token to use as input
	refreshToken, err := auth.GenerateRefreshToken(cfg.JWT.RefreshSecret, "user-uuid-1", "device-1", cfg.JWT.RefreshTTL)
	assert.NoError(t, err)

	// getUserRolesAndPermissions
	expectRolesAndPermissions(mock, "user-uuid-1", "role-uuid-1", "user", []string{"user:read"}, nil)

	rdb := newTestRedisClient()
	svc := NewAuthService(db, rdb)
	result, err := svc.RefreshToken(context.Background(), dto.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}, "device-1")

	assert.NoError(t, err)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)

	// Verify new access token
	claims, err := auth.ParseAccessToken(cfg.JWT.AccessSecret, result.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, "user-uuid-1", claims.GetUserID())
	assert.Equal(t, "device-1", claims.GetDeviceID())

	assert.NoError(t, mock.ExpectationsWereMet())
}
