package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MohummedSoliman/social/internal/store"
	"github.com/go-redis/redis/v8"
)

type UserStore struct {
	db *redis.Client
}

const UserExpTime = time.Minute

func (u *UserStore) Get(ctx context.Context, userID int64) (*store.User, error) {
	cacheKey := fmt.Sprintf("user-%v", userID)
	data, err := u.db.Get(ctx, cacheKey).Result()
	if err != nil {
		switch err {
		case redis.Nil:
			return nil, nil
		default:
			return nil, err
		}
	}

	var user store.User
	if data != "" {
		err := json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (u *UserStore) Set(ctx context.Context, user *store.User) error {
	cacheKey := fmt.Sprintf("user-%v", user.ID)

	json, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return u.db.SetEX(ctx, cacheKey, json, UserExpTime).Err()
}
