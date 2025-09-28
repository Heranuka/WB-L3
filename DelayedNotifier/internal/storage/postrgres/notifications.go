package postgres

import (
	"context"
	"delay/internal/domain"
	"fmt"

	"github.com/google/uuid"
)

type NotificationStorage interface {
	Create(ctx context.Context, notification *domain.Notification) (uuid.UUID, error)
	Status(ctx context.Context, noteID uuid.UUID) (string, error)
	Cancel(ctx context.Context, noteID uuid.UUID) error
	GetAll(ctx context.Context) (*[]domain.Notification, error)
	Get(ctx context.Context, noteID uuid.UUID) (*domain.Notification, error)
}

func (p *Postgres) Create(ctx context.Context, notification *domain.Notification) (uuid.UUID, error) {
	query := `INSERT INTO notifications(id, message, destination, channel, data_sent_at, status) VALUES($1, $2, $3, $4, $5, $6)`
	id := uuid.New()
	status := "created"
	_, err := p.db.Master.ExecContext(ctx, query, id, notification.Message, notification.Destination, notification.Channel, notification.DataToSent, status)
	if err != nil {
		return uuid.Nil, fmt.Errorf("storage.postgres.Create: %w", err)
	}
	return id, nil
}

func (p *Postgres) Status(ctx context.Context, noteID uuid.UUID) (string, error) {
	query := `SELECT status FROM notifications WHERE id = $1`

	var status string

	err := p.db.Master.QueryRowContext(ctx, query, noteID).Scan(&status)
	if err != nil {
		return "", fmt.Errorf("storage.postgres.Status: %w", err)
	}

	return status, nil
}
func (p *Postgres) Cancel(ctx context.Context, noteID uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1`

	_, err := p.db.Master.ExecContext(ctx, query, noteID)
	if err != nil {
		return fmt.Errorf("storage.postgres.Cancel: %w", err)
	}

	return nil
}

func (p *Postgres) GetAll(ctx context.Context) (*[]domain.Notification, error) {
	query := `SELECT id, message, destination, channel, status, data_sent_at, created_at FROM notifications`
	var notifications []domain.Notification

	rows, err := p.db.Master.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("storage.postgres.GetAll: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var notification domain.Notification
		err := rows.Scan(&notification.ID, &notification.Message, &notification.Destination, &notification.Channel, &notification.Status, &notification.DataToSent, &notification.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("storage.postgres.GetAll.Rows.Next: %w", err)
		}
		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage.postgres.GetAll.Rows: %w", err)
	}

	return &notifications, nil
}

func (p *Postgres) UpdateStatus(ctx context.Context, noteID uuid.UUID, status string) error {
	query := `UPDATE notifications SET status=$1 WHERE id = $2`
	_, err := p.db.Master.ExecContext(ctx, query, status, noteID)
	if err != nil {
		return fmt.Errorf("storage.postgres.UpdateStatus: %w", err)
	}

	return nil
}

func (p *Postgres) Get(ctx context.Context, noteID uuid.UUID) (*domain.Notification, error) {
	var note domain.Notification
	query := `SELECT id, message, destination, channel, status, data_sent_at, created_at FROM notifications WHERE id = $1`
	err := p.db.Master.QueryRowContext(ctx, query, noteID).Scan(&note.ID, &note.Message, &note.Destination, &note.Channel, &note.Status, &note.DataToSent, &note.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("storage.postgres.UpdateStatus: %w", err)
	}

	return &note, nil
}
