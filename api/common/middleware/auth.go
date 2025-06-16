package middleware

import (
	"context"
	"net/http"
)

type contextKey string

const (
	UserIDKey       contextKey = "user_id"
	SessionTokenKey contextKey = "session_token"
)

func AuthenticateSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		// TODO: Implement session validation using auth service
		// For now, just pass through
		ctx := context.WithValue(r.Context(), SessionTokenKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
