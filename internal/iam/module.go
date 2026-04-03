package iam

import (
	"gin-boilerplate/internal/iam/handler"
	"gin-boilerplate/internal/iam/service"
	"gin-boilerplate/pkg/deps"
)

type Module struct {
	AuthHandler *handler.AuthHandler
	UserHandler *handler.UserHandler
	RoleHandler *handler.RoleHandler
}

func NewModule(d *deps.Deps) *Module {
	authService := service.NewAuthService(d.DB, d.Redis)
	userService := service.NewUserService(d.DB)
	roleService := service.NewRoleService(d.DB)

	return &Module{
		AuthHandler: handler.NewAuthHandler(authService, userService),
		UserHandler: handler.NewUserHandler(userService),
		RoleHandler: handler.NewRoleHandler(roleService),
	}
}
