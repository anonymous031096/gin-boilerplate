package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

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
		headerDeviceID := c.GetHeader("X-Device-Id")
		if headerDeviceID == "" {
			headerDeviceID = "na"
		}
		if deviceID != headerDeviceID {
			response.Unauthorized(c, "device mismatch")
			c.Abort()
			return
		}

		// Check token revocation: if iat < redis timestamp, token is revoked
		revokedAt, err := redisClient.Get(context.Background(), revokeKey(userID, deviceID)).Int64()
		if err == nil && claims.IssuedAt != nil {
			if claims.IssuedAt.Time.Unix() < revokedAt {
				response.Unauthorized(c, "token has been revoked")
				c.Abort()
				return
			}
		}

		c.Set("user_id", userID)
		c.Set("device_id", deviceID)
		c.Set("roles", claims.Roles)
		c.Set("permissions", claims.AllPermissions())
		c.Next()
	}
}

// RevokeTokens saves current timestamp to redis, invalidating all tokens issued before this time
func RevokeTokens(redisClient *redis.Client, userID string, deviceID string) {
	cfg := configs.Get()
	redisClient.Set(context.Background(), revokeKey(userID, deviceID), time.Now().Unix(), cfg.JWT.RefreshTTL)
}

func revokeKey(userID string, deviceID string) string {
	return fmt.Sprintf("revoke:%s:%s", userID, deviceID)
}
