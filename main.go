package main

import (
	"driftGo/api/auth"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"
)

func main() {

	log.SetReportCaller(true)
	var r *chi.Mux = chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/auth", func(r chi.Router) {
		auth.SetupRoutes(r)
	})

	port := os.Getenv("PORT")
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
