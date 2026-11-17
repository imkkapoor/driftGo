package transfer

import "time"

type InitiateTransferRequest struct {
	ReceiverUserID    int64  `json:"receiver_user_id" validate:"required"`
	SenderAccountID   string `json:"sender_account_id" validate:"required"`
	ReceiverAccountID string `json:"receiver_account_id" validate:"required"`
	Amount            string `json:"amount" validate:"required"`
	Description       string `json:"description"`
}

type TransferResponse struct {
	ID                int64      `json:"id"`
	SenderUserID      int64      `json:"sender_user_id"`
	ReceiverUserID    int64      `json:"receiver_user_id"`
	SenderAccountID   string     `json:"sender_account_id"`
	ReceiverAccountID string     `json:"receiver_account_id"`
	Amount            string     `json:"amount"`
	Currency          string     `json:"currency"`
	Status            string     `json:"status"`
	Description       string     `json:"description,omitempty"`
	CreatedAt         *time.Time `json:"created_at,omitempty"`
	UpdatedAt         *time.Time `json:"updated_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
}

type TransferListResponse struct {
	Transfers []TransferResponse `json:"transfers"`
}
