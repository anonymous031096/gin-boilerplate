package middleware

import (
	"gin-boilerplate/pkg/response"

	"github.com/gin-gonic/gin"
)

func Permission(required string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions := GetCurrentPermissions(c)

		for _, p := range permissions {
			if p == required {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "permission denied")
		c.Abort()
	}
}
