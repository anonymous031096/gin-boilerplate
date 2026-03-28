package app

import (
	"context"
	"net/http"

	"gin-boilerplate/configs"
	"gin-boilerplate/internal/shared/auth"
	"gin-boilerplate/internal/shared/cache"
	"gin-boilerplate/internal/shared/db"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type App struct {
	cfg         configs.Config
	router      *gin.Engine
	server      *http.Server
	pg          *pgxpool.Pool
	redisClient *redis.Client
	jwtManager  *auth.JWTManager
}

func Build(ctx context.Context) (*App, error) {
	cfg, err := configs.Load()
	if err != nil {
		return nil, err
	}

	pg, err := db.NewPostgresPool(ctx, cfg.PostgresDSN)
	if err != nil {
		return nil, err
	}

	redisClient, err := cache.NewRedisClient(ctx, cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		pg.Close()
		return nil, err
	}

	jwtManager := auth.NewJWTManager(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessTTLMin, cfg.JWTRefreshTTLHour)
	router := gin.Default()
	registerModules(router, pg, jwtManager)

	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}

	return &App{
		cfg:         cfg,
		router:      router,
		server:      server,
		pg:          pg,
		redisClient: redisClient,
		jwtManager:  jwtManager,
	}, nil
}

func (a *App) Run() error {
	return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	_ = a.server.Shutdown(ctx)
	_ = a.redisClient.Close()
	a.pg.Close()
	return nil
}
