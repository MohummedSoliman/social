package cache

import (
	"context"

	"github.com/MohummedSoliman/social/internal/store"
)

func NewMockCacheStorage() Storage {
	return Storage{
		Users: &mockUserStore{},
	}
}

type mockUserStore struct{}

func (m *mockUserStore) Get(ctx context.Context, id int64) (*store.User, error) {
	return nil, nil
}

func (m *mockUserStore) Set(ctx context.Context, u *store.User) error {
	return nil
}
