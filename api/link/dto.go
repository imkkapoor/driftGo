package link

type ExchangePublicTokenCallRequest struct {
	PublicToken string `json:"public_token" validate:"required"`
}

type CreateStripeProcessorTokenCallRequest struct {
	AccountID string `json:"account_id" validate:"required"`
}
