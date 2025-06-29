package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

// Service handles user-related business logic using sqlc
type Service struct {
	querier Querier
}

// NewService creates a new user service
func NewService(db *pgxpool.Pool) *Service {
	return &Service{
		querier: New(db),
	}
}

// CreateUser creates a new user
func (s *Service) CreateUser(ctx context.Context, stytchUserID, firstName, lastName, email, status string) (*User, error) {
	userUUID := uuid.New()

	arg := CreateUserParams{
		Uuid:         userUUID,
		StytchUserID: stytchUserID,
		FirstName:    pgtype.Text{String: firstName, Valid: firstName != ""},
		LastName:     pgtype.Text{String: lastName, Valid: lastName != ""},
		Email:        email,
		Status:       UserStatus(status),
	}

	dbUser, err := s.querier.CreateUser(ctx, arg)
	if err != nil {
		return nil, err
	}

	return &dbUser, nil
}

// UpdateUser updates an existing user
func (s *Service) UpdateUser(ctx context.Context, stytchUserID, firstName, lastName, email, status string) (*User, error) {
	arg := UpdateUserParams{
		StytchUserID: stytchUserID,
		FirstName:    pgtype.Text{String: firstName, Valid: firstName != ""},
		LastName:     pgtype.Text{String: lastName, Valid: lastName != ""},
		Email:        email,
		Status:       UserStatus(status),
	}

	dbUser, err := s.querier.UpdateUser(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &dbUser, nil
}

// GetUserByStytchID retrieves a user by their Stytch user ID
func (s *Service) GetUserByStytchID(ctx context.Context, stytchUserID string) (*User, error) {
	dbUser, err := s.querier.GetUserByStytchID(ctx, stytchUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &dbUser, nil
}

// GetUserByID retrieves a user by their internal ID
func (s *Service) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	dbUser, err := s.querier.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &dbUser, nil
}

// GetUserByUUID retrieves a user by their UUID
func (s *Service) GetUserByUUID(ctx context.Context, userUUID uuid.UUID) (*User, error) {
	dbUser, err := s.querier.GetUserByUUID(ctx, userUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &dbUser, nil
}

// GetUserByEmail retrieves a user by their email
func (s *Service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	dbUser, err := s.querier.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &dbUser, nil
}

// DeleteUser deletes a user by their Stytch user ID
func (s *Service) DeleteUser(ctx context.Context, stytchUserID string) error {
	err := s.querier.DeleteUser(ctx, stytchUserID)
	if err != nil {
		return err
	}
	return nil
}

// UserExists checks if a user exists by their Stytch user ID
func (s *Service) UserExists(ctx context.Context, stytchUserID string) (bool, error) {
	_, err := s.querier.GetUserByStytchID(ctx, stytchUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
