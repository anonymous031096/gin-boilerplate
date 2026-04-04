package service

import (
	"context"
	"os"
	"testing"

	"gin-boilerplate/configs"
	"gin-boilerplate/internal/iam/dto"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestMain(m *testing.M) {
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
	os.Exit(m.Run())
}

func TestLogin_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	// Login query
	mock.ExpectQuery(`SELECT id, password FROM users WHERE email = \$1`).
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "password"}).
			AddRow("user-uuid-1", string(hashed)))

	// getUserRolesAndPermissions: get roles
	mock.ExpectQuery(`SELECT r.id, r.name FROM roles r`).
		WithArgs("user-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow("role-uuid-1", "admin"))

	// getUserRolesAndPermissions: get role permissions
	mock.ExpectQuery(`SELECT p.name FROM permissions p`).
		WithArgs("role-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"name"}).
			AddRow("user:read").
			AddRow("user:create"))

	// getUserRolesAndPermissions: get direct permissions
	mock.ExpectQuery(`SELECT p.name FROM permissions p`).
		WithArgs("user-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"name"}))

	svc := NewAuthService(db, nil)
	result, err := svc.Login(context.Background(), dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}, "unknown")

	assert.NoError(t, err)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogin_WrongPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	mock.ExpectQuery(`SELECT id, password FROM users WHERE email = \$1`).
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "password"}).
			AddRow("user-uuid-1", string(hashed)))

	svc := NewAuthService(db, nil)
	_, err = svc.Login(context.Background(), dto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}, "unknown")

	assert.Error(t, err)
}
