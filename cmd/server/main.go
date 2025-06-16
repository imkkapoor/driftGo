package main

import (
	"driftGo/api"
	"driftGo/config"
	"driftGo/pkg/logger"
	"net/http"

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
	log.Fatal(http.ListenAndServe(":"+port, r))
}
