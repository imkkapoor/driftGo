package link

type ExchangePublicTokenCallRequest struct {
	PublicToken string `json:"public_token" validate:"required"`
}
