package middleware

import (
	"context"
	"driftGo/api/auth"
	"driftGo/api/common"
	"driftGo/api/errors"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

var publicPrefixes = []string{
	"/auth/login",
	"/auth/create",
	"/auth/authenticate/",
}

func isPublicRoute(path string) bool {
	for _, prefix := range publicPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

/*
AuthenticateSession is a middleware that checks if the request has a valid session token.
If the token is valid, it adds the session token to the request context.
If the token is invalid or missing, it returns a 401 Unauthorized error.
This middleware is used to protect routes that require authentication.
*/
func AuthenticateSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if isPublicRoute(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			log.Warn("missing or malformed Authorization header")
			errors.UnauthorizedErrorHandler(w, "Missing or malformed Authorization header")
			return
		}

		sessionToken := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))

		var err error
		_, err = auth.AuthenticateSession(r.Context(), sessionToken)
		if err != nil {
			log.Warnf("invalid session: %v", err)
			errors.UnauthorizedErrorHandler(w, "Invalid session")
			return
		}

		ctx := context.WithValue(r.Context(), common.SessionTokenKey, sessionToken)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
