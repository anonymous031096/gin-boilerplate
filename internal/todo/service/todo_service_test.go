package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"gin-boilerplate/internal/todo/dto"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestGetTodoByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery(`SELECT id, title, description, completed, created_by, created_at, updated_at`).
		WithArgs("todo-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "completed", "created_by", "created_at", "updated_at"}).
			AddRow("todo-uuid-1", "Buy milk", "From store", false, "user-uuid-1", now, now))

	svc := NewTodoService(db)
	todo, err := svc.GetByID(context.Background(), "todo-uuid-1", "user-uuid-1")

	assert.NoError(t, err)
	assert.Equal(t, "Buy milk", todo.Title)
	assert.Equal(t, "user-uuid-1", todo.CreatedBy)
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

func TestGetTodoByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT id, title, description, completed, created_by, created_at, updated_at`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	svc := NewTodoService(db)
	_, err = svc.GetByID(context.Background(), "nonexistent", "user-uuid-1")

	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestListTodos_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs("user-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`SELECT id, title, description, completed, created_by, created_at, updated_at`).
		WithArgs("user-uuid-1", 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "completed", "created_by", "created_at", "updated_at"}).
			AddRow("t1", "Todo 1", "", false, "user-uuid-1", now, now).
			AddRow("t2", "Todo 2", "desc", true, "user-uuid-1", now, now))

	svc := NewTodoService(db)
	todos, total, err := svc.List(context.Background(), "user-uuid-1", 20, 0)

	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, todos, 2)
	assert.Equal(t, "Todo 1", todos[0].Title)
	assert.True(t, todos[1].Completed)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListTodos_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs("user-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`SELECT id, title, description, completed, created_by, created_at, updated_at`).
		WithArgs("user-uuid-1", 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "completed", "created_by", "created_at", "updated_at"}))

	svc := NewTodoService(db)
	todos, total, err := svc.List(context.Background(), "user-uuid-1", 20, 0)

	assert.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, todos)
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

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
	assert.False(t, todo.Completed)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTodo_WithDescription(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery(`INSERT INTO todos`).
		WithArgs("Buy milk", "From the store", "user-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "completed", "created_by", "created_at", "updated_at"}).
			AddRow("todo-uuid-1", "Buy milk", "From the store", false, "user-uuid-1", now, now))

	svc := NewTodoService(db)
	todo, err := svc.Create(context.Background(), dto.CreateTodoRequest{
		Title:       "Buy milk",
		Description: "From the store",
	}, "user-uuid-1")

	assert.NoError(t, err)
	assert.Equal(t, "From the store", todo.Description)
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestUpdateTodo_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	now := time.Now()

	// Ownership check
	mock.ExpectQuery(`SELECT created_by FROM todos WHERE id = \$1`).
		WithArgs("todo-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"created_by"}).AddRow("user-uuid-1"))

	completed := true
	mock.ExpectQuery(`UPDATE todos SET`).
		WithArgs("Updated title", "", &completed, "user-uuid-1", "todo-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "completed", "created_by", "created_at", "updated_at"}).
			AddRow("todo-uuid-1", "Updated title", "", true, "user-uuid-1", now, now))

	svc := NewTodoService(db)
	todo, err := svc.Update(context.Background(), "todo-uuid-1", dto.UpdateTodoRequest{
		Title:     "Updated title",
		Completed: &completed,
	}, "user-uuid-1")

	assert.NoError(t, err)
	assert.Equal(t, "Updated title", todo.Title)
	assert.True(t, todo.Completed)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTodo_NotOwner(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT created_by FROM todos WHERE id = \$1`).
		WithArgs("todo-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"created_by"}).AddRow("other-user"))

	svc := NewTodoService(db)
	_, err = svc.Update(context.Background(), "todo-uuid-1", dto.UpdateTodoRequest{
		Title: "Hacked",
	}, "user-uuid-1")

	assert.Error(t, err)
	assert.Equal(t, "forbidden: not owner", err.Error())
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestDeleteTodo_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT created_by FROM todos WHERE id = \$1`).
		WithArgs("todo-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"created_by"}).AddRow("user-uuid-1"))

	mock.ExpectExec(`DELETE FROM todos WHERE id = \$1`).
		WithArgs("todo-uuid-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	svc := NewTodoService(db)
	err = svc.Delete(context.Background(), "todo-uuid-1", "user-uuid-1")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteTodo_NotOwner(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT created_by FROM todos WHERE id = \$1`).
		WithArgs("todo-uuid-1").
		WillReturnRows(sqlmock.NewRows([]string{"created_by"}).AddRow("other-user"))

	svc := NewTodoService(db)
	err = svc.Delete(context.Background(), "todo-uuid-1", "user-uuid-1")

	assert.Error(t, err)
	assert.Equal(t, "forbidden: not owner", err.Error())
}

func TestDeleteTodo_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT created_by FROM todos WHERE id = \$1`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	svc := NewTodoService(db)
	err = svc.Delete(context.Background(), "nonexistent", "user-uuid-1")

	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}
