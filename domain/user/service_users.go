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

/*
Service handles user-related business logic using sqlc
*/
type Service struct {
	database Querier
}

/*
NewService creates a new user service
*/
func NewService(db *pgxpool.Pool) *Service {
	return &Service{
		database: New(db),
	}
}

/*
exec
*/
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

	dbUser, err := s.database.CreateUser(ctx, arg)
	if err != nil {
		return nil, err
	}

	return &dbUser, nil
}

/*
exec
*/
func (s *Service) UpdateUser(ctx context.Context, stytchUserID, firstName, lastName, email, status string) (*User, error) {
	arg := UpdateUserParams{
		StytchUserID: stytchUserID,
		FirstName:    pgtype.Text{String: firstName, Valid: firstName != ""},
		LastName:     pgtype.Text{String: lastName, Valid: lastName != ""},
		Email:        email,
		Status:       UserStatus(status),
	}

	dbUser, err := s.database.UpdateUser(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &dbUser, nil
}

/*
returns one
*/
func (s *Service) GetUserByStytchID(ctx context.Context, stytchUserID string) (*User, error) {
	dbUser, err := s.database.GetUserByStytchID(ctx, stytchUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &dbUser, nil
}

/*
returns one
*/
func (s *Service) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	dbUser, err := s.database.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &dbUser, nil
}

/*
returns one
*/
func (s *Service) GetUserByUUID(ctx context.Context, userUUID uuid.UUID) (*User, error) {
	dbUser, err := s.database.GetUserByUUID(ctx, userUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &dbUser, nil
}

/*
returns one
*/
func (s *Service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	dbUser, err := s.database.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &dbUser, nil
}

/*
exec
*/
func (s *Service) DeleteUser(ctx context.Context, stytchUserID string) error {
	err := s.database.DeleteUser(ctx, stytchUserID)
	if err != nil {
		return err
	}
	return nil
}

/*
returns bool
*/
func (s *Service) UserExists(ctx context.Context, stytchUserID string) (bool, error) {
	_, err := s.database.GetUserByStytchID(ctx, stytchUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
