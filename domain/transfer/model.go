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
	ErrAccountNotFound    = errors.New("account not found")
	ErrAccountNotOwned    = errors.New("account does not belong to the current user")
)

/*
AccountLookup returns the owning user ID and decrypted access token for a
stored (encrypted) account ID.
*/
type AccountLookup func(ctx context.Context, encryptedAccountID string) (ownerUserID int64, accessToken string, err error)
