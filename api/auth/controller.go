package auth

import (
	"driftGo/api/errors"
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"driftGo/api/common"

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
	r.Post("/logout", logoutCall)
	r.Post("/attachOAuth", attachOAuthCall)
}

/*
sendCreateAccountMagicLinkCall handles the request to send a create account magic link.
This is the main entry point for sending a magic link to create an account.
It is used in the signup flow.
The request body should contain the email and code challenge.
*/
func sendCreateAccountMagicLinkCall(w http.ResponseWriter, r *http.Request) {
	var sendCreateAccountMagicLinkCallRequest = SendCreateAccountMagicLinkCallRequest{}

	if err := json.NewDecoder(r.Body).Decode(&sendCreateAccountMagicLinkCallRequest); err != nil {
		errors.RequestErrorHandler(w, fmt.Errorf("invalid json body: %w", err))
		return
	}

	if sendCreateAccountMagicLinkCallRequest.Email == "" || sendCreateAccountMagicLinkCallRequest.CodeChallenge == "" {
		errors.RequestErrorHandler(w, fmt.Errorf("email or code challenge is missing"))
		return
	}

	resp, err := SendCreateAccountMagicLink(r.Context(), sendCreateAccountMagicLinkCallRequest)
	if err != nil {
		log.Warnf("error sending create account magic link: %v", err)
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
func setPasswordCall(w http.ResponseWriter, r *http.Request) {
	var setPasswordBySessionCallRequest = SetPasswordBySessionCallRequest{}

	if err := json.NewDecoder(r.Body).Decode(&setPasswordBySessionCallRequest); err != nil {
		errors.RequestErrorHandler(w, fmt.Errorf("invalid json body: %w", err))
		return
	}

	setPasswordBySessionCallRequest.SessionToken = common.GetSessionToken(r.Context())

	if setPasswordBySessionCallRequest.Password == "" || setPasswordBySessionCallRequest.SessionToken == "" {
		errors.RequestErrorHandler(w, fmt.Errorf("password or session token is missing"))
		return
	}

	resp, err := SetPasswordBySession(r.Context(), setPasswordBySessionCallRequest)
	if err != nil {
		log.Warnf("setting password failed:%v", err)
		errors.RequestErrorHandler(w, fmt.Errorf("setting password failed: %w", err))
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
func authenticateMagicLinkCall(w http.ResponseWriter, r *http.Request) {
	var authenticateMagicLinkCallRequest = AuthenticateMagicLinkCallRequest{}

	if err := json.NewDecoder(r.Body).Decode(&authenticateMagicLinkCallRequest); err != nil {
		errors.RequestErrorHandler(w, fmt.Errorf("invalid json body: %w", err))
		return
	}

	if authenticateMagicLinkCallRequest.Token == "" || authenticateMagicLinkCallRequest.CodeVerifier == "" || authenticateMagicLinkCallRequest.StytchTokenType != "magic_links" {
		errors.RequestErrorHandler(w, fmt.Errorf("token or code_verifier is missing or the token type is invalid"))
		return
	}

	resp, err := AuthenticateMagicLink(r.Context(), authenticateMagicLinkCallRequest)
	if err != nil {
		log.Warnf("magic link authentication failed:%v", err)
		errors.RequestErrorHandler(w, fmt.Errorf("authentication failed: %w", err))
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
func loginCall(w http.ResponseWriter, r *http.Request) {
	var loginCallRequest = LoginCallRequest{}

	if err := json.NewDecoder(r.Body).Decode(&loginCallRequest); err != nil {
		errors.RequestErrorHandler(w, fmt.Errorf("invalid json body: %w", err))
		return
	}

	if loginCallRequest.Email == "" || loginCallRequest.Password == "" {
		errors.RequestErrorHandler(w, fmt.Errorf("email or password is missing"))
		return
	}

	resp, err := Login(r.Context(), loginCallRequest)
	if err != nil {
		log.Warnf("error logging in: %v", err)
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
logoutCall handles the request to log out a user.
This is used in the logout flow.
The request body should contain the session token.
*/
func logoutCall(w http.ResponseWriter, r *http.Request) {
	var logoutCallRequest = LogoutCallRequest{}

	logoutCallRequest.SessionToken = common.GetSessionToken(r.Context())

	if logoutCallRequest.SessionToken == "" {
		errors.RequestErrorHandler(w, fmt.Errorf("session_token is missing"))
		return
	}

	resp, err := Logout(r.Context(), logoutCallRequest)
	if err != nil {
		log.Warnf("error logging out: %v", err)
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
func authenticateOAuthCall(w http.ResponseWriter, r *http.Request) {
	var authenticateOAuthCallRequest = AuthenticateOAuthCallRequest{}

	if err := json.NewDecoder(r.Body).Decode(&authenticateOAuthCallRequest); err != nil {
		errors.RequestErrorHandler(w, fmt.Errorf("invalid json body: %w", err))
		return
	}

	if authenticateOAuthCallRequest.Token == "" || authenticateOAuthCallRequest.StytchTokenType != "oauth" {
		errors.RequestErrorHandler(w, fmt.Errorf("token is missing or the token type is invalid"))
		return
	}

	resp, err := AuthenticateOAuth(r.Context(), authenticateOAuthCallRequest)
	if err != nil {
		log.Warnf("OAuth authentication failed:%v", err)
		errors.RequestErrorHandler(w, fmt.Errorf("authentication failed: %w", err))
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
func attachOAuthCall(w http.ResponseWriter, r *http.Request) {
	var attachOAuthCallRequest = AttachOAuthCallRequest{}

	if err := json.NewDecoder(r.Body).Decode(&attachOAuthCallRequest); err != nil {
		errors.RequestErrorHandler(w, fmt.Errorf("invalid json body: %w", err))
		return
	}

	attachOAuthCallRequest.SessionToken = common.GetSessionToken(r.Context())

	if attachOAuthCallRequest.Provider == "" || attachOAuthCallRequest.UserId == "" || attachOAuthCallRequest.SessionToken == "" {
		errors.RequestErrorHandler(w, fmt.Errorf("provider or user id is missing"))
		return
	}

	resp, err := AttachOAuth(r.Context(), attachOAuthCallRequest)
	if err != nil {
		log.Warnf("error attaching OAuth: %v", err)
		errors.InternalErrorHandler(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		errors.InternalErrorHandler(w)
		return
	}
}
