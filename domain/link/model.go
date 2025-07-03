package link

import (
	"time"
)

/*
Response structs
*/
type LinkTokenCallResponse struct {
	LinkToken  string    `json:"link_token"`
	Expiration time.Time `json:"expiration"`
	RequestID  string    `json:"request_id"`
}

type AccessTokenCallResponse struct {
	AccessToken string `json:"access_token"`
	ItemID      string `json:"item_id"`
}
