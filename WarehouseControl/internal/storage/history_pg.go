package storage

import (
	"context"
	"time"

	"wb-l3.7/internal/domain"
)

func (pg *Postgres) LogChange(ctx context.Context, userID, itemID int64, changeDesc string) error {
	now := time.Now()
	query := `
        INSERT INTO item_history (id, item_id, changed_by_user_id, change, changed_at)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := pg.pool.Exec(ctx, query, itemID, itemID, userID, changeDesc, now)
	if err != nil {
		return err
	}

	return nil
}

// GetItemHistory возвращает историю изменений для товара
func (pg *Postgres) GetItemHistory(ctx context.Context, itemID int64) ([]*domain.ItemHistoryEntry, error) {
	query := `
        SELECT id, item_id, changed_by_user_id, change, changed_at
        FROM item_history
        WHERE item_id = $1
        ORDER BY changed_at DESC
    `
	rows, err := pg.pool.Query(ctx, query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*domain.ItemHistoryEntry
	for rows.Next() {
		entry := &domain.ItemHistoryEntry{}
		if err := rows.Scan(&entry.ID, &entry.ItemID, &entry.ChangedByUserID, &entry.Change, &entry.ChangedAt); err != nil {
			return nil, err
		}
		history = append(history, entry)
	}
	return history, nil
}
