package service

import (
	"context"
	"database/sql"
	"errors"

	"gin-boilerplate/internal/todo/dto"
)

type TodoService struct {
	db *sql.DB
}

func NewTodoService(db *sql.DB) *TodoService {
	return &TodoService{db: db}
}

func (s *TodoService) GetByID(ctx context.Context, id string, userID string) (dto.TodoResponse, error) {
	var todo dto.TodoResponse
	err := s.db.QueryRowContext(ctx,
		`SELECT id, title, description, completed, created_by, created_at, updated_at
		 FROM todos
		 WHERE id = $1`,
		id,
	).Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedBy, &todo.CreatedAt, &todo.UpdatedAt)
	if err != nil {
		return dto.TodoResponse{}, err
	}

	if todo.CreatedBy != userID {
		return dto.TodoResponse{}, errors.New("forbidden: not owner")
	}

	return todo, nil
}

func (s *TodoService) List(ctx context.Context, userID string, limit, offset int) ([]dto.TodoResponse, int, error) {
	var total int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM todos WHERE created_by = $1`,
		userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, description, completed, created_by, created_at, updated_at
		 FROM todos
		 WHERE created_by = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var todos []dto.TodoResponse
	for rows.Next() {
		var todo dto.TodoResponse
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedBy, &todo.CreatedAt, &todo.UpdatedAt); err != nil {
			return nil, 0, err
		}
		todos = append(todos, todo)
	}
	return todos, total, nil
}

func (s *TodoService) Create(ctx context.Context, req dto.CreateTodoRequest, userID string) (dto.TodoResponse, error) {
	var todo dto.TodoResponse
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO todos (title, description, created_by, updated_by)
		 VALUES ($1, $2, $3, $3)
		 RETURNING id, title, description, completed, created_by, created_at, updated_at`,
		req.Title, req.Description, userID,
	).Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedBy, &todo.CreatedAt, &todo.UpdatedAt)
	if err != nil {
		return dto.TodoResponse{}, err
	}
	return todo, nil
}

func (s *TodoService) Update(ctx context.Context, id string, req dto.UpdateTodoRequest, userID string) (dto.TodoResponse, error) {
	// Ownership check
	var ownerID string
	err := s.db.QueryRowContext(ctx,
		`SELECT created_by FROM todos WHERE id = $1`,
		id,
	).Scan(&ownerID)
	if err != nil {
		return dto.TodoResponse{}, err
	}
	if ownerID != userID {
		return dto.TodoResponse{}, errors.New("forbidden: not owner")
	}

	var todo dto.TodoResponse
	err = s.db.QueryRowContext(ctx,
		`UPDATE todos SET
			title = COALESCE(NULLIF($1, ''), title),
			description = COALESCE(NULLIF($2, ''), description),
			completed = COALESCE($3, completed),
			updated_by = $4
		 WHERE id = $5
		 RETURNING id, title, description, completed, created_by, created_at, updated_at`,
		req.Title, req.Description, req.Completed, userID, id,
	).Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedBy, &todo.CreatedAt, &todo.UpdatedAt)
	if err != nil {
		return dto.TodoResponse{}, err
	}
	return todo, nil
}

func (s *TodoService) Delete(ctx context.Context, id string, userID string) error {
	var ownerID string
	err := s.db.QueryRowContext(ctx,
		`SELECT created_by FROM todos WHERE id = $1`,
		id,
	).Scan(&ownerID)
	if err != nil {
		return err
	}
	if ownerID != userID {
		return errors.New("forbidden: not owner")
	}

	_, err = s.db.ExecContext(ctx, `DELETE FROM todos WHERE id = $1`, id)
	return err
}
