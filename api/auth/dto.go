package auth

type SendCreateAccountMagicLinkCallRequest struct {
	Email         string `json:"email"`
	CodeChallenge string `json:"code_challenge"`
}

type SetPasswordBySessionCallRequest struct {
	Password               string `json:"password"`
	SessionDurationMinutes int32  `json:"session_duration_minutes"`
}

type AuthenticateMagicLinkCallRequest struct {
	Token        string `json:"token"`
	CodeVerifier string `json:"code_verifier"`
}

type LoginCallRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AttachOAuthCallRequest struct {
	UserId   string `json:"user_id"`
	Provider string `json:"provider"`
}

type AuthenticateOAuthCallRequest struct {
	Token string `json:"token"`
}

type ExtendSessionCallRequest struct {
	SessionDurationMinutes int32 `json:"session_duration_minutes"`
}
