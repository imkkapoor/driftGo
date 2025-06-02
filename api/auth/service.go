package auth

import (
	"context"

	"driftGo/api/common/utils"
	"driftGo/config"

	"github.com/stytchauth/stytch-go/v16/stytch/consumer/magiclinks"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/magiclinks/email"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/oauth"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/passwords"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/passwords/session"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/sessions"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/stytchapi"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/users"
)

// Service handles all auth-related operations
type Service struct {
	client *stytchapi.API
}

// NewService creates a new auth service
func NewService(projectID, secret string) (*Service, error) {
	client, err := stytchapi.NewClient(projectID, secret)
	if err != nil {
		return nil, err
	}
	return &Service{client: client}, nil
}

func (s *Service) SendCreateAccountMagicLink(ctx context.Context, sendCreateAccountMagicLinkCallRequest SendCreateAccountMagicLinkCallRequest) (*email.LoginOrCreateResponse, error) {
	params := &email.LoginOrCreateParams{
		Email:                   sendCreateAccountMagicLinkCallRequest.Email,
		CodeChallenge:           sendCreateAccountMagicLinkCallRequest.CodeChallenge,
		CreateUserAsPending:     true,
		SignupMagicLinkURL:      config.SignupMagicLinkURL,
		SignupExpirationMinutes: 10,
	}

	return s.client.MagicLinks.Email.LoginOrCreate(ctx, params)
}

func (s *Service) SetPasswordBySession(ctx context.Context, setPasswordBySessionCallRequest SetPasswordBySessionCallRequest) (*session.ResetResponse, error) {
	params := &session.ResetParams{
		Password:               setPasswordBySessionCallRequest.Password,
		SessionToken:           utils.GetSessionToken(ctx),
		SessionDurationMinutes: setPasswordBySessionCallRequest.SessionDurationMinutes,
	}

	return s.client.Passwords.Sessions.Reset(ctx, params)
}

func (s *Service) AuthenticateMagicLink(ctx context.Context, authenticateMagicLinkCallRequest AuthenticateMagicLinkCallRequest) (*magiclinks.AuthenticateResponse, error) {
	params := &magiclinks.AuthenticateParams{
		Token:                  authenticateMagicLinkCallRequest.Token,
		CodeVerifier:           authenticateMagicLinkCallRequest.CodeVerifier,
		SessionDurationMinutes: 60,
	}

	return s.client.MagicLinks.Authenticate(ctx, params)
}

func (s *Service) Login(ctx context.Context, loginCallRequest LoginCallRequest) (*passwords.AuthenticateResponse, error) {
	params := &passwords.AuthenticateParams{
		Email:                  loginCallRequest.Email,
		Password:               loginCallRequest.Password,
		SessionDurationMinutes: 60,
	}

	return s.client.Passwords.Authenticate(ctx, params)
}

func (s *Service) Logout(ctx context.Context) (*sessions.RevokeResponse, error) {
	params := &sessions.RevokeParams{
		SessionToken: utils.GetSessionToken(ctx),
	}

	return s.client.Sessions.Revoke(ctx, params)
}

func (s *Service) AttachOAuth(ctx context.Context, attachOAuthCallRequest AttachOAuthCallRequest) (*oauth.AttachResponse, error) {
	params := &oauth.AttachParams{
		UserID:   attachOAuthCallRequest.UserId,
		Provider: attachOAuthCallRequest.Provider,
	}

	return s.client.OAuth.Attach(ctx, params)
}

func (s *Service) AuthenticateOAuth(ctx context.Context, authenticateOAuthCallRequest AuthenticateOAuthCallRequest) (*oauth.AuthenticateResponse, error) {
	params := &oauth.AuthenticateParams{
		Token: authenticateOAuthCallRequest.Token,
	}

	return s.client.OAuth.Authenticate(ctx, params)
}

func (s *Service) AuthenticateSession(ctx context.Context, sessionToken string) (*sessions.AuthenticateResponse, error) {
	params := &sessions.AuthenticateParams{
		SessionToken: sessionToken,
	}

	return s.client.Sessions.Authenticate(ctx, params)
}

func (s *Service) ExtendSession(ctx context.Context, extendSessionCallRequest ExtendSessionCallRequest) (*sessions.AuthenticateResponse, error) {
	params := &sessions.AuthenticateParams{
		SessionToken:           utils.GetSessionToken(ctx),
		SessionDurationMinutes: extendSessionCallRequest.SessionDurationMinutes,
	}

	return s.client.Sessions.Authenticate(ctx, params)
}

func (s *Service) GetUser(ctx context.Context) (*users.GetResponse, error) {
	params := &users.GetParams{
		UserID: utils.GetUserID(ctx),
	}

	return s.client.Users.Get(ctx, params)
}
