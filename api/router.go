package api

import (
	"driftGo/api/auth"
	"driftGo/api/link"
	validateSessionMiddleware "driftGo/api/middleware"
	"driftGo/api/webhook"
	"driftGo/pkg/logger"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRoutes(r *chi.Mux, services *Services) chi.Router {
	r.Use(logger.RequestLogger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Setup webhook routes
	r.Route("/webhook", func(r chi.Router) {
		webhook.SetupRoutes(r, services.Webhook)
	})

	r.Group(func(protected chi.Router) {
		validateSessionMiddleware.SetAuthService(services.Auth)
		validateSessionMiddleware.SetUserService(services.User)
		protected.Use(validateSessionMiddleware.AuthenticateSession)

		// Setup auth routes
		protected.Route("/auth", func(r chi.Router) {
			auth.SetupRoutes(r, services.Auth)
		})

		// Setup link routes
		protected.Route("/link", func(r chi.Router) {
			link.SetupRoutes(r, services.Link)
		})
	})

	return r
}
