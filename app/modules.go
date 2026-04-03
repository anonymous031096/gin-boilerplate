package app

import (
	"gin-boilerplate/internal/iam"
	"gin-boilerplate/internal/todo"
	"gin-boilerplate/pkg/deps"
)

func (a *App) registerModules() {
	d := &deps.Deps{
		DB:     a.DB,
		Redis:  a.Redis,
		Router: a.Router.Group("/api"),
	}

	iam.RegisterRoutes(d)
	todo.RegisterRoutes(d)
}
