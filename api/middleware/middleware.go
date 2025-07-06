package middleware

import (
	"driftGo/api/common/errors"
	"driftGo/api/common/utils"
	domauth "driftGo/domain/auth"
	"driftGo/domain/user"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	publicPrefixes = []string{
		"/auth/login",
		"/auth/create",
		"/auth/authenticate/",
	}
	authService *domauth.Service
	userService user.UserInterface
)

/*
SetAuthService sets the auth service instance for the middleware
*/
func SetAuthService(service *domauth.Service) {
	authService = service
}

/*
SetUserService sets the user service instance for the middleware
*/
func SetUserService(service user.UserInterface) {
	userService = service
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
			log.Error("Missing or malformed Authorization header")
			errors.UnauthorizedErrorHandler(w, "Missing or malformed Authorization header")
			return
		}

		sessionToken := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))

		response, err := authService.AuthenticateSession(r.Context(), sessionToken)
		if err != nil {
			log.WithError(err).Error("Invalid session")
			errors.UnauthorizedErrorHandler(w, "Invalid session")
			return
		}

		// Look up the internal user ID using the Stytch user ID
		internalUser, err := userService.GetUserByStytchID(r.Context(), response.User.UserID)
		if err != nil {
			log.WithError(err).WithField("stytch_user_id", response.User.UserID).Error("Failed to find internal user")
			errors.UnauthorizedErrorHandler(w, "User not found")
			return
		}

		authContext := utils.AuthContext{
			UserID:       internalUser.ID,
			StytchUserID: response.User.UserID,
			SessionToken: sessionToken,
		}
		ctx := utils.WithAuthContext(r.Context(), authContext)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
