package link

import (
	"time"
)

/*
Types for the link package.
*/
type ExchangePublicTokenCallRequest struct {
	PublicToken string `json:"public_token" validate:"required"`
}

type LinkTokenCallResponse struct {
	LinkToken  string    `json:"link_token"`
	Expiration time.Time `json:"expiration"`
	RequestID  string    `json:"request_id"`
}

type AccessTokenCallResponse struct {
	AccessToken string `json:"access_token"`
	ItemID      string `json:"item_id"`
	RequestID   string `json:"request_id"`
}
