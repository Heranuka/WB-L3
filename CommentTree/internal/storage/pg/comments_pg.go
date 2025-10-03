package pg

import (
	"commentTree/internal/domain"
	"commentTree/pkg/e"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (s *Postgres) Create(ctx context.Context, comm *domain.Comment) (int, error) {
	query := `INSERT INTO comments (content, parent_id)
	 VALUES($1, $2) RETURNING id`
	var id int
	if err := s.db.Master.QueryRowContext(ctx, query, comm.Content, comm.ParentID).Scan(&id); err != nil {
		return -1, fmt.Errorf("storage/pg/comments_pg/Create: %w", err)
	}

	return id, nil
}

func (s *Postgres) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM comments WHERE id = $1`
	row, err := s.db.Master.ExecContext(ctx, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return e.ErrNotFound
		}
		return fmt.Errorf("storage/pg/comments_pg/Delete: %w", err)
	}

	affectedRows, err := row.RowsAffected()
	if err != nil {
		return fmt.Errorf("storage/pg/comments_pg/Delete/RowsAffected: %w", err)
	}

	if affectedRows == 0 {
		return e.ErrNotFound
	}
	return nil
}

func (s *Postgres) GetRootComments(ctx context.Context, search *string, limit, offset int) ([]*domain.Comment, error) {
	query := `
        SELECT id, content, parent_id, created_at, updated_at
        FROM comments
        WHERE parent_id IS NULL
        AND ($1::text IS NULL OR to_tsvector('russian', content) @@ plainto_tsquery('russian', $1::text))
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `
	var searchVal interface{}
	if search == nil {
		searchVal = nil
	} else {
		searchVal = *search
	}

	rows, err := s.db.Master.QueryContext(ctx, query, searchVal, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("storage/pg/comments_pg/GetRootsComments: %w", err)
	}
	defer rows.Close()

	var comments []*domain.Comment
	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(&c.ID, &c.Content, &c.ParentID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("storage/pg/comments_pg/GetRootsComments/rows.Scan: %w", err)
		}
		comments = append(comments, &c)
	}

	return comments, nil
}

func (s *Postgres) GetChildComments(ctx context.Context, parentID int) ([]*domain.Comment, error) {
	query := `
        SELECT id, content, parent_id, created_at, updated_at
        FROM comments
        WHERE parent_id = $1
        ORDER BY created_at ASC
    `
	rows, err := s.db.Master.QueryContext(ctx, query, parentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, e.ErrNotFound
		}
		return nil, fmt.Errorf("storage/pg/comments_pg/GetChildComments: %w", err)
	}
	defer rows.Close()

	var comments []*domain.Comment
	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(&c.ID, &c.Content, &c.ParentID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("storage/pg/comments_pg/GetChildComments/rows.Scan: %w", err)
		}
		comments = append(comments, &c)
	}
	return comments, nil
}

func (s *Postgres) GetAllComments(ctx context.Context) ([]domain.CommentNode, error) {
	query := `
        SELECT id, content, parent_id, created_at, updated_at
        FROM comments
        ORDER BY created_at DESC
    `

	rows, err := s.db.Master.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []domain.Comment
	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(&c.ID, &c.Content, &c.ParentID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	// Build tree
	nodeMap := make(map[int]*domain.CommentNode)
	var roots []domain.CommentNode

	for _, comment := range comments {
		node := domain.CommentNode{
			Comment:  comment,
			Children: []domain.CommentNode{},
		}
		nodeMap[comment.ID] = &node
	}

	for _, node := range nodeMap {
		if node.Comment.ParentID == nil {
			roots = append(roots, *node)
		} else {
			parentNode, ok := nodeMap[*node.Comment.ParentID]
			if ok {
				parentNode.Children = append(parentNode.Children, *node)
			} else {
				roots = append(roots, *node) // parent missing, treat as root
			}
		}
	}

	return roots, nil
}
