package service

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetUserByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at`).
		WithArgs("user-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
			AddRow("user-uuid-1", "test@example.com", "Test User", now, now))

	mock.ExpectQuery(`SELECT r.id, r.name FROM roles r`).
		WithArgs("user-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow("role-uuid-1", "admin"))

	svc := NewUserService(db)
	user, err := svc.GetByID(context.Background(), "user-uuid-1")

	assert.NoError(t, err)
	assert.Equal(t, "user-uuid-1", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Len(t, user.Roles, 1)
	assert.Equal(t, "admin", user.Roles[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListUsers_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	now := time.Now()
	mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at`).
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
			AddRow("user-uuid-1", "test@example.com", "Test User", now, now))

	mock.ExpectQuery(`SELECT r.id, r.name FROM roles r`).
		WithArgs("user-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow("role-uuid-1", "admin"))

	svc := NewUserService(db)
	users, total, err := svc.List(context.Background(), 20, 0)

	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, users, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}
