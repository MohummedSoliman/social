package store

import (
	"context"
	"database/sql"
	"time"
)

func NewMockStore() Storage {
	return Storage{
		Users: &MockUserStore{},
	}
}

type MockUserStore struct{}

func (m *MockUserStore) Create(ctx context.Context, tx *sql.Tx, u *User) error {
	return nil
}

func (m *MockUserStore) GetUserByID(ctx context.Context, id int64) (*User, error) {
	return nil, nil
}

func (m *MockUserStore) CreateAndInviate(ctx context.Context, u *User, token string, exp time.Duration) error {
	return nil
}

func (m *MockUserStore) ActivateUser(ctx context.Context, token string) error {
	return nil
}

func (m *MockUserStore) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *MockUserStore) GetByEmail(ctx context.Context, emil string) (*User, error) {
	return nil, nil
}
