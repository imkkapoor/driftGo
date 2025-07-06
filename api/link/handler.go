package link

import (
	"driftGo/api/common/errors"
	"driftGo/api/common/validation"
	"driftGo/domain/link"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
)

/*
Handler holds the service instance for handling Plaid link-related operations
*/
type Handler struct {
	service *link.Service
}

/*
SetupRoutes sets up the routes for the link package.
It registers the handlers for the various Plaid link-related endpoints.
*/
func SetupRoutes(r chi.Router, service *link.Service) {
	handler := &Handler{service: service}
	r.Post("/create", handler.createLinkToken)
	r.Post("/exchange", handler.exchangePublicToken)
	r.Post("/createStripeProcessorToken", handler.createStripeProcessorToken)
}

/*
createLinkToken handles the request to create a new Plaid link token.
This is used to initialize the Plaid Link interface for a user.
The link token is required to start the Plaid Link flow.
*/
func (h *Handler) createLinkToken(w http.ResponseWriter, r *http.Request) {
	linkToken, err := h.service.CreateLinkToken(r.Context())
	if err != nil {
		log.WithError(err).Error("Failed to create link token")
		errors.RequestErrorHandler(w, errors.NewErrorWithCode(http.StatusInternalServerError, "Failed to create link token", errors.ErrCodeInternalError))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(linkToken); err != nil {
		log.WithError(err).Error("Failed to encode link token response")
		errors.InternalErrorHandler(w)
		return
	}
}

/*
exchangePublicToken handles the request to exchange a public token for an access token and save it to the database.
This is used after a user successfully links their bank account through Plaid Link.
The public token is exchanged for an access token that can be used to access the user's bank account data.
*/
func (h *Handler) exchangePublicToken(w http.ResponseWriter, r *http.Request) {
	var exchangePublicTokenCallRequest ExchangePublicTokenCallRequest

	if err := json.NewDecoder(r.Body).Decode(&exchangePublicTokenCallRequest); err != nil {
		log.WithError(err).Error("Failed to decode exchange token request")
		errors.RequestErrorHandler(w, errors.NewInvalidFormatError())
		return
	}

	if !validation.ValidateRequest(w, exchangePublicTokenCallRequest) {
		return
	}

	err := h.service.ExchangePublicTokenAndSave(r.Context(), exchangePublicTokenCallRequest.PublicToken)
	if err != nil {
		log.WithError(err).Error("Failed to exchange public token and save")
		errors.InternalErrorHandler(w)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/*
createStripeProcessorToken handles the request to create a new Stripe processor token.
This is used to create a new Stripe processor token for a user's bank account.
The processor token is required to create a new Stripe customer.
*/
func (h *Handler) createStripeProcessorToken(w http.ResponseWriter, r *http.Request) {
	var createStripeProcessorTokenCallRequest CreateStripeProcessorTokenCallRequest

	if err := json.NewDecoder(r.Body).Decode(&createStripeProcessorTokenCallRequest); err != nil {
		log.WithError(err).Error("Failed to decode create stripe processor token request")
		errors.RequestErrorHandler(w, errors.NewInvalidFormatError())
		return
	}

	if !validation.ValidateRequest(w, createStripeProcessorTokenCallRequest) {
		return
	}

	accessToken, err := h.service.GetAccessTokenByAccountID(r.Context(), createStripeProcessorTokenCallRequest.AccountID)
	if err != nil {
		log.WithError(err).Error("Failed to get access token by plaid account id")
		errors.InternalErrorHandler(w)
		return
	}

	stripeProcessorToken, err := h.service.CreateStripeProcessorToken(r.Context(), accessToken, createStripeProcessorTokenCallRequest.AccountID)
	if err != nil {
		log.WithError(err).Error("Failed to create stripe processor token")
		errors.InternalErrorHandler(w)
		return
	}

	//remove after testing
	log.WithField("stripe_processor_token", stripeProcessorToken).Info("Stripe processor token created")

	w.WriteHeader(http.StatusOK)
}
