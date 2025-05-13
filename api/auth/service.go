package auth

import (
	"context"
	"log"
	"os"

	"driftGo/config"

	"github.com/stytchauth/stytch-go/v16/stytch/consumer/magiclinks"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/magiclinks/email"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/stytchapi"
)

var serviceClient *stytchapi.API

func init() {
	var err error
	serviceClient, err = stytchapi.NewClient(config.ProjectID, config.Secret)
	if err != nil {
		log.Fatalf("failed to initialize Stytch client in service: %v", err)
	}
}

// SendMagicLink sends a magic link to the provided email address.
func SendMagicLink(ctx context.Context, emailAddr string) (*email.SendResponse, error) {
	params := &email.SendParams{
		Email:             emailAddr,
		LoginMagicLinkURL: os.Getenv("STYTCH_LOGIN_REDIRECT_URL"),
	}
	return serviceClient.MagicLinks.Email.Send(ctx, params)
}

func AuthenticateMagicLink(ctx context.Context, token string) (*magiclinks.AuthenticateResponse, error) {
	params := &magiclinks.AuthenticateParams{
		Token: token,
	}
	return serviceClient.MagicLinks.Authenticate(ctx, params)
}
