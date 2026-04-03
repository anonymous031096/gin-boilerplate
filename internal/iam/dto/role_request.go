package dto

type CreateRoleRequest struct {
	Name          string   `json:"name" binding:"required"`
	PermissionIDs []string `json:"permissionIds" binding:"required,min=1,dive,uuid"`
}

type UpdateRoleRequest struct {
	Name          string   `json:"name" binding:"omitempty"`
	PermissionIDs []string `json:"permissionIds" binding:"omitempty,min=1,dive,uuid"`
}
