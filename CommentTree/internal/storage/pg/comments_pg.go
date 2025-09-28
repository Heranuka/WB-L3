package pg

import (
	"commentTree/internal/domain"
	"context"
	"time"
)

func (s *Postgres) Create(ctx context.Context, comm *domain.Comment) (int, error) {
	query := `INSERT INTO comments (user_id, post_id, content, created_at, updated_at, parent_id)
	 VALUES($1, $2, $3, $4, $5, $6) RETURNING id`
	var id int
	if err := s.pool.QueryRow(ctx, query, comm.ParentID, comm.PostID, comm.Content, time.Now(), nil, comm.ParentID).Scan(&id); err != nil {
		return -1, err
	}

	return id, nil
}

func (s *Postgres) GetById(ctx context.Context, id int) (*domain.Comment, error) {
	var comm domain.Comment

	query := `SELECT FROM comments user_id, post_id, content, created_at, updated_at, parent_id WHERE id = $1`
	err := s.pool.QueryRow(ctx, query, id).Scan(&comm.UserID, &comm.PostID, &comm.Content, &comm.CreatedAt, &comm.UpdatedAt, &comm.ParentID)
	if err != nil {
		return nil, err
	}
	return &comm, nil
}

func (s *Postgres) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM comments WHERE id = $1`
	_, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	return nil

}
