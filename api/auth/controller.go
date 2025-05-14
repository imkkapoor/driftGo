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
	r.Post("/create", SendMagicLinkCall)
	r.Get("/authenticate", AuthenticateMagicLinkCall)
}

func SendMagicLinkCall(w http.ResponseWriter, r *http.Request) {
	var sendMagicLinkCallRequest = api.SendMagicLinkCallRequest{}

	if err := json.NewDecoder(r.Body).Decode(&sendMagicLinkCallRequest); err != nil {
		api.RequestErrorHandler(w, fmt.Errorf("invalid json body: %w", err))
		return
	}

	if sendMagicLinkCallRequest.Email == "" {
		api.RequestErrorHandler(w, fmt.Errorf("missing email field"))
		return
	}

	// Call the service function to send the magic link.
	resp, err := SendMagicLink(r.Context(), sendMagicLinkCallRequest.Email)
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

	resp, err := AuthenticateMagicLink(r.Context(), authenticateMagicLinkCallRequest.Token)
	if err != nil {
		api.RequestErrorHandler(w, fmt.Errorf("authentication failed: %w", err))
		return
	}

	email := resp.User.Emails[0].Email
	fmt.Fprintf(w, "<h1>Welcome, %s!</h1><p>You're now logged in 🎉</p>", email)
}
