package user

const (
	CreateUserQuery = `
		INSERT INTO users (id, uuid, stytch_user_id, first_name, last_name, email)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, uuid, stytch_user_id, first_name, last_name, email, created_at, updated_at
	`

	GetUserByStytchIDQuery = `
		SELECT id, uuid, stytch_user_id, first_name, last_name, email, created_at, updated_at
		FROM users
		WHERE stytch_user_id = $1
	`

	UpdateUserQuery = `
		UPDATE users
		SET first_name = $1,
			last_name = $2,
			email = $3,
			updated_at = CURRENT_TIMESTAMP
		WHERE stytch_user_id = $4
		RETURNING id, uuid, stytch_user_id, first_name, last_name, email, created_at, updated_at
	`

	DeleteUserQuery = `
		DELETE FROM users
		WHERE stytch_user_id = $1
	`
)
