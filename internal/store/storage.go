// Package store for handling DB connections and db query methods.
package store

import (
	"context"
	"database/sql"
)

type Storage struct {
	Posts Posts
	Users Users
}

type Posts interface {
	Create(context.Context, *Post) error
}

type Users interface {
	Create(context.Context, *User) error
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts: &PostStore{db},
		Users: &UserStore{db},
	}
}
