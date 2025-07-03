package user

import "context"

type UserInterface interface {
	GetUserByStytchID(ctx context.Context, stytchUserID string) (*User, error)
	GetUserByID(ctx context.Context, userID int64) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}
