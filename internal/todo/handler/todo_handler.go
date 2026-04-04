package handler

import (
	"gin-boilerplate/internal/todo/dto"
	"gin-boilerplate/internal/todo/service"
	"gin-boilerplate/pkg/middleware"
	"gin-boilerplate/pkg/response"

	"github.com/gin-gonic/gin"
)

type TodoHandler struct {
	service *service.TodoService
}

func NewTodoHandler(service *service.TodoService) *TodoHandler {
	return &TodoHandler{service: service}
}

// GetByID godoc
// @Summary     Get todo by ID
// @Tags        Todos
// @Produce     json
// @Security    BearerAuth
// @Security    DeviceID
// @Param       id path string true "Todo ID"
// @Success     200 {object} dto.TodoResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Router      /todos/{id} [get]
func (h *TodoHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetCurrentUserID(c)

	todo, err := h.service.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		if err.Error() == "forbidden: not owner" {
			response.Forbidden(c, "you can only access your own todos")
			return
		}
		response.NotFound(c, "todo not found")
		return
	}

	response.Success(c, todo)
}

// List godoc
// @Summary     List my todos
// @Tags        Todos
// @Produce     json
// @Security    BearerAuth
// @Security    DeviceID
// @Param       page  query int false "Page" default(1)
// @Param       limit query int false "Limit" default(20)
// @Success     200 {object} dto.TodoListResponse
// @Failure     400 {object} response.ErrorResponse
// @Router      /todos [get]
func (h *TodoHandler) List(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	p := response.ParsePagination(c)

	todos, total, err := h.service.List(c.Request.Context(), userID, p.Limit, p.Offset)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.List(c, todos, response.PaginationMeta{
		Page:  p.Page,
		Limit: p.Limit,
		Total: total,
	})
}

// Create godoc
// @Summary     Create todo
// @Tags        Todos
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Security    DeviceID
// @Param       body body dto.CreateTodoRequest true "Create todo"
// @Success     200 {object} dto.TodoResponse
// @Failure     400 {object} response.ErrorResponse
// @Router      /todos [post]
func (h *TodoHandler) Create(c *gin.Context) {
	var req dto.CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	userID := middleware.GetCurrentUserID(c)
	todo, err := h.service.Create(c.Request.Context(), req, userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, todo)
}

// Update godoc
// @Summary     Update todo
// @Tags        Todos
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Security    DeviceID
// @Param       id   path string true "Todo ID"
// @Param       body body dto.UpdateTodoRequest true "Update todo"
// @Success     200 {object} dto.TodoResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     400 {object} response.ErrorResponse
// @Router      /todos/{id} [put]
func (h *TodoHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	userID := middleware.GetCurrentUserID(c)
	todo, err := h.service.Update(c.Request.Context(), id, req, userID)
	if err != nil {
		if err.Error() == "forbidden: not owner" {
			response.Forbidden(c, "you can only update your own todos")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, todo)
}

// Delete godoc
// @Summary     Delete todo
// @Tags        Todos
// @Produce     json
// @Security    BearerAuth
// @Security    DeviceID
// @Param       id path string true "Todo ID"
// @Success     200 {object} map[string]bool
// @Failure     403 {object} response.ErrorResponse
// @Failure     400 {object} response.ErrorResponse
// @Router      /todos/{id} [delete]
func (h *TodoHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetCurrentUserID(c)

	if err := h.service.Delete(c.Request.Context(), id, userID); err != nil {
		if err.Error() == "forbidden: not owner" {
			response.Forbidden(c, "you can only delete your own todos")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"deleted": true})
}
