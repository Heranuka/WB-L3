package domain

import "time"

type DayStats struct {
	Date  time.Time `db:"date" json:"date"`
	Count int64     `db:"count" json:"count"`
}

type MonthStats struct {
	Year  int   `db:"year" json:"year"`
	Month int   `db:"month" json:"month"`
	Count int64 `db:"count" json:"count"`
}

type UserAgentStats struct {
	UserAgent string `db:"user_agent" json:"user_agent"`
	Count     int64  `db:"count" json:"count"`
}

type ShortURL struct {
	ID          int       `db:"id" json:"id"`
	ShortCode   string    `db:"short_code" json:"short_code"`
	OriginalURL string    `db:"original_url" json:"original_url"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	Custom      bool      `db:"custom" json:"custom"`
}

type Click struct {
	ID         int       `db:"id" json:"id"`
	ShortURLID int       `db:"short_url_id" json:"short_url_id"`
	Timestamp  time.Time `db:"timestamp" json:"timestamp"`
	UserAgent  string    `db:"user_agent" json:"user_agent"`
	IPAddress  string    `db:"ip_address" json:"ip_address"`
}
