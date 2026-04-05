package iam

import (
	"gin-boilerplate/pkg/deps"
	"gin-boilerplate/pkg/middleware"
)

func RegisterRoutes(d *deps.Deps) {
	// cfg := configs.Get()
	m := NewModule(d)
	auth := middleware.Auth(d.Redis)
	perm := middleware.Permission

	// Public
	d.Router.POST("/auth/login", m.AuthHandler.Login)
	d.Router.POST("/auth/register", m.AuthHandler.Register)
	d.Router.POST("/auth/refresh", m.AuthHandler.RefreshToken)

	// Auth — authenticated
	d.Router.PUT("/auth/change-password", auth, m.AuthHandler.ChangePassword)

	// Users — auth + permission
	users := d.Router.Group("/users", auth)
	{
		users.GET("/me", m.UserHandler.Me)
		users.GET("", perm("user:read"), m.UserHandler.List)
		users.GET("/:id", perm("user:read"), m.UserHandler.GetByID)
		users.POST("", perm("user:create"), m.UserHandler.Create)
		users.PUT("/:id", perm("user:update"), m.UserHandler.Update)
		users.DELETE("/:id", perm("user:delete"), m.UserHandler.Delete)
	}

	// Permissions — auth + permission
	d.Router.GET("/permissions", auth, perm("permission:read"), m.PermissionHandler.List)

	// Roles — auth + permission
	roles := d.Router.Group("/roles", auth)
	{
		roles.GET("", perm("role:read"), m.RoleHandler.List)
		roles.GET("/:id", perm("role:read"), m.RoleHandler.GetByID)
		roles.POST("", perm("role:create"), m.RoleHandler.Create)
		roles.PUT("/:id", perm("role:update"), m.RoleHandler.Update)
		roles.DELETE("/:id", perm("role:delete"), m.RoleHandler.Delete)
	}
}
