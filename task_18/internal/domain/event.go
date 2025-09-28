package domain

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type Event struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`     // Уникальный идентификатор события (UUID)
	Title       string `json:"title"`       // Название события
	Description string `json:"description"` // Описание события
	EventDate   Date   `json:"event_date"`
	CreatedAt   Date   `json:"created_at"`
	UpdatedAt   Date   `json:"updated_at"`
}

type UpdateEventRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	EventDate   *Date   `json:"event_date"`
}

type Request struct {
	UserID      int    `json:"user_id"`
	Title       string `json:"title"`       // Название события
	Description string `json:"description"` // Описание события
	EventDate   Date   `json:"event_date"`
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
type Date time.Time

func (d Date) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	if t.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", t.Format(time.RFC3339))), nil
}
func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	fmt.Println("UnmarshalJSON input:", s) // лог

	if s == "" || s == "null" {
		*d = Date(time.Time{})
		return nil
	}

	if t, err := time.Parse("2006-01-02", s); err == nil {
		*d = Date(t.UTC())
		return nil
	}

	if t, err := time.Parse(time.RFC3339, s); err == nil {
		*d = Date(t.UTC())
		return nil
	} else {
		fmt.Println("Failed RFC3339 parse:", err)
	}

	return fmt.Errorf("unsupported date format: %s", s)
}

func (d *Date) Scan(value interface{}) error {
	if value == nil {
		*d = Date(time.Time{})
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		*d = Date(v)
		return nil
	case []byte:
		str := string(v)
		fmt.Println("Scan input (bytes):", str) // лог
		if t, err := time.Parse("2006-01-02", str); err == nil {
			*d = Date(t)
			return nil
		}
		t, err := time.Parse(time.RFC3339, str)
		if err != nil {
			fmt.Println("Scan parse error:", err)
			return err
		}
		*d = Date(t)
		return nil
	case string:
		fmt.Println("Scan input (string):", v) // лог
		if t, err := time.Parse("2006-01-02", v); err == nil {
			*d = Date(t)
			return nil
		}
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			fmt.Println("Scan parse error:", err)
			return err
		}
		*d = Date(t)
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into domain.Date", value)
	}
}

// Value реализует driver.Valuer для правильной записи в БД (если нужно)
func (d Date) Value() (driver.Value, error) {
	t := time.Time(d)
	if t.IsZero() {
		return nil, nil
	}
	return t, nil
}
