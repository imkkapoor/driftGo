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

/*
Service handles all auth-related operations
*/
type Service struct {
	client *stytchapi.API
}

/*
NewService creates a new auth service
*/
func NewService(projectID, secret string) (*Service, error) {
	client, err := stytchapi.NewClient(projectID, secret)
	if err != nil {
		return nil, err
	}
	return &Service{client: client}, nil
}

func (s *Service) SendCreateAccountMagicLink(ctx context.Context, userEmail, codeChallenge string) (*email.LoginOrCreateResponse, error) {
	params := &email.LoginOrCreateParams{
		Email:                   userEmail,
		CodeChallenge:           codeChallenge,
		CreateUserAsPending:     true,
		SignupMagicLinkURL:      config.SignupMagicLinkURL,
		SignupExpirationMinutes: 10,
	}

	return s.client.MagicLinks.Email.LoginOrCreate(ctx, params)
}

func (s *Service) SetPasswordBySession(ctx context.Context, password string, sessionDurationMinutes int32) (*session.ResetResponse, error) {
	params := &session.ResetParams{
		Password:               password,
		SessionToken:           utils.GetSessionToken(ctx),
		SessionDurationMinutes: sessionDurationMinutes,
	}

	return s.client.Passwords.Sessions.Reset(ctx, params)
}

func (s *Service) AuthenticateMagicLink(ctx context.Context, token, codeVerifier string) (*magiclinks.AuthenticateResponse, error) {
	params := &magiclinks.AuthenticateParams{
		Token:                  token,
		CodeVerifier:           codeVerifier,
		SessionDurationMinutes: 60,
	}

	return s.client.MagicLinks.Authenticate(ctx, params)
}

func (s *Service) Login(ctx context.Context, userEmail, password string) (*passwords.AuthenticateResponse, error) {
	params := &passwords.AuthenticateParams{
		Email:                  userEmail,
		Password:               password,
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

func (s *Service) AttachOAuth(ctx context.Context, userID, provider string) (*oauth.AttachResponse, error) {
	params := &oauth.AttachParams{
		UserID:   userID,
		Provider: provider,
	}

	return s.client.OAuth.Attach(ctx, params)
}

func (s *Service) AuthenticateOAuth(ctx context.Context, token string) (*oauth.AuthenticateResponse, error) {
	params := &oauth.AuthenticateParams{
		Token: token,
	}

	return s.client.OAuth.Authenticate(ctx, params)
}

func (s *Service) AuthenticateSession(ctx context.Context, sessionToken string) (*sessions.AuthenticateResponse, error) {
	params := &sessions.AuthenticateParams{
		SessionToken: sessionToken,
	}

	return s.client.Sessions.Authenticate(ctx, params)
}

func (s *Service) ExtendSession(ctx context.Context, sessionDurationMinutes int32) (*sessions.AuthenticateResponse, error) {
	params := &sessions.AuthenticateParams{
		SessionToken:           utils.GetSessionToken(ctx),
		SessionDurationMinutes: sessionDurationMinutes,
	}

	return s.client.Sessions.Authenticate(ctx, params)
}

func (s *Service) GetUser(ctx context.Context) (*users.GetResponse, error) {
	params := &users.GetParams{
		UserID: utils.GetUserID(ctx),
	}

	return s.client.Users.Get(ctx, params)
}
