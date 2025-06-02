package auth

/*
Types for the auth package.
*/
type SendCreateAccountMagicLinkCallRequest struct {
	Email         string `json:"email" validate:"required,email"`
	CodeChallenge string `json:"code_challenge" validate:"required"`
}

type SetPasswordBySessionCallRequest struct {
	Password               string `json:"password" validate:"required"`
	SessionDurationMinutes int32  `json:"session_duration_minutes" validate:"required,min=5,max=525600"`
}

type AuthenticateMagicLinkCallRequest struct {
	Token           string `json:"token" validate:"required"`
	StytchTokenType string `json:"stytch_token_type" validate:"required,oneof=magic_links"`
	CodeVerifier    string `json:"code_verifier" validate:"required"`
}

type LoginCallRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AttachOAuthCallRequest struct {
	Provider string `json:"provider" validate:"required"`
	UserId   string `json:"user_id" validate:"required"`
}

type AuthenticateOAuthCallRequest struct {
	Token           string `json:"token" validate:"required"`
	StytchTokenType string `json:"stytch_token_type" validate:"required,oneof=oauth"`
}

type ExtendSessionCallRequest struct {
	SessionDurationMinutes int32 `json:"session_duration_minutes" validate:"required,min=5,max=525600"`
}
