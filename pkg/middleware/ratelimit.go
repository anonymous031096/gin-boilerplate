package middleware

import (
	"context"
	"fmt"
	"time"

	"gin-boilerplate/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimit(redisClient *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("ratelimit:%s", c.ClientIP())
		ctx := context.Background()

		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			redisClient.Expire(ctx, key, window)
		}

		if count > int64(limit) {
			c.JSON(429, gin.H{"message": "too many requests"})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", int64(limit)-count))

		c.Next()
	}
}

func RateLimitLogin(redisClient *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("ratelimit:login:%s", c.ClientIP())
		ctx := context.Background()

		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			redisClient.Expire(ctx, key, window)
		}

		if count > int64(limit) {
			response.BadRequest(c, "too many login attempts, try again later")
			c.Abort()
			return
		}

		c.Next()
	}
}
