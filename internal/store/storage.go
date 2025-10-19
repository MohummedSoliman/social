// Package store for handling DB connections and db query methods.
package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("record not found")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	Posts     Posts
	Users     Users
	Comments  Comments
	Followers Followers
}

type Posts interface {
	Create(context.Context, *Post) error
	GetPostByID(context.Context, int) (*Post, error)
	DeletePostByID(context.Context, int64) error
	UpdatePost(context.Context, *Post) error
	GetUserFeed(context.Context, int64, PaginatedFeedQuery) ([]PostWithMetadata, error)
}

type Users interface {
	Create(context.Context, *sql.Tx, *User) error
	GetUserByID(context.Context, int64) (*User, error)
	CreateAndInviate(context.Context, *User, string, time.Duration) error
	ActivateUser(ctx context.Context, token string) error
	Delete(context.Context, int64) error
	GetByEmail(context.Context, string) (*User, error)
}

type Comments interface {
	GetByPostID(context.Context, int64) ([]*Comment, error)
	Create(context.Context, *Comment) error
}

type Followers interface {
	Follow(ctx context.Context, followerID, userID int64) error
	UnFollow(ctx context.Context, unfollowedID, userID int64) error
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostStore{db},
		Users:     &UserStore{db},
		Comments:  &CommentStore{db},
		Followers: &FollowerStore{db},
	}
}

func WithTransaction(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		err := tx.Rollback()
		if err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}
