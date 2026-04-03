package todo

import (
	"gin-boilerplate/internal/todo/handler"
	"gin-boilerplate/internal/todo/service"
	"gin-boilerplate/pkg/deps"
)

type Module struct {
	TodoHandler *handler.TodoHandler
}

func NewModule(d *deps.Deps) *Module {
	todoService := service.NewTodoService(d.DB)

	return &Module{
		TodoHandler: handler.NewTodoHandler(todoService),
	}
}
