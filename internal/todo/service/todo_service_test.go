package service

import (
	"context"
	"testing"
	"time"

	"gin-boilerplate/internal/todo/dto"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCreateTodo_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery(`INSERT INTO todos`).
		WithArgs("Buy milk", "", "user-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "completed", "created_by", "created_at", "updated_at"}).
			AddRow("todo-uuid-1", "Buy milk", "", false, "user-uuid-1", now, now))

	svc := NewTodoService(db)
	todo, err := svc.Create(context.Background(), dto.CreateTodoRequest{
		Title: "Buy milk",
	}, "user-uuid-1")

	assert.NoError(t, err)
	assert.Equal(t, "Buy milk", todo.Title)
	assert.Equal(t, "user-uuid-1", todo.CreatedBy)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTodoByID_NotOwner(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery(`SELECT id, title, description, completed, created_by, created_at, updated_at`).
		WithArgs("todo-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "completed", "created_by", "created_at", "updated_at"}).
			AddRow("todo-uuid-1", "Buy milk", "", false, "other-user", now, now))

	svc := NewTodoService(db)
	_, err = svc.GetByID(context.Background(), "todo-uuid-1", "user-uuid-1")

	assert.Error(t, err)
	assert.Equal(t, "forbidden: not owner", err.Error())
}
