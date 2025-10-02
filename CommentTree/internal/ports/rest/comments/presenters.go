package comments

import "commentTree/internal/domain"

type commentCreate struct {
	Content  string `json:"content"`
	ParentID *int   `json:"parent_id"`
}

type CommentResponse struct {
	domain.Comment
	Children []*CommentResponse `json:"children,omitempty"`
}
