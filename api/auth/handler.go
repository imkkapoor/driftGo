package auth

import (
	"driftGo/api/common/errors"
	"driftGo/api/common/validation"
	"driftGo/domain/auth"
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
)

var decoder *schema.Decoder = schema.NewDecoder()

func init() {
	decoder.IgnoreUnknownKeys(true)
}

/*
SetupRoutes sets up the routes for the auth package.
It registers the handlers for the various auth-related endpoints.
*/
func SetupRoutes(r chi.Router, service *auth.Service) {
	handler := &Handler{service: service}
	r.Post("/create", handler.sendCreateAccountMagicLinkCall)
	r.Route("/authenticate", func(r chi.Router) {
		r.Post("/OAuth", handler.authenticateOAuthCall)
		r.Post("/magiclink", handler.authenticateMagicLinkCall)
	})
	r.Post("/setPassword", handler.setPasswordCall)
	r.Post("/login", handler.loginCall)
	r.Post("/logout", handler.logoutCall)
	r.Post("/attachOAuth", handler.attachOAuthCall)
	r.Post("/extendSession", handler.extendSessionCall)
}

/*
Handler holds the service instance
*/
type Handler struct {
	service *auth.Service
}

/*
sendCreateAccountMagicLinkCall handles the request to send a create account magic link.
This is the main entry point for sending a magic link to create an account.
It is used in the signup flow.
The request body should contain the email and code challenge.
*/
func (h *Handler) sendCreateAccountMagicLinkCall(w http.ResponseWriter, r *http.Request) {
	var sendCreateAccountMagicLinkCallRequest SendCreateAccountMagicLinkCallRequest

	if err := json.NewDecoder(r.Body).Decode(&sendCreateAccountMagicLinkCallRequest); err != nil {
		errors.RequestErrorHandler(w, errors.NewInvalidFormatError())
		return
	}

	if !validation.ValidateRequest(w, sendCreateAccountMagicLinkCallRequest) {
		return
	}

	resp, err := h.service.SendCreateAccountMagicLink(r.Context(), sendCreateAccountMagicLinkCallRequest.Email, sendCreateAccountMagicLinkCallRequest.CodeChallenge)
	if err != nil {
		log.WithError(err).Error("Failed to send create account magic link")
		errors.InternalErrorHandler(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		errors.InternalErrorHandler(w)
		return
	}
}

/*
setPasswordCall handles the request to set a password for a user.
This is used in the password set flow.
The request body should contain the password and session token.
*/
func (h *Handler) setPasswordCall(w http.ResponseWriter, r *http.Request) {
	var setPasswordBySessionCallRequest SetPasswordBySessionCallRequest

	if err := json.NewDecoder(r.Body).Decode(&setPasswordBySessionCallRequest); err != nil {
		errors.RequestErrorHandler(w, errors.NewInvalidFormatError())
		return
	}

	if !validation.ValidateRequest(w, setPasswordBySessionCallRequest) {
		return
	}

	resp, err := h.service.SetPasswordBySession(r.Context(), setPasswordBySessionCallRequest.Password, setPasswordBySessionCallRequest.SessionDurationMinutes)
	if err != nil {
		log.WithError(err).Error("Failed to set password")
		errors.RequestErrorHandler(w, errors.NewErrorWithCode(http.StatusBadRequest, "Failed to set password. Please try again.", errors.ErrCodeAuthentication))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		errors.InternalErrorHandler(w)
		return
	}
}

/*
authenticateMagicLinkCall handles the request to authenticate a magic link.
This is used in the magic link login flow.
The request body should contain the token, code verifier, and token type.
*/
func (h *Handler) authenticateMagicLinkCall(w http.ResponseWriter, r *http.Request) {
	var authenticateMagicLinkCallRequest AuthenticateMagicLinkCallRequest

	if err := json.NewDecoder(r.Body).Decode(&authenticateMagicLinkCallRequest); err != nil {
		errors.RequestErrorHandler(w, errors.NewInvalidFormatError())
		return
	}

	if !validation.ValidateRequest(w, authenticateMagicLinkCallRequest) {
		return
	}

	resp, err := h.service.AuthenticateMagicLink(r.Context(), authenticateMagicLinkCallRequest.Token, authenticateMagicLinkCallRequest.CodeVerifier)
	if err != nil {
		log.WithError(err).Error("Failed to authenticate magic link")
		errors.RequestErrorHandler(w, errors.NewErrorWithCode(http.StatusUnauthorized, "Magic link authentication failed", errors.ErrCodeAuthentication))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		errors.InternalErrorHandler(w)
		return
	}
}

/*
loginCall handles the request to log in a user.
This is used in the password login flow.
The request body should contain the email and password.
*/
func (h *Handler) loginCall(w http.ResponseWriter, r *http.Request) {
	var loginCallRequest LoginCallRequest

	if err := json.NewDecoder(r.Body).Decode(&loginCallRequest); err != nil {
		errors.RequestErrorHandler(w, errors.NewInvalidFormatError())
		return
	}

	if !validation.ValidateRequest(w, loginCallRequest) {
		return
	}

	resp, err := h.service.Login(r.Context(), loginCallRequest.Email, loginCallRequest.Password)
	if err != nil {
		log.WithError(err).Error("Failed to login")
		errors.RequestErrorHandler(w, errors.NewErrorWithCode(http.StatusUnauthorized, "Invalid email or password", errors.ErrCodeAuthentication))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		errors.InternalErrorHandler(w)
		return
	}
}

/*
logoutCall handles the request to log out a user.
This is used in the logout flow.
The request body should contain the session token.
*/
func (h *Handler) logoutCall(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.Logout(r.Context())
	if err != nil {
		log.WithError(err).Error("Failed to logout")
		errors.InternalErrorHandler(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		errors.InternalErrorHandler(w)
		return
	}
}

/*
authenticateOAuthCall handles the request to authenticate an OAuth call.
This is used in the OAuth login flow.
*/
func (h *Handler) authenticateOAuthCall(w http.ResponseWriter, r *http.Request) {
	var authenticateOAuthCallRequest AuthenticateOAuthCallRequest

	if err := json.NewDecoder(r.Body).Decode(&authenticateOAuthCallRequest); err != nil {
		errors.RequestErrorHandler(w, errors.NewInvalidFormatError())
		return
	}

	if !validation.ValidateRequest(w, authenticateOAuthCallRequest) {
		return
	}

	resp, err := h.service.AuthenticateOAuth(r.Context(), authenticateOAuthCallRequest.Token)
	if err != nil {
		log.WithError(err).Error("Failed to authenticate OAuth")
		errors.RequestErrorHandler(w, errors.NewErrorWithCode(http.StatusUnauthorized, "OAuth authentication failed", errors.ErrCodeAuthentication))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		errors.InternalErrorHandler(w)
		return
	}
}

/*
attachOAuthCall handles the request to attach an OAuth call.
This is used in the OAuth attach flow.
*/
func (h *Handler) attachOAuthCall(w http.ResponseWriter, r *http.Request) {
	var attachOAuthCallRequest AttachOAuthCallRequest

	if err := json.NewDecoder(r.Body).Decode(&attachOAuthCallRequest); err != nil {
		errors.RequestErrorHandler(w, errors.NewInvalidFormatError())
		return
	}

	if !validation.ValidateRequest(w, attachOAuthCallRequest) {
		return
	}

	resp, err := h.service.AttachOAuth(r.Context(), attachOAuthCallRequest.UserId, attachOAuthCallRequest.Provider)
	if err != nil {
		log.WithError(err).Error("Failed to attach OAuth")
		errors.InternalErrorHandler(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		errors.InternalErrorHandler(w)
		return
	}
}

/*
extendSessionCall handles the request to extend a session.
This is used in the session extension flow.
*/
func (h *Handler) extendSessionCall(w http.ResponseWriter, r *http.Request) {
	var extendSessionCallRequest ExtendSessionCallRequest

	if err := json.NewDecoder(r.Body).Decode(&extendSessionCallRequest); err != nil {
		errors.RequestErrorHandler(w, errors.NewInvalidFormatError())
		return
	}

	if !validation.ValidateRequest(w, extendSessionCallRequest) {
		return
	}

	resp, err := h.service.ExtendSession(r.Context(), extendSessionCallRequest.SessionDurationMinutes)
	if err != nil {
		log.WithError(err).Error("Failed to extend session")
		errors.InternalErrorHandler(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		errors.InternalErrorHandler(w)
		return
	}
}
