package pg

import (
	"context"
	"fmt"
	"log/slog"
	"test_18/internal/config"
	"test_18/internal/domain"
	"test_18/pkg/e"

	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	logger *slog.Logger
	pool   *pgxpool.Pool
}

func NewPostgres(ctx context.Context, cfg config.Config, logger *slog.Logger) (*Postgres, error) {
	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Database, cfg.Postgres.SSLMode,
	)
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, e.Wrap("storage.pg.NewPostgres.ParseConfig", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, e.Wrap("storage.pg.NewPostgres.NewWithConfig", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, e.Wrap("storage.pg.NewPostgres.Ping", err)
	}

	return &Postgres{
		logger: logger,
		pool:   pool,
	}, nil
}

func (p *Postgres) CreateUser(ctx context.Context) (int, error) {
	var id int
	err := p.pool.QueryRow(ctx, `INSERT INTO users (created_at) VALUES (NOW()) RETURNING id`).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (p *Postgres) CreateEvent(ctx context.Context, event domain.Event) (int, error) {
	var id int
	query := `INSERT INTO events (user_id, title, description, event_date) 
              VALUES ($1, $2, $3, $4) RETURNING id`
	err := p.pool.QueryRow(ctx, query,
		event.UserID,
		event.Title,
		event.Description,
		time.Time(event.EventDate).In(time.UTC),
	).Scan(&id)
	if err != nil {
		return 0, e.Wrap("storage.pg.CreateEvent", err)
	}
	return id, nil
}

type EventWithUserTimestamps struct {
	domain.Event
	UserCreatedAt time.Time
	UserUpdatedAt *time.Time
}

func (p *Postgres) UpdateEvent(ctx context.Context, id int, req domain.UpdateEventRequest) error {
	query := `UPDATE events e
SET 
    title = COALESCE($2, e.title),
    description = COALESCE($3, e.description),
    event_date = COALESCE($4, e.event_date),
    updated_at = NOW()
FROM users u
WHERE e.id = $1 AND e.user_id = u.id
RETURNING 
    e.id, e.title, e.description, e.event_date, e.created_at, e.updated_at,
    u.created_at AS user_created_at,
    u.updated_at AS user_updated_at`

	var event EventWithUserTimestamps

	err := p.pool.QueryRow(ctx, query,
		id,
		req.Title,
		req.Description,
		startTimeToTimeUTC(req.EventDate),
	).Scan(
		&event.ID,
		&event.Title,
		&event.Description,
		&event.EventDate,
		&event.CreatedAt,
		&event.UpdatedAt,
		&event.UserCreatedAt,
		&event.UserUpdatedAt,
	)
	if err != nil {
		return e.Wrap("storage.pg.UpdateEvent", err)
	}
	return nil
}

func startTimeToTimeUTC(date *domain.Date) *time.Time {
	if date == nil {
		return nil
	}
	t := time.Time(*date).In(time.UTC)
	return &t
}

func (p *Postgres) DeleteEvent(ctx context.Context, id int) error {
	query := `DELETE FROM events WHERE id = $1`
	_, err := p.pool.Exec(ctx, query, id)
	if err != nil {
		return e.Wrap("storage.pg.DeleteEvent", err)
	}
	return nil
}

func (p *Postgres) GetEventsForDay(ctx context.Context, userId int, day time.Time) (domain.User, error) {
	var user domain.User
	user.Events = make([]domain.Event, 0)

	start := day.Truncate(24 * time.Hour).UTC()
	end := start.Add(24 * time.Hour)

	query := `
SELECT 
    u.id, u.created_at, u.updated_at,
    e.id, e.user_id, e.title, e.description, e.event_date, e.created_at, e.updated_at
FROM users u
LEFT JOIN events e ON e.user_id = u.id AND e.event_date >= $2 AND e.event_date < $3
WHERE u.id = $1
ORDER BY e.event_date ASC
`

	rows, err := p.pool.Query(ctx, query, userId, start, end)
	if err != nil {
		return user, e.Wrap("storage.pg.GetUserWithEventsForDay.Query", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			event                          domain.Event
			eventID                        *int
			userIDFromRow                  int
			userCreatedAt, userUpdatedAt   domain.Date
			eventCreatedAt, eventUpdatedAt domain.Date
		)

		err := rows.Scan(
			&user.ID,
			&userCreatedAt,
			&userUpdatedAt,
			&eventID, // изменили на *int
			&userIDFromRow,
			&event.Title,
			&event.Description,
			&event.EventDate,
			&eventCreatedAt,
			&eventUpdatedAt,
		)
		if err != nil {
			return user, e.Wrap("storage.pg.GetUserWithEventsForDay.Scan", err)
		}

		user.CreatedAt = userCreatedAt
		user.UpdatedAt = userUpdatedAt

		if eventID != nil {
			event.ID = *eventID
			event.UserID = userIDFromRow
			event.CreatedAt = eventCreatedAt
			event.UpdatedAt = eventUpdatedAt
			user.Events = append(user.Events, event)
		}
	}

	if err := rows.Err(); err != nil {
		return user, e.Wrap("storage.pg.GetUserWithEventsForDay.Rows", err)
	}

	return user, nil
}

func (p *Postgres) GetEventsForWeek(ctx context.Context, userId int, date time.Time) (domain.User, error) {
	var user domain.User
	user.Events = make([]domain.Event, 0)

	start := date.Truncate(24*time.Hour).AddDate(0, 0, -7).UTC()
	end := date.Truncate(24*time.Hour).AddDate(0, 0, 1).UTC()

	query := `
SELECT 
    u.id, u.created_at, u.updated_at,
    e.id, e.user_id, e.title, e.description, e.event_date, e.created_at, e.updated_at
FROM users u
LEFT JOIN events e ON e.user_id = u.id AND e.event_date >= $2 AND e.event_date < $3
WHERE u.id = $1
ORDER BY e.event_date ASC
`

	rows, err := p.pool.Query(ctx, query, userId, start, end)
	if err != nil {
		return user, e.Wrap("storage.pg.GetUserWithEventsForWeek.Query", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			event                          domain.Event
			eventID                        *int
			userIDFromRow                  int
			userCreatedAt, userUpdatedAt   domain.Date
			eventCreatedAt, eventUpdatedAt domain.Date
		)

		err := rows.Scan(
			&user.ID,
			&userCreatedAt,
			&userUpdatedAt,
			&eventID,
			&userIDFromRow,
			&event.Title,
			&event.Description,
			&event.EventDate,
			&eventCreatedAt,
			&eventUpdatedAt,
		)
		if err != nil {
			return user, e.Wrap("storage.pg.GetUserWithEventsForWeek.Scan", err)
		}

		user.CreatedAt = userCreatedAt
		user.UpdatedAt = userUpdatedAt

		if eventID != nil {
			event.ID = *eventID
			event.UserID = userIDFromRow
			event.CreatedAt = eventCreatedAt
			event.UpdatedAt = eventUpdatedAt
			user.Events = append(user.Events, event)
		}
	}

	if err := rows.Err(); err != nil {
		return user, e.Wrap("storage.pg.GetUserWithEventsForWeek.Rows", err)
	}

	return user, nil
}

func (p *Postgres) GetEventsForMonth(ctx context.Context, userId int, date time.Time) (domain.User, error) {
	var user domain.User
	user.Events = make([]domain.Event, 0)

	start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	query := `
SELECT 
    u.id, u.created_at, u.updated_at,
    e.id, e.user_id, e.title, e.description, e.event_date, e.created_at, e.updated_at
FROM users u
LEFT JOIN events e ON e.user_id = u.id AND e.event_date >= $2 AND e.event_date < $3
WHERE u.id = $1
ORDER BY e.event_date ASC
`

	rows, err := p.pool.Query(ctx, query, userId, start, end)
	if err != nil {
		return user, e.Wrap("storage.pg.GetUserWithEventsForMonth.Query", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			event                          domain.Event
			userIDFromRow                  int
			eventID                        *int
			userCreatedAt, userUpdatedAt   domain.Date
			eventCreatedAt, eventUpdatedAt domain.Date
		)

		err := rows.Scan(
			&user.ID,
			&userCreatedAt,
			&userUpdatedAt,
			&eventID,
			&userIDFromRow,
			&event.Title,
			&event.Description,
			&event.EventDate,
			&eventCreatedAt,
			&eventUpdatedAt,
		)
		if err != nil {
			return user, e.Wrap("storage.pg.GetUserWithEventsForMonth.Scan", err)
		}

		user.CreatedAt = userCreatedAt
		user.UpdatedAt = userUpdatedAt

		if eventID != nil {
			event.ID = *eventID
			event.UserID = userIDFromRow
			event.CreatedAt = eventCreatedAt
			event.UpdatedAt = eventUpdatedAt
			user.Events = append(user.Events, event)
		}
	}

	if err := rows.Err(); err != nil {
		return user, e.Wrap("storage.pg.GetUserWithEventsForMonth.Rows", err)
	}

	return user, nil
}

func (p *Postgres) CloseDB() {
	p.pool.Close()
	stat := p.pool.Stat()
	if stat.AcquiredConns() > 0 {
		p.logger.Warn("postgres connections not fully closed after Close()", slog.Any("acquired connections", stat.AcquiredConns()))
	}
}
