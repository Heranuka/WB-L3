package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"wb-l3.7/internal/domain"
)

func (pg *Postgres) CreateItem(ctx context.Context, item *domain.Item) (int64, error) {
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	var id int64
	query := `
        INSERT INTO items (name, description, price, stock, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
    `
	_ = pg.pool.QueryRow(ctx, query, item.Name, item.Description, item.Price, item.Stock, item.CreatedAt, item.UpdatedAt).Scan(&id)

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
	row := pg.pool.QueryRow(ctx, query, itemID)
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
	rows, err := pg.pool.Query(ctx, query)
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
func (pg *Postgres) UpdateItem(ctx context.Context, item *domain.Item) error {
	item.UpdatedAt = time.Now()
	query := `
        UPDATE items
        SET name=$1, description=$2, price=$3, stock=$4, updated_at=$5
        WHERE id=$6
    `
	res, err := pg.pool.Exec(ctx, query,
		item.Name, item.Description, item.Price, item.Stock, item.UpdatedAt, item.ID,
	)
	if err != nil {
		return err
	}
	rowsAffected := res.RowsAffected()

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteItem удаляет товар по ID
func (pg *Postgres) DeleteItem(ctx context.Context, itemID int64) error {
	query := `DELETE FROM items WHERE id=$1`
	res, err := pg.pool.Exec(ctx, query, itemID)
	if err != nil {
		return err
	}
	rowsAffected := res.RowsAffected()

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
