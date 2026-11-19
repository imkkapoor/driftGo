package api

import (
	"driftGo/api/webhook"
	"driftGo/config"
	"driftGo/db"
	authDomain "driftGo/domain/auth"
	linkDomain "driftGo/domain/link"
	transferDomain "driftGo/domain/transfer"
	userDomain "driftGo/domain/user"
)

/*
Services holds all the service instances
*/
type Services struct {
	Auth     *authDomain.Service
	Link     *linkDomain.Service
	Transfer *transferDomain.Service
	User     *userDomain.Service
	Webhook  *webhook.WebhookHandler
}

/*
InitializeServices creates and initializes all services
*/
func InitializeServices() (*Services, error) {
	// Initialize database
	pool := db.InitDB()

	// Initialize User Service
	userService := userDomain.NewService(pool)

	// Initialize Auth Service
	authService, err := authDomain.NewService(config.ProjectID, config.Secret)
	if err != nil {
		return nil, err
	}

	// Initialize Link Service
	linkService, err := linkDomain.NewService(
		config.PlaidClientID,
		config.PlaidSecret,
		config.PlaidEnv,
		config.PlaidWebhookURL,
		userService,
		pool,
		config.EncryptionKey,
	)
	if err != nil {
		return nil, err
	}

	// Initialize Transfer Service
	transferService, err := transferDomain.NewService(
		config.PlaidClientID,
		config.PlaidSecret,
		config.PlaidEnv,
		userService,
		linkService.GetAccountOwnerAndAccessToken,
		pool,
		config.EncryptionKey,
	)
	if err != nil {
		return nil, err
	}

	// Initialize Webhook Handler
	webhookHandler := webhook.NewWebhookHandler(userService, config.WebhookSecret, transferService)

	return &Services{
		Auth:     authService,
		Link:     linkService,
		Transfer: transferService,
		User:     userService,
		Webhook:  webhookHandler,
	}, nil
}
