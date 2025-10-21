package cache

import (
	"context"

	"github.com/MohummedSoliman/social/internal/store"
	"github.com/go-redis/redis/v8"
)

type Storage struct {
	Users Users
}

func NewRedisStorage(rdb *redis.Client) Storage {
	return Storage{
		Users: &UserStore{rdb},
	}
}

type Users interface {
	Get(context.Context, int64) (*store.User, error)
	Set(context.Context, *store.User) error
}
