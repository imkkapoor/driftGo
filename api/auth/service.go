package auth

import (
	"context"
	"log"
	"os"

	apiRequestStructs "driftGo/api"
	"driftGo/config"

	"github.com/stytchauth/stytch-go/v16/stytch/consumer/magiclinks"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/magiclinks/email"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/oauth"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/passwords"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/passwords/session"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/stytchapi"
)

/*
Logic for setting up the Stytch client and sending requests to the Stytch API.
*/
var serviceClient *stytchapi.API

func init() {
	var err error
	serviceClient, err = stytchapi.NewClient(config.ProjectID, config.Secret)
	if err != nil {
		log.Fatalf("failed to initialize Stytch client in service: %v", err)
	}
}

func SendCreateAccountMagicLink(ctx context.Context, sendCreateAccountMagicLinkCallRequest apiRequestStructs.SendCreateAccountMagicLinkCallRequest) (*email.LoginOrCreateResponse, error) {

	params := &email.LoginOrCreateParams{
		Email:                   sendCreateAccountMagicLinkCallRequest.Email,
		CodeChallenge:           sendCreateAccountMagicLinkCallRequest.CodeChallenge,
		CreateUserAsPending:     true,
		SignupMagicLinkURL:      os.Getenv("STYTCH_SIGNUP_REDIRECT_URL"),
		SignupExpirationMinutes: 10,
	}

	return serviceClient.MagicLinks.Email.LoginOrCreate(ctx, params)
}

func SetPasswordBySession(ctx context.Context, setPasswordBySessionCallRequest apiRequestStructs.SetPasswordBySessionCallRequest) (*session.ResetResponse, error) {

	params := &session.ResetParams{
		Password:               setPasswordBySessionCallRequest.Password,
		SessionToken:           setPasswordBySessionCallRequest.SessionToken,
		SessionDurationMinutes: 60,
	}

	return serviceClient.Passwords.Sessions.Reset(ctx, params)
}

func AuthenticateMagicLink(ctx context.Context, authenticateMagicLinkCallRequest apiRequestStructs.AuthenticateMagicLinkCallRequest) (*magiclinks.AuthenticateResponse, error) {

	params := &magiclinks.AuthenticateParams{
		Token:                  authenticateMagicLinkCallRequest.Token,
		CodeVerifier:           authenticateMagicLinkCallRequest.CodeVerifier,
		SessionDurationMinutes: 60,
	}

	return serviceClient.MagicLinks.Authenticate(ctx, params)
}

func Login(ctx context.Context, loginCallRequest apiRequestStructs.LoginCallRequest) (*passwords.AuthenticateResponse, error) {

	params := &passwords.AuthenticateParams{
		Email:                  loginCallRequest.Email,
		Password:               loginCallRequest.Password,
		SessionDurationMinutes: 60,
	}

	return serviceClient.Passwords.Authenticate(ctx, params)
}

func AttachOAuth(ctx context.Context, attachOathCallRequest apiRequestStructs.AttachOathCallRequest) (*oauth.AttachResponse, error) {

	params := &oauth.AttachParams{
		UserID:   attachOathCallRequest.UserId,
		Provider: attachOathCallRequest.Provider,
	}

	return serviceClient.OAuth.Attach(ctx, params)
}

func AuthenticateOAuth(ctx context.Context, authenticateOAuthCallRequest apiRequestStructs.AuthenticateOAuthCallRequest) (*oauth.AuthenticateResponse, error) {

	params := &oauth.AuthenticateParams{
		Token:        authenticateOAuthCallRequest.Token,
		CodeVerifier: authenticateOAuthCallRequest.CodeVerifier,
	}

	return serviceClient.OAuth.Authenticate(ctx, params)
}
