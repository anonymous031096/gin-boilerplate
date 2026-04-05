package service

import (
	"context"
	"database/sql"
	"testing"

	"gin-boilerplate/internal/iam/dto"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestGetRoleByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT r.id, r.name, r.is_system, r.is_superadmin, r.is_default`).
		WithArgs("role-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "is_system", "is_superadmin", "is_default", "perm_id", "perm_name"}).
			AddRow("role-uuid-1", "editor", false, false, false, "perm-1", "user:read").
			AddRow("role-uuid-1", "editor", false, false, false, "perm-2", "user:create"))

	svc := NewRoleService(db, nil)
	role, err := svc.GetByID(context.Background(), "role-uuid-1")

	assert.NoError(t, err)
	assert.Equal(t, "editor", role.Name)
	assert.False(t, role.IsSystem)
	assert.Len(t, role.Permissions, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRoleByID_Superadmin_AllPermissions(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT r.id, r.name, r.is_system, r.is_superadmin, r.is_default`).
		WithArgs("role-uuid-admin").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "is_system", "is_superadmin", "is_default", "perm_id", "perm_name"}).
			AddRow("role-uuid-admin", "admin", true, true, false, "p1", "user:create").
			AddRow("role-uuid-admin", "admin", true, true, false, "p2", "user:read").
			AddRow("role-uuid-admin", "admin", true, true, false, "p3", "role:create"))

	svc := NewRoleService(db, nil)
	role, err := svc.GetByID(context.Background(), "role-uuid-admin")

	assert.NoError(t, err)
	assert.Equal(t, "admin", role.Name)
	assert.True(t, role.IsSuperadmin)
	assert.Len(t, role.Permissions, 3)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRoleByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT r.id, r.name, r.is_system, r.is_superadmin, r.is_default`).
		WithArgs("nonexistent").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "is_system", "is_superadmin", "is_default", "perm_id", "perm_name"}))

	svc := NewRoleService(db, nil)
	_, err = svc.GetByID(context.Background(), "nonexistent")

	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestListRoles_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT id, name, is_system, is_superadmin, is_default FROM roles`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "is_system", "is_superadmin", "is_default"}).
			AddRow("r1", "admin", true, true, false).
			AddRow("r2", "user", true, false, true))

	svc := NewRoleService(db, nil)
	roles, err := svc.List(context.Background())

	assert.NoError(t, err)
	assert.Len(t, roles, 2)
	assert.Equal(t, "admin", roles[0].Name)
	assert.True(t, roles[0].IsSuperadmin)
	assert.True(t, roles[1].IsDefault)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListRoles_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT id, name, is_system, is_superadmin, is_default FROM roles`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "is_system", "is_superadmin", "is_default"}))

	svc := NewRoleService(db, nil)
	roles, err := svc.List(context.Background())

	assert.NoError(t, err)
	assert.Empty(t, roles)
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestCreateRole_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO roles`).
		WithArgs("editor", "creator-uuid").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("new-role-uuid"))
	mock.ExpectExec(`INSERT INTO role_permissions`).
		WithArgs("new-role-uuid", "perm-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// GetByID after create
	mock.ExpectQuery(`SELECT r.id, r.name, r.is_system, r.is_superadmin, r.is_default`).
		WithArgs("new-role-uuid").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "is_system", "is_superadmin", "is_default", "perm_id", "perm_name"}).
			AddRow("new-role-uuid", "editor", false, false, false, "perm-1", "user:read"))

	svc := NewRoleService(db, nil)
	role, err := svc.Create(context.Background(), dto.CreateRoleRequest{
		Name:          "editor",
		PermissionIDs: []string{"perm-1"},
	}, "creator-uuid")

	assert.NoError(t, err)
	assert.Equal(t, "editor", role.Name)
	assert.Len(t, role.Permissions, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestUpdateRole_SystemRoleCannotBeModified(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT is_system FROM roles WHERE id = \$1`).
		WithArgs("role-uuid-admin").
		WillReturnRows(sqlmock.NewRows([]string{"is_system"}).AddRow(true))

	svc := NewRoleService(db, nil)
	_, err = svc.Update(context.Background(), "role-uuid-admin", dto.UpdateRoleRequest{
		Name: "renamed",
	}, "updater-uuid")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "system role cannot be modified")
}

func TestUpdateRole_CannotModifyOwnRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT is_system FROM roles WHERE id = \$1`).
		WithArgs("role-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"is_system"}).AddRow(false))

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("role-uuid-1", "user-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	svc := NewRoleService(db, nil)
	_, err = svc.Update(context.Background(), "role-uuid-1", dto.UpdateRoleRequest{
		Name: "renamed",
	}, "user-uuid-1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot modify a role you are currently using")
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestDeleteRole_SystemRoleCannotBeDeleted(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT is_system FROM roles WHERE id = \$1`).
		WithArgs("role-uuid-admin").
		WillReturnRows(sqlmock.NewRows([]string{"is_system"}).AddRow(true))

	svc := NewRoleService(db, nil)
	err = svc.Delete(context.Background(), "role-uuid-admin")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "system role cannot be deleted")
}

func TestDeleteRole_CannotDeleteAssignedRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT is_system FROM roles WHERE id = \$1`).
		WithArgs("role-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"is_system"}).AddRow(false))

	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs("role-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	svc := NewRoleService(db, nil)
	err = svc.Delete(context.Background(), "role-uuid-1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete role that is assigned to users")
}

func TestDeleteRole_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT is_system FROM roles WHERE id = \$1`).
		WithArgs("role-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"is_system"}).AddRow(false))

	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs("role-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectExec(`DELETE FROM roles WHERE id = \$1`).
		WithArgs("role-uuid-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	svc := NewRoleService(db, nil)
	err = svc.Delete(context.Background(), "role-uuid-1")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
