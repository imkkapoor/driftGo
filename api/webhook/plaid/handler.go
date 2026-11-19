package plaid

import (
	"context"
	"driftGo/api/common/errors"
	"driftGo/pkg/logger"
	"encoding/json"
	"io"
	"net/http"
)

/*
TransferEventProcessor syncs and applies new Plaid transfer events.
*/
type TransferEventProcessor interface {
	ProcessTransferWebhook(ctx context.Context) error
}

/*
Handler handles Plaid webhook events for transfer status updates
*/
type Handler struct {
	transferService TransferEventProcessor
	verifier        *Verifier
}

/*
NewHandler creates a new Plaid webhook handler
*/
func NewHandler(transferService TransferEventProcessor, verifier *Verifier) *Handler {
	return &Handler{
		transferService: transferService,
		verifier:        verifier,
	}
}

/*
HandleWebhook processes incoming Plaid webhook events.
For TRANSFER type webhooks with TRANSFER_EVENTS_UPDATE code,
it triggers a transfer event sync to process new transfer events.
*/
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to read Plaid webhook request body")
		errors.RequestErrorHandler(w, errors.NewInvalidFormatError())
		return
	}

	if err := h.verifier.Verify(r.Context(), r.Header.Get("Plaid-Verification"), body); err != nil {
		logger.Log.WithError(err).Warn("Rejected Plaid webhook")
		errors.UnauthorizedErrorHandler(w, "Invalid webhook signature")
		return
	}

	var event PlaidWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		logger.Log.WithError(err).Error("Failed to parse Plaid webhook event")
		errors.RequestErrorHandler(w, errors.NewInvalidFormatError())
		return
	}

	logger.Log.WithField("webhook_type", event.WebhookType).
		WithField("webhook_code", event.WebhookCode).
		Info("Incoming Plaid webhook event")

	switch event.WebhookType {
	case "TRANSFER":
		if event.WebhookCode == "TRANSFER_EVENTS_UPDATE" {
			if err := h.transferService.ProcessTransferWebhook(r.Context()); err != nil {
				logger.Log.WithError(err).Error("Failed to process transfer webhook")
				errors.InternalErrorHandler(w)
				return
			}
		} else {
			logger.Log.WithField("webhook_code", event.WebhookCode).Info("Unhandled transfer webhook code")
		}
	default:
		logger.Log.WithField("webhook_type", event.WebhookType).Info("Unhandled Plaid webhook type")
	}

	w.WriteHeader(http.StatusOK)
}
