package handler

import (
	_ "gin-boilerplate/internal/iam/dto"
	"gin-boilerplate/internal/iam/service"
	"gin-boilerplate/pkg/response"

	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	service *service.PermissionService
}

func NewPermissionHandler(service *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{service: service}
}

// List godoc
// @Summary     List all permissions
// @Tags        Permissions
// @Produce     json
// @Security    BearerAuth
// @Security    DeviceID
// @Success     200 {array} dto.PermissionResponse
// @Failure     400 {object} response.ErrorResponse
// @Router      /permissions [get]
func (h *PermissionHandler) List(c *gin.Context) {
	permissions, err := h.service.List(c.Request.Context())
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, permissions)
}
