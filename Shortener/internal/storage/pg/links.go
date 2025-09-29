package pg

import (
	"context"
	"database/sql"
	"errors"
	"shortener/internal/domain"
	"shortener/pkg/e"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver
)

type ShortLinkRepository interface {
	Create(ctx context.Context, link *domain.ShortURL) error
	Get(ctx context.Context, shortURL string) (domain.ShortURL, error)
}

type ClickRepository interface {
	Save(ctx context.Context, click *domain.Click) error
	AggregateByDay(ctx context.Context, shortURL string) ([]domain.DayStats, error)
	AggregateByMonth(ctx context.Context, shortURL string) ([]domain.MonthStats, error)
	AggregateByUserAgent(ctx context.Context, shortURL string) ([]domain.UserAgentStats, error)
}

func (p *Postgres) AggregateByDay(ctx context.Context, shortURL string) ([]domain.DayStats, error) {
	query := `
        SELECT DATE(timestamp) AS date, COUNT(*) AS count
        FROM clicks
        WHERE short_url_id = (SELECT id FROM short_urls WHERE short_code = $1)
        GROUP BY DATE(timestamp)
        ORDER BY date DESC`
	var stats []domain.DayStats
	rows, err := p.db.Master.QueryContext(ctx, query, shortURL)
	if err != nil {
		return nil, e.Wrap("storage/AggregateByDay/QueryContext", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stat domain.DayStats
		if err := rows.Scan(&stat.Date, &stat.Count); err != nil {
			return nil, e.Wrap("storage/AggregateByDay/rows.Scan", err)
		}
		stats = append(stats, stat)
	}

	return stats, rows.Err()
}

func (p *Postgres) AggregateByMonth(ctx context.Context, shortURL string) ([]domain.MonthStats, error) {
	query := `
        SELECT EXTRACT(YEAR FROM timestamp) AS year, EXTRACT(MONTH FROM timestamp) AS month, COUNT(*) AS count
        FROM clicks
        WHERE short_url_id = (SELECT id FROM short_urls WHERE short_code = $1)
        GROUP BY year, month
        ORDER BY year DESC, month DESC`
	var stats []domain.MonthStats
	rows, err := p.db.Master.QueryContext(ctx, query, shortURL)
	if err != nil {
		return nil, e.Wrap("storage/AggregateByMonth/QueryContext", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stat domain.MonthStats
		if err := rows.Scan(&stat.Year, &stat.Month, &stat.Count); err != nil {
			return nil, e.Wrap("storage/AggregateByMonth/rows.Scan", err)
		}
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

func (p *Postgres) AggregateByUserAgent(ctx context.Context, shortURL string) ([]domain.UserAgentStats, error) {
	query := `
        SELECT user_agent, COUNT(*) AS count
        FROM clicks
        WHERE short_url_id = (SELECT id FROM short_urls WHERE short_code = $1)
        GROUP BY user_agent
        ORDER BY count DESC`
	var stats []domain.UserAgentStats
	rows, err := p.db.Master.QueryContext(ctx, query, shortURL)
	if err != nil {
		return nil, e.Wrap("storage/AggregateByUserAgent/QueryContext", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stat domain.UserAgentStats
		if err := rows.Scan(&stat.UserAgent, &stat.Count); err != nil {
			return nil, e.Wrap("storage/AggregateByUserAgent/rows.Scan", err)
		}
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

func (p *Postgres) Create(ctx context.Context, link *domain.ShortURL) error {
	query := `
        INSERT INTO short_urls (short_code, original_url, created_at, custom)
        VALUES ($1, $2, $3, $4)
        RETURNING id`
	err := p.db.Master.QueryRowContext(ctx, query,
		link.ShortCode, link.OriginalURL, link.CreatedAt, link.Custom).
		Scan(&link.ID)

	if err != nil {
		return e.Wrap("storage/Create", err)
	}
	return nil
}

func (p *Postgres) Get(ctx context.Context, shortURL string) (domain.ShortURL, error) {
	var link domain.ShortURL
	query := `
        SELECT id, short_code, original_url, created_at, custom
        FROM short_urls
        WHERE short_code = $1`
	row := p.db.Master.QueryRow(query, shortURL)
	err := row.Scan(&link.ID, &link.ShortCode, &link.OriginalURL, &link.CreatedAt, &link.Custom)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return link, domain.ErrLinkNotFound
		}
		return link, e.Wrap("storage/Get", err)
	}
	return link, nil
}

func (p *Postgres) Save(ctx context.Context, click *domain.Click) error {
	query := `
        INSERT INTO clicks (short_url_id, timestamp, user_agent, ip_address)
        VALUES ($1, $2, $3, $4)`
	_, err := p.db.Master.ExecContext(ctx, query, click.ShortURLID, click.Timestamp, click.UserAgent, click.IPAddress)
	if err != nil {
		return e.Wrap("storage/Save", err)
	}
	return nil
}
