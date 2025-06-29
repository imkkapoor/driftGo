package main

import (
	"context"
	"driftGo/api"
	"driftGo/config"
	"driftGo/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
)

func main() {
	logger.Init(config.Env)
	var r *chi.Mux = chi.NewRouter()

	services, err := api.InitializeServices()
	if err != nil {
		log.Fatal("Failed to initialize services:", err)
	}

	api.SetupRoutes(r, services)

	port := config.Port
	if port == "" {
		port = "8080"
	}
	log.Println(`
     __    _ _____  _____   
 ___/ /___(_) _/ /_/ ___/__ 
/ _  / __/ / _/ __/ (_ / _ \
\_,_/_/ /_/_/ \__/\___/\___/                                  
   `)

	log.Println("Server running on port", port)

	// Create server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
