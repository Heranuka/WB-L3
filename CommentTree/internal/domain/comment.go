package domain

import (
	"time"
)

type Comment struct {
	ID        int       `json:"id"` // Пример: для API
	Content   string    `json:"content"`
	ParentID  *int      `json:"parent_id,omitempty"` // Указатель, чтобы можно было отличить 0 ID от отсутствия родителя
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"` // omitempty - поле не будет сериализовано, если оно zero value
}

type CommentNode struct {
	Comment  Comment       // Сам комментарий
	Children []CommentNode // Слайс его прямых потомков (рекурсивно)
}
