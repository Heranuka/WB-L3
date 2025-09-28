package domain

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID          uuid.UUID           `json:"id"`          // уникальный ID
	Message     string              `json:"message"`     // текст уведомления
	Destination string              `json:"destination"` // куда отправлять (email, telegram username, телефон)
	Channel     NotificationChannel `json:"channel"`     // через какой канал
	Status      NotificationStatus  `json:"status"`      // текущий статус
	DataToSent  time.Time           `json:"data_sent_at"`
	CreatedAt   time.Time           `json:"created_at"`
}

/*
	func (n *Notification) UnmarshalJSON(data []byte) error {
		log.Println("UnmarshalJSON called") // <--- ДОБА
		// 1. Временная структура
		var aux struct {
			Message     string              `json:"message"`
			DataToSent  string              `json:"data_to_sent"`
			Channel     NotificationChannel `json:"channel"`
			Destination string              `json:"destination"`
		}

		// 2. Парсим во временную структуру
		if err := json.Unmarshal(data, &aux); err != nil {
			return err // Ошибка при парсинге JSON
		}

		// 3. Заполняем поля Notification
		n.Message = aux.Message
		n.Channel = aux.Channel         // <--- Добавляем
		n.Destination = aux.Destination // <

		// 4. Парсим DataToSent (если есть)
		if aux.DataToSent != "" {
			duration, err := time.ParseDuration(aux.DataToSent)
			if err != nil {
				return fmt.Errorf("invalid duration format: %w", err)
			}
			n.DataToSent = duration
		} else {
			n.DataToSent = 0 // Установите значение по умолчанию, если DataToSent отсутствует
		}

		return nil
	}
*/
type StatusResponse struct {
	NoteID uuid.UUID `json:"note_id"`
	Status string    `json:"status"`
}
type CancelResponse struct {
	NoteID  uuid.UUID `json:"note_id"`
	Message string    `json:"message"`
}

// RetryPolicy описывает стратегию повторных попыток с экспоненциальной задержкой
type RetryPolicy struct {
	MaxAttempts   int           `json:"max_attempts"`   // максимальное число попыток
	InitialDelay  time.Duration `json:"initial_delay"`  // первая задержка
	MaxDelay      time.Duration `json:"max_delay"`      // максимальная задержка
	BackoffFactor float64       `json:"backoff_factor"` // множитель задержки
}

// NotificationChannelSender интерфейс для отправки уведомлений через разные каналы
type NotificationChannelSender interface {
	Send(message, destination string) error
}
