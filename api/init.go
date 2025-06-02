package api

import (
	"driftGo/api/auth"
	"driftGo/api/link"
	"driftGo/config"
)

// Services holds all the service instances
type Services struct {
	Auth *auth.Service
	Link *link.Service
}

// InitializeServices creates and initializes all services
func InitializeServices() (*Services, error) {
	// Initialize Auth Service
	authService, err := auth.NewService(config.ProjectID, config.Secret)
	if err != nil {
		return nil, err
	}

	// Initialize Link Service
	linkService, err := link.NewService(
		config.PlaidClientID,
		config.PlaidSecret,
		config.PlaidEnv,
	)
	if err != nil {
		return nil, err
	}

	return &Services{
		Auth: authService,
		Link: linkService,
	}, nil
}
