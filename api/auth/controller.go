package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"driftGo/api"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
)

var decoder *schema.Decoder = schema.NewDecoder()

func init() {
	decoder.IgnoreUnknownKeys(true)
}

func SetupRoutes(r chi.Router) {
	r.Post("/invite", SendInviteMagicLinkCall)
	r.Post("/setPassword", SetPasswordCall)
	r.Post("/login", LoginCall)
	r.Get("/authenticate", AuthenticateMagicLinkCall)
}

func SendInviteMagicLinkCall(w http.ResponseWriter, r *http.Request) {
	var sendInviteMagicLinkCallRequest = api.SendInviteMagicLinkCallRequest{}

	if err := json.NewDecoder(r.Body).Decode(&sendInviteMagicLinkCallRequest); err != nil {
		api.RequestErrorHandler(w, fmt.Errorf("invalid json body: %w", err))
		return
	}

	if sendInviteMagicLinkCallRequest.Email == "" {
		api.RequestErrorHandler(w, fmt.Errorf("missing email field"))
		return
	}

	resp, err := SendInviteMagicLink(r.Context(), sendInviteMagicLinkCallRequest)
	if err != nil {
		log.Printf("error sending magic link: %v", err)
		api.InternalErrorHandler(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		api.InternalErrorHandler(w)
		return
	}
}

func SetPasswordCall(w http.ResponseWriter, r *http.Request) {
	var setPasswordBySessionCallRequest = api.SetPasswordBySessionCallRequest{}

	if err := json.NewDecoder(r.Body).Decode(&setPasswordBySessionCallRequest); err != nil {
		api.RequestErrorHandler(w, fmt.Errorf("invalid json body: %w", err))
		return
	}

	if setPasswordBySessionCallRequest.Password == "" {
		api.RequestErrorHandler(w, fmt.Errorf("missing password field"))
		return
	}

	resp, err := SetPasswordBySession(r.Context(), setPasswordBySessionCallRequest)
	if err != nil {
		api.RequestErrorHandler(w, fmt.Errorf("setting password failed: %w", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		api.InternalErrorHandler(w)
		return
	}
}
func AuthenticateMagicLinkCall(w http.ResponseWriter, r *http.Request) {
	var authenticateMagicLinkCallRequest = api.AuthenticateMagicLinkCallRequest{}

	if err := decoder.Decode(&authenticateMagicLinkCallRequest, r.URL.Query()); err != nil {
		api.RequestErrorHandler(w, fmt.Errorf("invalid query parameters: %w", err))
		return
	}

	if authenticateMagicLinkCallRequest.Token == "" {
		api.RequestErrorHandler(w, fmt.Errorf("token is required"))
		return
	}

	resp, err := AuthenticateMagicLink(r.Context(), authenticateMagicLinkCallRequest)
	if err != nil {
		api.RequestErrorHandler(w, fmt.Errorf("authentication failed: %w", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		api.InternalErrorHandler(w)
		return
	}
}

func LoginCall(w http.ResponseWriter, r *http.Request) {
	var loginCallRequest = api.LoginCallRequest{}

	if err := json.NewDecoder(r.Body).Decode(&loginCallRequest); err != nil {
		api.RequestErrorHandler(w, fmt.Errorf("invalid json body: %w", err))
		return
	}

	if loginCallRequest.Email == "" {
		api.RequestErrorHandler(w, fmt.Errorf("missing email field"))
		return
	}

	if loginCallRequest.Password == "" {
		api.RequestErrorHandler(w, fmt.Errorf("missing password field"))
		return
	}

	resp, err := Login(r.Context(), loginCallRequest)
	if err != nil {
		log.Printf("error logging in: %v", err)
		api.InternalErrorHandler(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		api.InternalErrorHandler(w)
		return
	}
}
