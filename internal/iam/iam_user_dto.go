package iam

type UpdateUserRequest struct {
	FullName string `json:"fullName" binding:"required"`
}
