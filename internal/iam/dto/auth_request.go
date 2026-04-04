package dto

type LoginRequest struct {
	Email    string `json:"email" binding:"email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"email"`
	Password string `json:"password" binding:"strong_password"`
	Name     string `json:"name" binding:"min=2,max=32"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword      string `json:"oldPassword" binding:"required"`
	NewPassword      string `json:"newPassword" binding:"required,strong_password"`
	LogoutOtherDevices bool   `json:"logoutOtherDevices"`
}
