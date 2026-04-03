package dto

type RoleResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Permissions []PermissionResponse `json:"permissions"`
}

type PermissionResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
