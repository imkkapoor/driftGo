package user

import "context"

// UserInterface defines the user operations that can be used by other domains
type UserInterface interface {
	GetUserByStytchID(ctx context.Context, stytchUserID string) (*User, error)
	GetUserByID(ctx context.Context, userID int64) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, stytchUserID, firstName, lastName, email, status string) (*User, error)
	UpdateUser(ctx context.Context, stytchUserID, firstName, lastName, email, status string) (*User, error)
	DeleteUser(ctx context.Context, stytchUserID string) error
}
