package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateUser(ctx context.Context, stytchUserID, firstName, lastName, email string) (*User, error) {
	var user User

	userUUID := uuid.New()

	var id int64
	err := r.pool.QueryRow(ctx, "SELECT nextval('users_id_seq')").Scan(&id)
	if err != nil {
		return nil, err
	}

	err = r.pool.QueryRow(ctx, CreateUserQuery,
		id,
		userUUID,
		stytchUserID,
		firstName,
		lastName,
		email,
	).Scan(
		&user.ID,
		&user.UUID,
		&user.StytchUserID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetUserByStytchID(ctx context.Context, stytchUserID string) (*User, error) {
	var user User

	err := r.pool.QueryRow(ctx, GetUserByStytchIDQuery, stytchUserID).Scan(
		&user.ID,
		&user.UUID,
		&user.StytchUserID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

/*
UpdateUser updates an existing user
*/
func (r *Repository) UpdateUser(ctx context.Context, stytchUserID, firstName, lastName, email string) (*User, error) {
	var user User

	err := r.pool.QueryRow(ctx, UpdateUserQuery,
		firstName,
		lastName,
		email,
		stytchUserID,
	).Scan(
		&user.ID,
		&user.UUID,
		&user.StytchUserID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

/*
DeleteUser deletes a user by their Stytch user ID
*/
func (r *Repository) DeleteUser(ctx context.Context, stytchUserID string) error {
	result, err := r.pool.Exec(ctx, DeleteUserQuery, stytchUserID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}
