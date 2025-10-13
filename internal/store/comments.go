package store

import (
	"context"
	"database/sql"
	"errors"
)

type Comment struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	PostID    int64  `json:"post_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	User      User   `json:"user"`
}

type CommentStore struct {
	db *sql.DB
}

func (c *CommentStore) GetByPostID(ctx context.Context, postID int64) ([]*Comment, error) {
	query := `SELECT c.id, c.user_id, c.post_id, c.content, c.created_at, users.id, users.username FROM commnets c
			  JOIN users ON c.user_id = users.id
			  WHERE c.post_id = $1 ORDER BY c.created_at DESC`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var comments []*Comment
	rows, err := c.db.QueryContext(ctx, query, postID)
	if err != nil {
		return comments, err
	}
	defer rows.Close()

	for rows.Next() {
		var c Comment
		c.User = User{}
		err := rows.Scan(
			&c.ID,
			&c.UserID,
			&c.PostID,
			&c.Content,
			&c.CreatedAt,
			&c.User.ID,
			&c.User.Username,
		)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return comments, ErrNotFound
			default:
				return comments, err
			}
		}
		comments = append(comments, &c)
	}
	return comments, nil
}

func (c *CommentStore) DeleteCommentsByPostID(ctx context.Context, postID int64) error {
	query := `DELETE FROM comments WHERE post_id := $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := c.db.ExecContext(ctx, query, postID)
	if err != nil {
		return err
	}
	return nil
}

func (c *CommentStore) Create(ctx context.Context, comment *Comment) error {
	stmt := `INSERT INTO comments (user_id, post_id, content)
			 VALUES ($1, $2, $3)`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := c.db.ExecContext(ctx, stmt, comment.UserID, comment.PostID, comment.Content)
	if err != nil {
		return err
	}

	return nil
}
