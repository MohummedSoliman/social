package store

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx/v5"
	"github.com/lib/pq"
)

type Post struct {
	ID        int64      `json:"id"`
	Content   string     `json:"content"`
	Title     string     `json:"title"`
	UserID    int64      `json:"user_id"`
	Tags      []string   `json:"tags"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_At"`
	Comments  []*Comment `json:"comments"`
}

type PostStore struct {
	db *sql.DB
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `INSERT INTO posts (content, title, user_id, tags)
			  VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`

	row := s.db.QueryRowContext(ctx, query, post.Content, post.Title, post.UserID, pq.Array(post.Tags))
	row.Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err := row.Err(); err != nil {
		return err
	}
	return nil
}

func (s *PostStore) GetPostByID(ctx context.Context, postID int) (*Post, error) {
	query := `SELECT id, content, title, user_id, tags, created_at, updated_at FROM posts
			  WHERE id = $1`
	var post Post

	err := s.db.QueryRowContext(ctx, query, postID).Scan(
		&post.ID,
		&post.Content,
		&post.Title,
		&post.UserID,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &post, nil
}

func (s *PostStore) DeletePostByID(ctx context.Context, postID int64) error {
	// tx, err := s.db.BeginTx(ctx, nil)
	// if err != nil {
	// 	return err
	// }

	// comments := &CommentStore{db: s.db}
	// err = comments.DeleteCommentsByPostID(ctx, postID)
	// if err != nil {
	// 	tx.Rollback()
	// 	return err
	// }

	query := `DELETE FROM posts WHERE id = $1`
	res, err := s.db.ExecContext(ctx, query, postID)
	if err != nil {
		// tx.Rollback()
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	// if err := tx.Commit(); err != nil {
	// 	return err
	// }
	return nil
}

func (s *PostStore) UpdatePost(ctx context.Context, post *Post) error {
	query := `UPDATE posts SET content = $1 , title = $2 WHERE id = $3`

	res, err := s.db.ExecContext(ctx, query, post.Content, post.Title, post.ID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
