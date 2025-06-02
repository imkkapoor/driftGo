package api

import (
	"driftGo/api/auth"
	"driftGo/api/link"
	validateSessionMiddleware "driftGo/middleware"
	"driftGo/pkg/logger"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRoutes(r *chi.Mux, services *Services) chi.Router {
	r.Use(logger.RequestLogger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Initialize auth service in middleware
	validateSessionMiddleware.SetAuthService(services.Auth)
	r.Use(validateSessionMiddleware.AuthenticateSession)

	// Setup auth routes
	r.Route("/auth", func(r chi.Router) {
		auth.SetupRoutes(r, services.Auth)
	})

	// Setup link routes
	r.Route("/link", func(r chi.Router) {
		link.SetupRoutes(r, services.Link)
	})

	return r
}
