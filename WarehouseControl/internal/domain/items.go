package domain

import "time"

type Item struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ItemHistoryEntry представляет одну запись в истории изменений
type ItemHistoryEntry struct {
	ID              int64     `json:"id"`
	ItemID          int64     `json:"item_id"`
	ChangedByUserID string    `json:"changed_by_user_id"`
	Change          string    `json:"change"` // Описание изменения, например: "price updated from 100.00 to 120.00"
	ChangedAt       time.Time `json:"changed_at"`
	// Optional: Version int `json:"version"`
	// Optional: UserNickname string `json:"user_nickname"` // Для удобства, если нужно
}
