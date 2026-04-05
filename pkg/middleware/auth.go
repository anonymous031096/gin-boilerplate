package middleware

import (
	"fmt"
	"strings"

	"gin-boilerplate/configs"
	"gin-boilerplate/pkg/auth"
	"gin-boilerplate/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func Auth(redisClient *redis.Client) gin.HandlerFunc {
	cfg := configs.Get()

	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "invalid authorization format")
			c.Abort()
			return
		}

		claims, err := auth.ParseAccessToken(cfg.JWT.AccessSecret, parts[1])
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		userID := claims.GetUserID()
		deviceID := claims.GetDeviceID()

		// Check device ID matches X-Device-Id header
		headerDeviceID := GetDeviceFingerprint(c)
		if deviceID != headerDeviceID {
			response.Unauthorized(c, "device mismatch")
			c.Abort()
			return
		}

		// Check token revocation (in-memory, no Redis call)
		if claims.IssuedAt != nil && IsRevoked(userID, deviceID, claims.IssuedAt.Time.Unix()) {
			response.Unauthorized(c, "token has been revoked")
			c.Abort()
			return
		}

		SetCurrentUser(c, userID, deviceID, claims.AllPermissions())
		c.Next()
	}
}

// RevokeTokens revokes tokens for a specific user+device.
func RevokeTokens(redisClient *redis.Client, userID string, deviceID string) {
	publishRevoke(redisClient, revokeKey(userID, deviceID))
}

// RevokeAllTokens revokes tokens on ALL devices for a user.
func RevokeAllTokens(redisClient *redis.Client, userID string) {
	publishRevoke(redisClient, revokeAllKey(userID))
}

func revokeKey(userID string, deviceID string) string {
	return fmt.Sprintf("revoke:%s:%s", userID, deviceID)
}

func revokeAllKey(userID string) string {
	return fmt.Sprintf("revoke:%s:*", userID)
}
