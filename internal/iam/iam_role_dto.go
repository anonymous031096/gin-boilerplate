package iam

type CreateRoleRequest struct {
	Code          string   `json:"code" binding:"required"`
	Name          string   `json:"name" binding:"required"`
	PermissionIDs []string `json:"permissionIds"`
}

type UpdateRoleRequest struct {
	Name          string   `json:"name" binding:"required"`
	PermissionIDs []string `json:"permissionIds"`
}
