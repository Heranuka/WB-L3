package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"wb-l3.7/internal/domain"
)

type Items interface {
	CreateItem(ctx context.Context, item *domain.Item, userID int64) (int64, error)
	GetItem(ctx context.Context, itemID int64) (*domain.Item, error)
	GetAllItems(ctx context.Context) ([]*domain.Item, error)
	UpdateItem(ctx context.Context, item *domain.Item, userID int64) error
	DeleteItem(ctx context.Context, itemID int64) error
}

func (pg *Postgres) CreateItem(ctx context.Context, item *domain.Item, userID int64) (int64, error) {
	tx, err := pg.db.Master.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Устанавливаем параметр сессии для триггера
	// Формируем команду SET LOCAL с подставленным userID
	setUserIDQuery := fmt.Sprintf("SET LOCAL myapp.current_user_id = %d", userID)
	_, err = tx.ExecContext(ctx, setUserIDQuery)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	var id int64
	query := `
        INSERT INTO items (name, description, price, stock)
        VALUES ($1, $2, $3, $4) RETURNING id
    `
	err = tx.QueryRowContext(ctx, query, item.Name, item.Description, item.Price, item.Stock).Scan(&id)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetItem возвращает товар по ID
func (pg *Postgres) GetItem(ctx context.Context, itemID int64) (*domain.Item, error) {
	query := `
        SELECT id, name, description, price, stock, created_at, updated_at
        FROM items
        WHERE id = $1
    `
	item := &domain.Item{}
	row := pg.db.Master.QueryRowContext(ctx, query, itemID)
	err := row.Scan(&item.ID, &item.Name, &item.Description, &item.Price, &item.Stock, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

// GetAllItems возвращает все товары
func (pg *Postgres) GetAllItems(ctx context.Context) ([]*domain.Item, error) {
	query := `
        SELECT id, name, description, price, stock, created_at, updated_at
        FROM items
        ORDER BY name
    `
	rows, err := pg.db.Master.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.Item
	for rows.Next() {
		item := &domain.Item{}
		err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price, &item.Stock, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// UpdateItem обновляет существующий товар
func (pg *Postgres) UpdateItem(ctx context.Context, item *domain.Item, userID int64) error {
	tx, err := pg.db.Master.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	item.UpdatedAt = time.Now()
	query := `
        UPDATE items
        SET name=$1, description=$2, price=$3, stock=$4, updated_at=$5
        WHERE id=$6
    `

	item.UpdatedAt = time.Now()
	setUserIDQuery := fmt.Sprintf("SET LOCAL myapp.current_user_id = %d", userID)
	_, err = tx.ExecContext(ctx, setUserIDQuery)
	if err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, query,
		item.Name, item.Description, item.Price, item.Stock, item.UpdatedAt, item.ID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to number of affected rows : %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// DeleteItem удаляет товар по ID
func (pg *Postgres) DeleteItem(ctx context.Context, itemID int64) error {
	tx, err := pg.db.Master.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := `DELETE FROM items WHERE id=$1`

	res, err := tx.ExecContext(ctx, query, itemID)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

/*
func (pg *Postgres) GetItemHistory(ctx context.Context, itemID int64) ([]domain.ItemHistoryEntry, error) {
	const query = `
        SELECT id, item_id, changed_by_user_id, change, changed_at, version
        FROM item_history
        WHERE item_id = $1
        ORDER BY version DESC
    `

	rows, err := pg.db.QueryContext(ctx, query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []domain.ItemHistoryEntry
	for rows.Next() {
		var entry domain.ItemHistoryEntry
		if err := rows.Scan(&entry.ID, &entry.ItemID, &entry.ChangedByUserID, &entry.Change, &entry.ChangedAt, &entry.Version); err != nil {
			return nil, err
		}
		history = append(history, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}
*/
