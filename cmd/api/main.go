package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"gin-boilerplate/app"

	_ "gin-boilerplate/docs"
)

// @title           Gin Boilerplate API
// @version         1.0
// @description     Enterprise REST API boilerplate

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer {token}"

// @securityDefinitions.apikey DeviceID
// @in header
// @name X-Device-Id
// @description Device identifier

// @host localhost:8080
// @BasePath /api

func main() {
	a := app.New()
	defer a.Close()

	srv := &http.Server{
		Addr:    ":" + a.Config.Port,
		Handler: a.Router,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("server starting on port %s", a.Config.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}

	log.Println("server stopped")
}
