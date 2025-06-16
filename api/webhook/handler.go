package webhook

import (
	"driftGo/api/webhook/stytch"
	"driftGo/domain/user"

	"github.com/go-chi/chi/v5"
)

/*
WebhookHandler handles webhook requests.
*/
type WebhookHandler struct {
	stytchHandler *stytch.Handler
}

/*
NewWebhookHandler creates a new webhook handler.
*/
func NewWebhookHandler(userRepo *user.Repository, secret string) *WebhookHandler {
	return &WebhookHandler{
		stytchHandler: stytch.NewHandler(userRepo, secret),
	}
}

/*
SetupRoutes sets up the webhook routes.
*/
func SetupRoutes(r chi.Router, handler *WebhookHandler) {
	r.Post("/stytch", handler.stytchHandler.HandleWebhook)
}
