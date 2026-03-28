package http

import "github.com/gin-gonic/gin"

type APIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message"`
}

func OK(c *gin.Context, data any) {
	c.JSON(200, APIResponse{Success: true, Data: data, Message: "success"})
}

func Created(c *gin.Context, data any) {
	c.JSON(201, APIResponse{Success: true, Data: data, Message: "created"})
}

func BadRequest(c *gin.Context, err string) {
	c.JSON(400, APIResponse{Success: false, Message: err})
}

func Unauthorized(c *gin.Context, err string) {
	c.JSON(401, APIResponse{Success: false, Message: err})
}

func Forbidden(c *gin.Context, err string) {
	c.JSON(403, APIResponse{Success: false, Message: err})
}

func NotFound(c *gin.Context, err string) {
	c.JSON(404, APIResponse{Success: false, Message: err})
}

func Conflict(c *gin.Context, err string) {
	c.JSON(409, APIResponse{Success: false, Message: err})
}

func Internal(c *gin.Context, err string) {
	c.JSON(500, APIResponse{Success: false, Message: err})
}
