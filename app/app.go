package app

import (
	"database/sql"

	"gin-boilerplate/configs"
	"gin-boilerplate/pkg/cache"
	"gin-boilerplate/pkg/db"
	"gin-boilerplate/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type App struct {
	Config configs.Config
	DB     *sql.DB
	Redis  *redis.Client
	Router *gin.Engine
}

func New() *App {
	cfg := configs.Load()

	database := db.NewPostgres(cfg.Postgres.DSN())
	redisClient := cache.NewRedis(cfg.Redis.Addr())

	router := gin.Default()

	// CORS
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Device-Id")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Rate limit
	// router.Use(middleware.RateLimit(redisClient, 100, 1*time.Minute))

	// Validator
	validator.Init()

	app := &App{
		Config: cfg,
		DB:     database,
		Redis:  redisClient,
		Router: router,
	}

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	app.registerModules()

	return app
}

func (a *App) Run() error {
	return a.Router.Run(":" + a.Config.Port)
}

func (a *App) Close() {
	a.DB.Close()
	a.Redis.Close()
}
