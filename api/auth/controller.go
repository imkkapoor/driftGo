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

/*
SetupRoutes sets up the routes for the auth package.
It registers the handlers for the various auth-related endpoints.
*/
func SetupRoutes(r chi.Router) {
	r.Post("/create", sendCreateAccountMagicLinkCall)
	r.Route("/authenticate", func(r chi.Router) {
		r.Post("/OAuth", authenticateOAuthCall)
		r.Post("/magiclink", authenticateMagicLinkCall)
	})
	r.Post("/setPassword", setPasswordCall)
	r.Post("/login", loginCall)
	r.Post("/attachOAuth", attachOAuthCall)
}

/*
sendCreateAccountMagicLinkCall handles the request to send a create account magic link.
This is the main entry point for sending a magic link to create an account.
It is used in the signup flow.
The request body should contain the email and code challenge.
*/
func sendCreateAccountMagicLinkCall(w http.ResponseWriter, r *http.Request) {
	var sendCreateAccountMagicLinkCallRequest = api.SendCreateAccountMagicLinkCallRequest{}

	if err := json.NewDecoder(r.Body).Decode(&sendCreateAccountMagicLinkCallRequest); err != nil {
		api.RequestErrorHandler(w, fmt.Errorf("invalid json body: %w", err))
		return
	}

	if sendCreateAccountMagicLinkCallRequest.Email == "" || sendCreateAccountMagicLinkCallRequest.CodeChallenge == "" {
		api.RequestErrorHandler(w, fmt.Errorf("email or code challenge is missing"))
		return
	}

	resp, err := SendCreateAccountMagicLink(r.Context(), sendCreateAccountMagicLinkCallRequest)
	if err != nil {
		log.Printf("error sending create account magic link: %v", err)
		api.InternalErrorHandler(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		api.InternalErrorHandler(w)
		return
	}
}

/*
setPasswordCall handles the request to set a password for a user.
This is used in the password set flow.
The request body should contain the password and session token.
*/
func setPasswordCall(w http.ResponseWriter, r *http.Request) {
	var setPasswordBySessionCallRequest = api.SetPasswordBySessionCallRequest{}

	if err := json.NewDecoder(r.Body).Decode(&setPasswordBySessionCallRequest); err != nil {
		api.RequestErrorHandler(w, fmt.Errorf("invalid json body: %w", err))
		return
	}

	if setPasswordBySessionCallRequest.Password == "" || setPasswordBySessionCallRequest.SessionToken == "" {
		api.RequestErrorHandler(w, fmt.Errorf("password or session token is missing"))
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

/*
authenticateMagicLinkCall handles the request to authenticate a magic link.
This is used in the magic link login flow.
*/
func authenticateMagicLinkCall(w http.ResponseWriter, r *http.Request) {
	var authenticateMagicLinkCallRequest = api.AuthenticateMagicLinkCallRequest{}

	if err := decoder.Decode(&authenticateMagicLinkCallRequest, r.URL.Query()); err != nil {
		api.RequestErrorHandler(w, fmt.Errorf("invalid query parameters: %w", err))
		return
	}

	if authenticateMagicLinkCallRequest.StytchTokenType != "magiclink" {
		api.RequestErrorHandler(w, fmt.Errorf("invalid token type"))
		return
	}

	if authenticateMagicLinkCallRequest.Token == "" || authenticateMagicLinkCallRequest.CodeVerifier == "" {
		api.RequestErrorHandler(w, fmt.Errorf("token is missing"))
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

/*
loginCall handles the request to log in a user.
This is used in the password login flow.
The request body should contain the email and password.
*/
func loginCall(w http.ResponseWriter, r *http.Request) {
	var loginCallRequest = api.LoginCallRequest{}

	if err := json.NewDecoder(r.Body).Decode(&loginCallRequest); err != nil {
		api.RequestErrorHandler(w, fmt.Errorf("invalid json body: %w", err))
		return
	}

	if loginCallRequest.Email == "" || loginCallRequest.Password == "" {
		api.RequestErrorHandler(w, fmt.Errorf("email or password is missing"))
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

/*
authenticateOAuthCall handles the request to authenticate an OAuth call.
This is used in the OAuth login flow.
*/
func authenticateOAuthCall(w http.ResponseWriter, r *http.Request) {
	var authenticateOAuthCallRequest = api.AuthenticateOAuthCallRequest{}

	if err := decoder.Decode(&authenticateOAuthCallRequest, r.URL.Query()); err != nil {
		api.RequestErrorHandler(w, fmt.Errorf("invalid query parameters: %w", err))
		return
	}

	if authenticateOAuthCallRequest.StytchTokenType != "oauth" {
		api.RequestErrorHandler(w, fmt.Errorf("invalid token type"))
		return
	}

	if authenticateOAuthCallRequest.Token == "" || authenticateOAuthCallRequest.StytchTokenType == "" {
		api.RequestErrorHandler(w, fmt.Errorf("token is missing"))
		return
	}

	resp, err := AuthenticateOAuth(r.Context(), authenticateOAuthCallRequest)
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

/*
attachOAuthCall handles the request to attach an OAuth call.
This is used in the OAuth attach flow.
*/
func attachOAuthCall(w http.ResponseWriter, r *http.Request) {
	var attachOAuthCallRequest = api.AttachOAuthCallRequest{}

	if err := json.NewDecoder(r.Body).Decode(&attachOAuthCallRequest); err != nil {
		api.RequestErrorHandler(w, fmt.Errorf("invalid json body: %w", err))
		return
	}

	if attachOAuthCallRequest.Provider == "" || attachOAuthCallRequest.UserId == "" {
		api.RequestErrorHandler(w, fmt.Errorf("provider or user id is missing"))
		return
	}

	resp, err := AttachOAuth(r.Context(), attachOAuthCallRequest)
	if err != nil {
		log.Printf("error attaching OAuth: %v", err)
		api.InternalErrorHandler(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		api.InternalErrorHandler(w)
		return
	}
}
