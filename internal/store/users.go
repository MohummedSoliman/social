package store

import (
	"context"
	"database/sql"
)

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"user_name"`
	Email     string `json:"email"`
	Password  string `json:"_"`
	CreatedAt string `json:"created_at"`
}

type UserStore struct {
	db *sql.DB
}

func (u *UserStore) Create(ctx context.Context, user *User) error {
	query := `INSERT INTO users (user_name, email, password) VALUES ($1, $2, $3) RETURNING id, created_at`

	row := u.db.QueryRowContext(ctx, query, user.Username, user.Email, user.Password)
	row.Scan(
		&user.ID,
		&user.CreatedAt,
	)

	if err := row.Err(); err != nil {
		return err
	}
	return nil
}
