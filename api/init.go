package api

import (
	"context"

	"driftGo/api/webhook"
	"driftGo/config"
	"driftGo/db"
	domauth "driftGo/domain/auth"
	domlink "driftGo/domain/link"
	"driftGo/domain/user"

	"github.com/jackc/pgx/v5/pgxpool"
)

/*
Services holds all the service instances
*/
type Services struct {
	Auth    *domauth.Service
	Link    *domlink.Service
	Webhook *webhook.WebhookHandler
	DB      *pgxpool.Pool
}

/*
InitializeServices creates and initializes all services
*/
func InitializeServices() (*Services, error) {
	pool, err := db.InitPostgres(context.Background(), config.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// Initialize User Repository
	userRepo := user.NewRepository(pool)

	// Initialize Auth Service
	authService, err := domauth.NewService(config.ProjectID, config.Secret, userRepo)
	if err != nil {
		return nil, err
	}

	// Initialize Link Service
	linkService, err := domlink.NewService(
		config.PlaidClientID,
		config.PlaidSecret,
		config.PlaidEnv,
	)
	if err != nil {
		return nil, err
	}

	// Initialize Webhook Handler
	webhookHandler := webhook.NewWebhookHandler(userRepo, config.WebhookSecret)

	return &Services{
		Auth:    authService,
		Link:    linkService,
		Webhook: webhookHandler,
		DB:      pool,
	}, nil
}
