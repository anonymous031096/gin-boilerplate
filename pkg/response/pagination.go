package response

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type PaginationMeta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type PaginationParams struct {
	Page   int
	Limit  int
	Offset int
}

func ParsePagination(c *gin.Context) PaginationParams {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return PaginationParams{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}
