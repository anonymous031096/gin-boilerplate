package dto

type CreateUserRequest struct {
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=6"`
	Name     string   `json:"name" binding:"required"`
	RoleIDs  []string `json:"roleIds" binding:"required,min=1,dive,uuid"`
}

type UpdateUserRequest struct {
	Name    string   `json:"name" binding:"omitempty"`
	RoleIDs []string `json:"roleIds" binding:"omitempty,dive,uuid"`
}
