package deps

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Deps struct {
	DB     *sql.DB
	Redis  *redis.Client
	Router *gin.RouterGroup
}
