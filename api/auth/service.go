package auth

import (
	"context"
	"log"
	"os"

	apiRequestStructs "driftGo/api"
	"driftGo/config"

	"github.com/stytchauth/stytch-go/v16/stytch/consumer/magiclinks"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/magiclinks/email"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/passwords"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/passwords/session"
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

func SendInviteMagicLink(ctx context.Context, sendInviteMagicLinkCallRequest apiRequestStructs.SendInviteMagicLinkCallRequest) (*email.InviteResponse, error) {

	params := &email.InviteParams{
		Email:                   sendInviteMagicLinkCallRequest.Email,
		InviteMagicLinkURL:      os.Getenv("INVITE_MAGIC_LINK_URL"),
		InviteExpirationMinutes: 10,
	}

	return serviceClient.MagicLinks.Email.Invite(ctx, params)
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
