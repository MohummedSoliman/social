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
	Version   int        `json:"version"`
	User      User       `json:"user"`
}

type PostWithMetadata struct {
	Post         Post `json:"post"`
	CommentCount int  `json:"comments_count"`
}

type PostStore struct {
	db *sql.DB
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `INSERT INTO posts (content, title, user_id, tags)
			  VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

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
	query := `SELECT id, content, title, user_id, tags, created_at, updated_at, version FROM posts
			  WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var post Post

	err := s.db.QueryRowContext(ctx, query, postID).Scan(
		&post.ID,
		&post.Content,
		&post.Title,
		&post.UserID,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Version,
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

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

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
	query := `UPDATE posts SET content = $1 , title = $2, version = version + 1
			  WHERE id = $3 AND version = $4 RETURNING version`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, post.Content, post.Title, post.ID, post.Version).Scan(&post.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}

func (s *PostStore) GetUserFeed(ctx context.Context, userID int64, fq PaginatedFeedQuery) ([]PostWithMetadata, error) {
	query := `SELECT p.id, p.title, p.content, p.user_id, p.tags, p.version, p.created_at, u.username,
	          COUNT(c.id) comments_count
			  FROM posts p LEFT JOIN comments c ON c.post_id = p.id
			  LEFT JOIN users u ON p.user_id = u.id
			  JOIN followers f ON f.follower_id = p.user_id OR p.user_id = $1
			  WHERE f.user_id = $1 AND
					(p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%')
			  GROUP BY p.id, u.username ORDER BY p.created_at ` + fq.Sort + `
			  LIMIT $2 OFFSET $3`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, userID, fq.Limit, fq.Offset, fq.Search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var postsWithMetaData []PostWithMetadata

	for rows.Next() {
		var postMeta PostWithMetadata
		err := rows.Scan(
			&postMeta.Post.ID,
			&postMeta.Post.Title,
			&postMeta.Post.Content,
			&postMeta.Post.UserID,
			pq.Array(&postMeta.Post.Tags),
			&postMeta.Post.Version,
			&postMeta.Post.CreatedAt,
			&postMeta.Post.User.Username,
			&postMeta.CommentCount,
		)
		if err != nil {
			return nil, err
		}

		postsWithMetaData = append(postsWithMetaData, postMeta)
	}

	return postsWithMetaData, nil
}
