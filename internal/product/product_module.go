package product

import (
	"gin-boilerplate/internal/shared/auth"
	"gin-boilerplate/internal/shared/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, jwtManager *auth.JWTManager, h *Handler, permissionSvc middleware.PermissionService) {
	g := r.Group("/products", middleware.AuthMiddleware(jwtManager))
	g.POST("", middleware.RequirePermission(permissionSvc, "products:create"), h.Create)
	g.GET("", middleware.RequirePermission(permissionSvc, "products:read"), h.List)
	g.GET("/:id", middleware.RequirePermission(permissionSvc, "products:read"), h.GetByID)
	g.PUT("/:id", middleware.RequirePermission(permissionSvc, "products:update"), h.Update)
	g.DELETE("/:id", middleware.RequirePermission(permissionSvc, "products:delete"), h.Delete)
}
