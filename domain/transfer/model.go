package transfer

import (
	"context"
	"errors"
)

var (
	ErrTransferNotFound   = errors.New("transfer not found")
	ErrInsufficientFunds  = errors.New("insufficient funds for transfer")
	ErrReceiverNoAccount  = errors.New("receiver has no linked account")
	ErrTransferNotAllowed = errors.New("transfer not allowed")
	ErrSelfTransfer       = errors.New("cannot transfer to yourself")
	ErrInvalidAmount      = errors.New("invalid transfer amount")
)

/*
LinkInterface defines the methods needed from the link domain for transfer operations
*/
type LinkInterface interface {
	GetAccessTokenByAccountID(ctx context.Context, accountID string) (string, error)
	GetLinkAccountByID(ctx context.Context, ID int64) (interface{}, error)
}
