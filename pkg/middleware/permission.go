package middleware

import (
	"gin-boilerplate/pkg/response"

	"github.com/gin-gonic/gin"
)

func Permission(required string) gin.HandlerFunc {
	return func(c *gin.Context) {
		perms, exists := c.Get("permissions")
		if !exists {
			response.Forbidden(c, "permission denied")
			c.Abort()
			return
		}

		permissions, ok := perms.([]string)
		if !ok {
			response.Forbidden(c, "permission denied")
			c.Abort()
			return
		}

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
