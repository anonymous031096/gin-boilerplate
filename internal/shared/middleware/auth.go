package middleware

import (
	"strings"

	"gin-boilerplate/internal/shared/auth"
	httpresp "gin-boilerplate/internal/shared/http"

	"github.com/gin-gonic/gin"
)

const userIDKey = "userID"

func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			httpresp.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			httpresp.Unauthorized(c, "invalid authorization format")
			c.Abort()
			return
		}
		claims, err := jwtManager.ParseAccessToken(parts[1])
		if err != nil {
			httpresp.Unauthorized(c, "invalid or expired access token")
			c.Abort()
			return
		}
		c.Set(userIDKey, claims.UserID)
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) (string, bool) {
	v, ok := c.Get(userIDKey)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}
