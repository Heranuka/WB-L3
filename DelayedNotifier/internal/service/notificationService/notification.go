package notificationService

import "context"

type NotificationChannel interface {
	Send(ctx context.Context, message, destination string) error
}
