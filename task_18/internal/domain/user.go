package domain

type User struct {
	ID        int     `json:"id"`
	CreatedAt Date    `json:"created_at"`
	UpdatedAt Date    `json:"updated_at"`
	Events    []Event `json:"events"`
}
