package dto

import "time"

type UserResponse struct {
	ID        string         `json:"id"`
	Email     string         `json:"email"`
	Name      string         `json:"name"`
	Roles     []UserRoleItem `json:"roles"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

type UserDetailResponse struct {
	ID          string               `json:"id"`
	Email       string               `json:"email"`
	Name        string               `json:"name"`
	Roles       []UserRoleItem       `json:"roles"`
	Permissions []UserPermissionItem `json:"permissions"`
	CreatedAt   time.Time            `json:"createdAt"`
	UpdatedAt   time.Time            `json:"updatedAt"`
}

type UserRoleItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserPermissionItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserListResponse struct {
	Data []UserResponse `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type PaginationMeta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}
