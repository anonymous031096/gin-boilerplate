package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gin-boilerplate/internal/app"
	_ "gin-boilerplate/docs"

	"github.com/joho/godotenv"
)

// @title Gin Boilerplate API
// @version 1.0
// @description Enterprise boilerplate API with IAM and Product modules.
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter JWT token as: Bearer {token}
func main() {
	_ = godotenv.Load()

	application, err := app.Build(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := application.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := application.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
