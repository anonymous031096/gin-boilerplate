package handler

import (
	"gin-boilerplate/internal/iam/dto"
	"gin-boilerplate/internal/iam/service"
	"gin-boilerplate/pkg/response"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	service *service.RoleService
}

func NewRoleHandler(service *service.RoleService) *RoleHandler {
	return &RoleHandler{service: service}
}

// GetByID godoc
// @Summary     Get role by ID
// @Tags        Roles
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "Role ID"
// @Success     200 {object} dto.RoleResponse
// @Failure     404 {object} response.ErrorResponse
// @Router      /roles/{id} [get]
func (h *RoleHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	role, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "role not found")
		return
	}

	response.Success(c, role)
}

// List godoc
// @Summary     List roles
// @Tags        Roles
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} dto.RoleResponse
// @Failure     400 {object} response.ErrorResponse
// @Router      /roles [get]
func (h *RoleHandler) List(c *gin.Context) {
	roles, err := h.service.List(c.Request.Context())
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, roles)
}

// Create godoc
// @Summary     Create role
// @Tags        Roles
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body dto.CreateRoleRequest true "Create role"
// @Success     200 {object} dto.RoleResponse
// @Failure     400 {object} response.ErrorResponse
// @Router      /roles [post]
func (h *RoleHandler) Create(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	createdBy := c.GetString("user_id")
	role, err := h.service.Create(c.Request.Context(), req, createdBy)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, role)
}

// Update godoc
// @Summary     Update role
// @Tags        Roles
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path string true "Role ID"
// @Param       body body dto.UpdateRoleRequest true "Update role"
// @Success     200 {object} dto.RoleResponse
// @Failure     400 {object} response.ErrorResponse
// @Router      /roles/{id} [put]
func (h *RoleHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updatedBy := c.GetString("user_id")
	role, err := h.service.Update(c.Request.Context(), id, req, updatedBy)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, role)
}

// Delete godoc
// @Summary     Delete role
// @Tags        Roles
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "Role ID"
// @Success     200 {object} map[string]bool
// @Failure     400 {object} response.ErrorResponse
// @Router      /roles/{id} [delete]
func (h *RoleHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"deleted": true})
}
