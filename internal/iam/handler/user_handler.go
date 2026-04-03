package handler

import (
	"gin-boilerplate/internal/iam/dto"
	"gin-boilerplate/internal/iam/service"
	"gin-boilerplate/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// GetByID godoc
// @Summary     Get user by ID
// @Tags        Users
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "User ID"
// @Success     200 {object} dto.UserResponse
// @Failure     404 {object} response.ErrorResponse
// @Router      /users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	user, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	response.Success(c, user)
}

// List godoc
// @Summary     List users
// @Tags        Users
// @Produce     json
// @Security    BearerAuth
// @Param       page  query int false "Page" default(1)
// @Param       limit query int false "Limit" default(20)
// @Success     200 {object} dto.UserListResponse
// @Failure     400 {object} response.ErrorResponse
// @Router      /users [get]
func (h *UserHandler) List(c *gin.Context) {
	p := response.ParsePagination(c)

	users, total, err := h.service.List(c.Request.Context(), p.Limit, p.Offset)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.List(c, users, response.PaginationMeta{
		Page:  p.Page,
		Limit: p.Limit,
		Total: total,
	})
}

// Create godoc
// @Summary     Create user
// @Tags        Users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body dto.CreateUserRequest true "Create user"
// @Success     200 {object} dto.UserResponse
// @Failure     400 {object} response.ErrorResponse
// @Router      /users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	createdBy := c.GetString("user_id")
	user, err := h.service.Create(c.Request.Context(), req, createdBy)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, user)
}

// Update godoc
// @Summary     Update user
// @Tags        Users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path string true "User ID"
// @Param       body body dto.UpdateUserRequest true "Update user"
// @Success     200 {object} dto.UserResponse
// @Failure     400 {object} response.ErrorResponse
// @Router      /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updatedBy := c.GetString("user_id")
	user, err := h.service.Update(c.Request.Context(), id, req, updatedBy)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, user)
}

// Delete godoc
// @Summary     Delete user
// @Tags        Users
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "User ID"
// @Success     200 {object} map[string]bool
// @Failure     400 {object} response.ErrorResponse
// @Router      /users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"deleted": true})
}
