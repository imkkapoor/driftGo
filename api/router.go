package api

import (
	"driftGo/api/auth"
	"driftGo/api/link"
	validateSessionMiddleware "driftGo/api/middleware"
	"driftGo/api/transfer"
	"driftGo/api/user"
	"driftGo/api/webhook"
	"driftGo/config"
	"driftGo/pkg/logger"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func SetupRoutes(r *chi.Mux, services *Services) chi.Router {
	r.Use(logger.RequestLogger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   config.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true, // Required for cookies
		MaxAge:           300,  // Maximum value not ignored by any of major browsers
	}))

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

		// Setup user routes
		protected.Route("/user", func(r chi.Router) {
			user.SetupRoutes(r, services.User)
		})

		// Setup transfer routes
		protected.Route("/transfer", func(r chi.Router) {
			transfer.SetupRoutes(r, services.Transfer)
		})
	})

	return r
}
