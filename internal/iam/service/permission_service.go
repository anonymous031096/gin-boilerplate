package service

import (
	"context"
	"database/sql"

	"gin-boilerplate/internal/iam/dto"
)

type PermissionService struct {
	db *sql.DB
}

func NewPermissionService(db *sql.DB) *PermissionService {
	return &PermissionService{db: db}
}

func (s *PermissionService) List(ctx context.Context) ([]dto.PermissionResponse, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name FROM permissions ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []dto.PermissionResponse
	for rows.Next() {
		var perm dto.PermissionResponse
		if err := rows.Scan(&perm.ID, &perm.Name); err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}
	return permissions, nil
}
