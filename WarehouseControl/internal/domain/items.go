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

/*
type ItemHistoryEntry struct {
	ID              int64     `json:"id"`
	ItemID          int64     `json:"item_id"`
	ChangedByUserID string    `json:"changed_by_user_id"`
	Change          string    `json:"change"` // Описание изменения, например: "price updated from 100.00 to 120.00"
	ChangedAt       time.Time `json:"changed_at"`
	Version         int       `json:"version"`
} */

type ChangeDiff struct {
	Old interface{} `json:"old,omitempty"` // старое значение
	New interface{} `json:"new,omitempty"` // новое значение
}

// ItemHistoryRecord представляет одну запись истории изменений товара
type ItemHistoryRecord struct {
	ID                int                   `json:"id"`
	ItemID            int                   `json:"item_id"`
	ChangedByUser     string                `json:"changed_by_user"`
	ChangeDescription string                `json:"change_description"`
	ChangedAt         time.Time             `json:"changed_at"`
	Version           int                   `json:"version"`
	ChangeDiff        map[string]ChangeDiff `json:"change_diff,omitempty"` // ключ - название поля
}
