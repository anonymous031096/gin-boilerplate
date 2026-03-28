package app

import (
	"gin-boilerplate/internal/iam"
	"gin-boilerplate/internal/product"
	"gin-boilerplate/internal/shared/auth"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func registerModules(r *gin.Engine, pg *pgxpool.Pool, jwtManager *auth.JWTManager) {
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	iamHandler := iam.NewHandler(pg, jwtManager)
	iam.RegisterRoutes(r, jwtManager, iamHandler)

	productHandler := product.NewHandler(pg)
	product.RegisterRoutes(r, jwtManager, productHandler, iamHandler)
}
