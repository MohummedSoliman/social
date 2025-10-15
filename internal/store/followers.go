package store

import (
	"context"
	"database/sql"
	"time"
)

type Follower struct {
	UserID     int64  `json:"user_id"`
	FollowerID int64  `json:"follower_id"`
	CreatedAt  string `json:"created_at"`
}

type FollowerStore struct {
	db *sql.DB
}

func (f *FollowerStore) Follow(ctx context.Context, followerID int64, userID int64) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	stmt := `INSERT INTO followers (user_id, follower_id)
			 VALUES ($1, $2)`

	res, err := f.db.ExecContext(ctx, stmt, userID, followerID)
	if err != nil {
		return err
	}

	row, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if row == 0 {
		return err
	}

	return nil
}

func (f *FollowerStore) UnFollow(ctx context.Context, unfollowedID int64, userID int64) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	stmt := `DELETE FROM followers WHERE user_id = $1 AND follower_id = $2`

	res, err := f.db.ExecContext(ctx, stmt, userID, unfollowedID)
	if err != nil {
		return err
	}

	row, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if row == 0 {
		return err
	}

	return nil
}
