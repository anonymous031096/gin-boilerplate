package todo

import (
	"gin-boilerplate/pkg/deps"
	"gin-boilerplate/pkg/middleware"
)

func RegisterRoutes(d *deps.Deps) {
	m := NewModule(d)
	auth := middleware.Auth(d.Redis)
	perm := middleware.Permission

	todos := d.Router.Group("/todos", auth)
	{
		todos.GET("", perm("todo:read"), m.TodoHandler.List)
		todos.GET("/:id", perm("todo:read"), m.TodoHandler.GetByID)
		todos.POST("", perm("todo:create"), m.TodoHandler.Create)
		todos.PUT("/:id", perm("todo:update"), m.TodoHandler.Update)
		todos.DELETE("/:id", perm("todo:delete"), m.TodoHandler.Delete)
	}
}
