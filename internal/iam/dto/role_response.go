package dto

type RoleResponse struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	IsSystem      bool   `json:"isSystem"`
	IsSuperadmin  bool   `json:"isSuperadmin"`
	IsDefault     bool   `json:"isDefault"`
}

type RoleDetailResponse struct {
	ID            string               `json:"id"`
	Name          string               `json:"name"`
	IsSystem      bool                 `json:"isSystem"`
	IsSuperadmin  bool                 `json:"isSuperadmin"`
	IsDefault     bool                 `json:"isDefault"`
	Permissions   []PermissionResponse `json:"permissions"`
}

type PermissionResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
