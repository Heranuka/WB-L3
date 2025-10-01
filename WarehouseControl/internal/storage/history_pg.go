package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"wb-l3.7/internal/domain"
)

type History interface {
	LogChange(ctx context.Context, userID, itemID int64, changeDesc string, changeDiff map[string]domain.ChangeDiff) error
	GetItemHistory(ctx context.Context, itemID int64) ([]*domain.ItemHistoryRecord, error)
}

func (pg *Postgres) LogChange(ctx context.Context, userID, itemID int64, changeDesc string, changeDiff map[string]domain.ChangeDiff) error {
	now := time.Now()

	// Сериализуем changeDiff в JSON
	diffJSON, err := json.Marshal(changeDiff)
	if err != nil {
		return err
	}

	query := `
        INSERT INTO item_history (item_id, changed_by_user_id, change_description, changed_at, change_diff)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err = pg.db.Master.ExecContext(ctx, query, itemID, userID, changeDesc, now, diffJSON)
	if err != nil {
		return err
	}
	return nil
}

func (pg *Postgres) GetItemHistory(ctx context.Context, itemID int64) ([]*domain.ItemHistoryRecord, error) {
	const query = `
        SELECT id, item_id, changed_by_user_id, change_description, changed_at, version, change_diff
        FROM item_history
        WHERE item_id = $1
        ORDER BY version DESC
    `
	rows, err := pg.db.QueryContext(ctx, query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*domain.ItemHistoryRecord
	for rows.Next() {
		var entry domain.ItemHistoryRecord
		var changeDiffJSON []byte

		err = rows.Scan(&entry.ID, &entry.ItemID, &entry.ChangedByUser, &entry.ChangeDescription, &entry.ChangedAt, &entry.Version, &changeDiffJSON)
		if err != nil {
			return nil, err
		}

		// Для отладки: распечатать исходный JSON change_diff
		var raw map[string]interface{}
		if err := json.Unmarshal(changeDiffJSON, &raw); err != nil {
			fmt.Printf("Failed to unmarshal change_diff raw JSON: %v\n", err)
		} else {
			fmt.Printf("Raw change_diff: %+v\n", raw)
		}

		// Здесь можно попробовать распарсить в map[string]ChangeDiff
		var changeDiff map[string]domain.ChangeDiff
		err = json.Unmarshal(changeDiffJSON, &changeDiff)
		if err != nil {
			fmt.Printf("Error unmarshaling to ChangeDiff map: %v\n", err)
		}
		entry.ChangeDiff = changeDiff

		history = append(history, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return history, nil
}
