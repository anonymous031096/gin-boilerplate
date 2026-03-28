package middleware

import (
	"context"

	httpresp "gin-boilerplate/internal/shared/http"

	"github.com/gin-gonic/gin"
)

type PermissionService interface {
	HasPermission(ctx context.Context, userID string, permissionCode string) (bool, error)
}

func RequirePermission(permissionSvc PermissionService, code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := CurrentUserID(c)
		if !ok {
			httpresp.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}
		has, err := permissionSvc.HasPermission(c, userID, code)
		if err != nil {
			httpresp.Internal(c, "failed to check permission")
			c.Abort()
			return
		}
		if !has {
			httpresp.Forbidden(c, "insufficient permission")
			c.Abort()
			return
		}
		c.Next()
	}
}
