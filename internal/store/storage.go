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
	Create(context.Context, *User) error
	GetUserByID(context.Context, int64) (*User, error)
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
