package auth

/*
Types for the auth package.
*/
type SendCreateAccountMagicLinkCallRequest struct {
	Email         string `json:"email"`
	CodeChallenge string `json:"code_challenge"`
}

type SetPasswordBySessionCallRequest struct {
	Password               string `json:"password"`
	SessionToken           string `json:"session_token"`
	SessionDurationMinutes int32  `json:"session_duration_minutes"`
}

type AuthenticateMagicLinkCallRequest struct {
	Token           string `json:"token"`
	StytchTokenType string `json:"stytch_token_type"`
	CodeVerifier    string `json:"code_verifier"`
}

type LoginCallRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogoutCallRequest struct {
	SessionToken string `json:"session_token"`
}

type AttachOAuthCallRequest struct {
	Provider     string `json:"provider"`
	UserId       string `json:"user_id"`
	SessionToken string `json:"session_token"`
}

type AuthenticateOAuthCallRequest struct {
	Token           string `json:"token"`
	StytchTokenType string `json:"stytch_token_type"`
}

type AuthenticateSessionCallRequest struct {
	SessionToken string `json:"session_token"`
}

type ExtendSessionCallRequest struct {
	SessionToken           string `json:"session_token"`
	SessionDurationMinutes int32  `json:"session_duration_minutes"`
}
