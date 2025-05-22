package api

import (
	"driftGo/api/auth"
	validateSessionMiddleware "driftGo/api/middleware"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRoutes(r *chi.Mux) chi.Router {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(validateSessionMiddleware.AuthenticateSession)
	r.Route("/auth", func(r chi.Router) {
		auth.SetupRoutes(r)
	})

	return r
}
