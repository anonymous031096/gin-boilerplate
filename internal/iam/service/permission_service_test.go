package service

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestListPermissions_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT id, name FROM permissions`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow("p1", "user:create").
			AddRow("p2", "user:read").
			AddRow("p3", "role:create"))

	svc := NewPermissionService(db)
	perms, err := svc.List(context.Background())

	assert.NoError(t, err)
	assert.Len(t, perms, 3)
	assert.Equal(t, "user:create", perms[0].Name)
	assert.Equal(t, "user:read", perms[1].Name)
	assert.Equal(t, "role:create", perms[2].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListPermissions_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT id, name FROM permissions`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	svc := NewPermissionService(db)
	perms, err := svc.List(context.Background())

	assert.NoError(t, err)
	assert.Empty(t, perms)
	assert.NoError(t, mock.ExpectationsWereMet())
}
