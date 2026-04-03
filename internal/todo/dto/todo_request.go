package dto

type CreateTodoRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"omitempty"`
}

type UpdateTodoRequest struct {
	Title       string `json:"title" binding:"omitempty"`
	Description string `json:"description" binding:"omitempty"`
	Completed   *bool  `json:"completed" binding:"omitempty"`
}
