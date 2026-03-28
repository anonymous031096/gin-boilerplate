package iam

import (
	"gin-boilerplate/internal/shared/auth"
	"gin-boilerplate/internal/shared/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, jwtManager *auth.JWTManager, h *Handler) {
	r.POST("/auth/signup", h.SignUp)
	r.POST("/auth/signin", h.SignIn)
	r.POST("/auth/refresh", h.Refresh)

	authGroup := r.Group("", middleware.AuthMiddleware(jwtManager))
	authGroup.GET("/users/me", middleware.RequirePermission(h, "users:read-me"), h.GetMe)
	authGroup.GET("/users", middleware.RequirePermission(h, "users:read"), h.ListUsers)
	authGroup.GET("/users/:id", middleware.RequirePermission(h, "users:read"), h.GetUserByID)
	authGroup.PUT("/users/:id", middleware.RequirePermission(h, "users:update"), h.UpdateUser)

	authGroup.POST("/roles", middleware.RequirePermission(h, "roles:create"), h.CreateRole)
	authGroup.GET("/roles", middleware.RequirePermission(h, "roles:read"), h.ListRoles)
	authGroup.PUT("/roles/:id", middleware.RequirePermission(h, "roles:update"), h.UpdateRole)
	authGroup.DELETE("/roles/:id", middleware.RequirePermission(h, "roles:delete"), h.DeleteRole)

	authGroup.GET("/permissions", middleware.RequirePermission(h, "permissions:read"), h.ListPermissions)
	authGroup.GET("/permissions/:id", middleware.RequirePermission(h, "permissions:read"), h.GetPermissionByID)
}
