package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"gin-boilerplate/internal/iam/dto"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

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

	mock.ExpectQuery(`SELECT p.id, p.name FROM permissions p`).
		WithArgs("user-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	svc := NewUserService(db, nil)
	user, err := svc.GetByID(context.Background(), "user-uuid-1")

	assert.NoError(t, err)
	assert.Equal(t, "user-uuid-1", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Len(t, user.Roles, 1)
	assert.Equal(t, "admin", user.Roles[0].Name)
	assert.Empty(t, user.Permissions)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	svc := NewUserService(db, nil)
	_, err = svc.GetByID(context.Background(), "nonexistent")

	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestGetUserByID_WithPermissions(t *testing.T) {
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
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	mock.ExpectQuery(`SELECT p.id, p.name FROM permissions p`).
		WithArgs("user-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow("perm-uuid-1", "todo:create").
			AddRow("perm-uuid-2", "todo:read"))

	svc := NewUserService(db, nil)
	user, err := svc.GetByID(context.Background(), "user-uuid-1")

	assert.NoError(t, err)
	assert.Empty(t, user.Roles)
	assert.Len(t, user.Permissions, 2)
	assert.Equal(t, "todo:create", user.Permissions[0].Name)
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestListUsers_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at`).
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
			AddRow("user-1", "a@example.com", "User A", now, now).
			AddRow("user-2", "b@example.com", "User B", now, now))

	mock.ExpectQuery(`SELECT r.id, r.name FROM roles r`).WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("role-1", "admin"))
	mock.ExpectQuery(`SELECT r.id, r.name FROM roles r`).WithArgs("user-2").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("role-2", "user"))

	svc := NewUserService(db, nil)
	users, total, err := svc.List(context.Background(), 20, 0)

	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, users, 2)
	assert.Equal(t, "User A", users[0].Name)
	assert.Equal(t, "User B", users[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListUsers_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at`).
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}))

	svc := NewUserService(db, nil)
	users, total, err := svc.List(context.Background(), 20, 0)

	assert.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, users)
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestCreateUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("new@example.com", sqlmock.AnyArg(), "New User", "creator-uuid").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("new-uuid"))
	mock.ExpectExec(`INSERT INTO user_roles`).
		WithArgs("new-uuid", "role-uuid-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO user_permissions`).
		WithArgs("new-uuid", "perm-uuid-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// GetByID after create
	mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at`).
		WithArgs("new-uuid").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
			AddRow("new-uuid", "new@example.com", "New User", now, now))
	mock.ExpectQuery(`SELECT r.id, r.name FROM roles r`).WithArgs("new-uuid").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("role-uuid-1", "user"))
	mock.ExpectQuery(`SELECT p.id, p.name FROM permissions p`).WithArgs("new-uuid").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("perm-uuid-1", "todo:create"))

	svc := NewUserService(db, nil)
	user, err := svc.Create(context.Background(), dto.CreateUserRequest{
		Email:         "new@example.com",
		Password:      "Password1!",
		Name:          "New User",
		RoleIDs:       []string{"role-uuid-1"},
		PermissionIDs: []string{"perm-uuid-1"},
	}, "creator-uuid")

	assert.NoError(t, err)
	assert.Equal(t, "new-uuid", user.ID)
	assert.Equal(t, "new@example.com", user.Email)
	assert.Len(t, user.Roles, 1)
	assert.Len(t, user.Permissions, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestUpdateUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE users SET name`).
		WithArgs("Updated Name", "updater-uuid", "user-uuid-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// GetByID after update
	mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at`).
		WithArgs("user-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
			AddRow("user-uuid-1", "test@example.com", "Updated Name", now, now))
	mock.ExpectQuery(`SELECT r.id, r.name FROM roles r`).WithArgs("user-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("role-1", "user"))
	mock.ExpectQuery(`SELECT p.id, p.name FROM permissions p`).WithArgs("user-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	svc := NewUserService(db, nil)
	user, err := svc.Update(context.Background(), "user-uuid-1", dto.UpdateUserRequest{
		Name: "Updated Name",
	}, "updater-uuid")

	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", user.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateUser_CannotModifyOwnRoles(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Check if user is superadmin
	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("user-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	svc := NewUserService(db, nil)
	_, err = svc.Update(context.Background(), "user-uuid-1", dto.UpdateUserRequest{
		RoleIDs: []string{"role-uuid-new"},
	}, "user-uuid-1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot modify your own roles or permissions")
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestDeleteUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectExec(`DELETE FROM users WHERE id = \$1`).
		WithArgs("user-uuid-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	rdb := newTestRedisClient()
	svc := NewUserService(db, rdb)
	err = svc.Delete(context.Background(), "user-uuid-1", "admin-uuid")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteUser_CannotDeleteSelf(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	svc := NewUserService(db, nil)
	err = svc.Delete(context.Background(), "user-uuid-1", "user-uuid-1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete yourself")
}
