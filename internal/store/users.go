package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail    = errors.New("a user with this email is already exists")
	ErrDuplicateUsername = errors.New("a user with this username is already exists")
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"user_name"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
	RoleID    int64    `json:"role_id"`
	Role      Role     `json:"role"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

func (p *password) Compare(text string) error {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(text))
}

type UserStore struct {
	db *sql.DB
}

func (u *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `INSERT INTO users (username, email, password, role_id) VALUES ($1, $2, $3, $4)
			  RETURNING id, created_at`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	row := tx.QueryRowContext(ctx, query, user.Username, user.Email, user.Password.hash, user.RoleID)
	err := row.Scan(
		&user.ID,
		&user.CreatedAt,
	)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}
	return nil
}

func (u *UserStore) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	query := `SELECT u.id, u.username, u.email, u.password, u.created_at, r.id, r.name, r.level, r.description
			  FROM users u JOIN roles r ON u.role_id = r.id
			  WHERE u.id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var user User
	row := u.db.QueryRowContext(ctx, query, userID)
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (u *UserStore) CreateAndInviate(ctx context.Context, user *User, token string, invitationExp time.Duration) error {
	return WithTransaction(u.db, ctx, func(tx *sql.Tx) error {
		if err := u.Create(ctx, tx, user); err != nil {
			return err
		}

		if err := u.createUserInvitation(ctx, tx, user.ID, token, invitationExp); err != nil {
			return err
		}
		return nil
	})
}

func (u *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, userID int64, token string, exp time.Duration) error {
	stmt := `INSERT INTO user_invitations (token, user_id, expiry) VALUES ($1, $2, $3)`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, stmt, token, userID, time.Now().Add(exp))
	if err != nil {
		return err
	}

	return nil
}

func (u *UserStore) ActivateUser(ctx context.Context, token string) error {
	return WithTransaction(u.db, ctx, func(tx *sql.Tx) error {
		user, err := u.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}

		user.IsActive = true
		if err := u.update(ctx, tx, user); err != nil {
			return err
		}

		err = u.deleteUserInvitation(ctx, tx, user.ID)
		if err != nil {
			return err
		}

		return nil
	})
}

func (u *UserStore) deleteUserInvitation(ctx context.Context, tx *sql.Tx, userID int64) error {
	stmt := `DELETE FROM user_invitations WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, stmt, userID)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	stmt := `UPDATE users SET is_active = TRUE WHERE id = $1 AND is_active = true`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, stmt, user.ID)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `SELECT u.id, u.username, u.email, u.created_at, u.is_active
			  FROM users u JOIN user_invitations ui
			  ON u.id = ui.user_id WHERE ui.token = $1 AND ui.expiry > $2`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	hash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])

	var user User
	row := tx.QueryRowContext(ctx, query, hashedToken, time.Now())
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (u *UserStore) Delete(ctx context.Context, userID int64) error {
	return WithTransaction(u.db, ctx, func(tx *sql.Tx) error {
		err := u.deleteUser(ctx, tx, userID)
		if err != nil {
			return err
		}

		err = u.deleteUserInvitation(ctx, tx, userID)
		if err != nil {
			return err
		}

		return nil
	})
}

func (u *UserStore) deleteUser(ctx context.Context, tx *sql.Tx, userID int64) error {
	stmt := `DELETE FROM users WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := tx.ExecContext(ctx, stmt, userID)
	if err != nil {
		return err
	}

	row, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if row == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (u *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, username, email, password, created_at, is_active FROM users
			  WHERE email = $1 AND is_active = true`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var user User
	row := u.db.QueryRowContext(ctx, query, email)
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.IsActive,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}
