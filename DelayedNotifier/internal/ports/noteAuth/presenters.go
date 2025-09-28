package noteAuth

import (
	"delay/internal/domain"
	"time"
)

type RequestNote struct {
	Message     string                     `json:"message"`
	Destination string                     `json:"destination"`
	Channel     domain.NotificationChannel `json:"channel"`
	DataToSent  time.Time                  `json:"data_sent_at"`
}
