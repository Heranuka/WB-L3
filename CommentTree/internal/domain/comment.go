package domain

import (
	"time"
)

type Comment struct {
	ID        int       `json:"id"` // Пример: для API
	UserID    int       `json:"user_id"`
	PostID    int       `json:"post_id"` // Или ItemID, ArticleID и т.д.
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"` // omitempty - поле не будет сериализовано, если оно zero value
	ParentID  *int      `json:"parent_id,omitempty"`  // Указатель, чтобы можно было отличить 0 ID от отсутствия родителя
	// IsDeleted bool      `json:"-"` // Может быть скрыто от API
}

type NewComment struct {
	AuthorID int    `json:"authorID"`           // ID автора
	PostID   int    `json:"postID"`             // ID поста/статьи/элемента, к которому относится коммент
	ParentID *int   `json:"parentID,omitempty"` // ID родительского комментария (nullable)
	Text     string `json:"text"`               // Текст комментария
}

type CommentNode struct {
	Comment  Comment       // Сам комментарий
	Children []CommentNode // Слайс его прямых потомков (рекурсивно)
}
