package webhook

import (
	plaidWebhook "driftGo/api/webhook/plaid"
	"driftGo/api/webhook/stytch"
	"driftGo/domain/transfer"
	"driftGo/domain/user"

	"github.com/go-chi/chi/v5"
)

/*
WebhookHandler handles webhook requests.
*/
type WebhookHandler struct {
	stytchHandler *stytch.Handler
	plaidHandler  *plaidWebhook.Handler
}

/*
NewWebhookHandler creates a new webhook handler.
*/
func NewWebhookHandler(userService *user.Service, secret string, transferService *transfer.Service, plaidKeys plaidWebhook.KeyFetcher) *WebhookHandler {
	return &WebhookHandler{
		stytchHandler: stytch.NewHandler(userService, secret),
		plaidHandler:  plaidWebhook.NewHandler(transferService, plaidWebhook.NewVerifier(plaidKeys)),
	}
}

/*
SetupRoutes sets up the webhook routes.
*/
func SetupRoutes(r chi.Router, handler *WebhookHandler) {
	r.Post("/stytch", handler.stytchHandler.HandleWebhook)
	r.Post("/plaid", handler.plaidHandler.HandleWebhook)
}
