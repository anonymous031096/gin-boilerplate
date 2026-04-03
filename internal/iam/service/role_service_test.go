package service

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetRoleByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT id, name FROM roles WHERE id = \$1`).
		WithArgs("role-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow("role-uuid-1", "admin"))

	mock.ExpectQuery(`SELECT p.id, p.name`).
		WithArgs("role-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow("perm-uuid-1", "user:read").
			AddRow("perm-uuid-2", "user:create"))

	svc := NewRoleService(db)
	role, err := svc.GetByID(context.Background(), "role-uuid-1")

	assert.NoError(t, err)
	assert.Equal(t, "admin", role.Name)
	assert.Len(t, role.Permissions, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}
