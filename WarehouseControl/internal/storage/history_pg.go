package storage

import (
	"context"
	"database/sql"
	"encoding/json"
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
	query := `
        SELECT id, item_id, changed_by_user_id, change_description, changed_at, version, change_diff
        FROM item_history
        WHERE item_id = $1
        ORDER BY changed_at DESC
    `
	rows, err := pg.db.Master.QueryContext(ctx, query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*domain.ItemHistoryRecord
	for rows.Next() {
		entry := &domain.ItemHistoryRecord{}
		var changedByUserID string
		var changeDiffJSON []byte

		// Сканируем все поля, включая JSON поля в []byte
		err := rows.Scan(&entry.ID, &entry.ItemID, &changedByUserID, &entry.ChangeDescription, &entry.ChangedAt, &entry.Version, &changeDiffJSON)
		if err != nil {
			return nil, err
		}

		// Распарсим JSON change_diff в map[string]ChangeDiff
		err = json.Unmarshal(changeDiffJSON, &entry.ChangeDiff)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		// Преобразование или поиск пользователя по changedByUserID можно сделать здесь,
		// пока просто сохраняем как строку в ChangedByUser для фронтенда
		entry.ChangedByUser = changedByUserID

		history = append(history, entry)
	}
	return history, nil
}
