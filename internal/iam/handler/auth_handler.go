package handler

import (
	"gin-boilerplate/internal/iam/dto"
	"gin-boilerplate/internal/iam/service"
	"gin-boilerplate/pkg/middleware"
	"gin-boilerplate/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
	userService *service.UserService
}

func NewAuthHandler(authService *service.AuthService, userService *service.UserService) *AuthHandler {
	return &AuthHandler{authService: authService, userService: userService}
}

// Login godoc
// @Summary     Login
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Security    DeviceID
// @Param       body body dto.LoginRequest true "Login"
// @Success     200 {object} dto.TokenResponse
// @Failure     401 {object} response.ErrorResponse
// @Router      /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	deviceID := middleware.GetDeviceID(c)
	token, err := h.authService.Login(c.Request.Context(), req, deviceID)
	if err != nil {
		response.Unauthorized(c, "invalid email or password")
		return
	}

	response.Success(c, token)
}

// Register godoc
// @Summary     Register
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Security    DeviceID
// @Param       body body dto.RegisterRequest true "Register"
// @Success     200
// @Failure     400 {object} response.ErrorResponse
// @Router      /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.authService.Register(c.Request.Context(), req); err != nil {
		response.HandleError(c, err)
		return
	}

	c.Status(200)
}

// RefreshToken godoc
// @Summary     Refresh token
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Security    DeviceID
// @Param       body body dto.RefreshTokenRequest true "Refresh token"
// @Success     200 {object} dto.TokenResponse
// @Failure     401 {object} response.ErrorResponse
// @Router      /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	deviceID := middleware.GetDeviceID(c)
	token, err := h.authService.RefreshToken(c.Request.Context(), req, deviceID)
	if err != nil {
		response.Unauthorized(c, "invalid refresh token")
		return
	}

	response.Success(c, token)
}

// ChangePassword godoc
// @Summary     Change password
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Security    DeviceID
// @Param       body body dto.ChangePasswordRequest true "Change password"
// @Success     200 {object} dto.TokenResponse "If logoutAllDevices=true, returns new tokens"
// @Failure     400 {object} response.ErrorResponse
// @Router      /auth/change-password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	userID := middleware.GetCurrentUserID(c)
	deviceID := middleware.GetDeviceID(c)
	token, err := h.authService.ChangePassword(c.Request.Context(), userID, deviceID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	if token != nil {
		response.Success(c, token)
		return
	}

	c.Status(200)
}

// Me godoc
// @Summary     Get current user
// @Tags        Auth
// @Produce     json
// @Security    BearerAuth
// @Security    DeviceID
// @Success     200 {object} dto.UserResponse
// @Failure     404 {object} response.ErrorResponse
// @Router      /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)

	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	response.Success(c, user)
}
