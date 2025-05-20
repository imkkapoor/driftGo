package api

import (
	"encoding/json"
	"net/http"
)

type SendCreateAccountMagicLinkCallRequest struct {
	Email         string `json:"email"`
	CodeChallenge string `json:"code_challenge"`
}

type SetPasswordBySessionCallRequest struct {
	Password     string `json:"password"`
	SessionToken string `json:"session_token"`
}

type AuthenticateMagicLinkCallRequest struct {
	Token           string `json:"token"`
	StytchTokenType string `json:"stytch_token_type,omitempty"`
	CodeVerifier    string `json:"code_verifier"`
}

type LoginCallRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AttachOAuthCallRequest struct {
	Provider     string `json:"provider"`
	UserId       string `json:"user_id"`
	SessionToken string `json:"session_token"`
}

type AuthenticateOAuthCallRequest struct {
	Token           string `schema:"token"`
	StytchTokenType string `schema:"stytch_token_type,omitempty"`
	CodeVerifier    string `schema:"code_verifier"`
}

type Error struct {
	Code    int
	Message string
}

func writeError(w http.ResponseWriter, message string, code int) {
	resp := Error{
		Code:    code,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	json.NewEncoder(w).Encode(resp)
}

var (
	RequestErrorHandler = func(w http.ResponseWriter, err error) {
		writeError(w, err.Error(), http.StatusBadRequest)
	}
	InternalErrorHandler = func(w http.ResponseWriter) {
		writeError(w, "An Unexpected Error Occurred.", http.StatusInternalServerError)
	}
)
