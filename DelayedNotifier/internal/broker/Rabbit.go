package broker

import (
	"context"
	"delay/internal/config"
	"delay/internal/domain"
	"delay/internal/service/notificationService"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/wb-go/wbf/rabbitmq"
)

//go:generate mockgen -source=rabbit.go -destination=mocks/mock.go
type Rabbit interface {
	UpdateStatus(ctx context.Context, noteID uuid.UUID, status string) error
	Consume(ctx context.Context, emailChannel, telegramChannel notificationService.NotificationChannel) error
	Publish(note *domain.Notification) error
}

type RabbitMQ struct {
	logger              zerolog.Logger
	wg                  *sync.WaitGroup
	notificationService Rabbit
	conn                *rabbitmq.Connection
	ch                  *rabbitmq.Channel
	queueName           string
	dlxName             string
	dlqName             string
	notifyCloseChan     chan *amqp.Error
	notifyCloseChanLock sync.Mutex
	cfg                 *config.Config
}

func NewRabbitMQ(ctx context.Context, logger zerolog.Logger, cfg *config.Config, queueName, dlxName, dlqName string, notificationService Rabbit) (*RabbitMQ, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
	)

	conn, err := rabbitmq.Connect(url, 10, 5*time.Second)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to RabbitMQ")
		return nil, err
	}
	logger.Info().Msg("Successfully connected to RabbitMQ")
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		logger.Error().Err(err).Msg("Failed to open channel")
		return nil, err
	}

	queueManager := rabbitmq.NewQueueManager(ch)

	// Объявляем DLX и DLQ
	_, err = queueManager.DeclareQueue(dlxName, rabbitmq.QueueConfig{Durable: true})
	if err != nil {
		return nil, fmt.Errorf("failed to declare DLX: %w", err)
	}

	_, err = queueManager.DeclareQueue(dlqName, rabbitmq.QueueConfig{Durable: true})
	if err != nil {
		return nil, fmt.Errorf("failed to declare DLQ: %w", err)
	}

	args := rabbitmq.QueueConfig{
		Durable: true,
		Args: amqp.Table{
			"x-dead-letter-exchange":    dlxName,
			"x-dead-letter-routing-key": dlqName,
		},
	}

	_, err = queueManager.DeclareQueue(queueName, args)
	if err != nil {
		return nil, fmt.Errorf("failed to declare main queue: %w", err)
	}

	r := &RabbitMQ{
		conn:                conn,
		ch:                  ch,
		queueName:           queueName,
		dlxName:             dlxName,
		dlqName:             dlqName,
		wg:                  &sync.WaitGroup{},
		logger:              logger,
		notificationService: notificationService,
		notifyCloseChan:     make(chan *amqp.Error),
		cfg:                 cfg,
	}

	r.conn.NotifyClose(r.notifyCloseChan)
	r.handleReconnect(ctx)

	return r, nil
}

func (r *RabbitMQ) handleReconnect(ctx context.Context) {
	go func() {
		for {
			select {
			case <-r.notifyCloseChan:
				r.logger.Error().Msg("Connection closed, trying to reconnect...")
				for {
					if ctx.Err() != nil {
						r.logger.Info().Msg("Context cancelled, stopping reconnect attempts")
						return
					}
					conn, err := rabbitmq.Connect(fmt.Sprintf("amqp://%s:%s@%s:%d/", r.cfg.RabbitMQ.User, r.cfg.RabbitMQ.Password, r.cfg.RabbitMQ.Host, r.cfg.RabbitMQ.Port), 10, 5*time.Second)
					if err == nil {
						ch, err := conn.Channel()
						if err == nil {
							// Закрыть старый канал notifyCloseChan, чтобы избежать утечек
							r.notifyCloseChanLock.Lock()
							close(r.notifyCloseChan)
							r.notifyCloseChanLock.Unlock()

							r.conn = conn
							r.ch = ch

							r.notifyCloseChanLock.Lock()
							r.notifyCloseChan = make(chan *amqp.Error)
							r.conn.NotifyClose(r.notifyCloseChan)
							r.notifyCloseChanLock.Unlock()

							if err := r.setupQueues(); err != nil {
								r.logger.Error().Err(err).Msg("Error setting up queues during reconnect")
								conn.Close()
							} else {
								r.logger.Info().Msg("RabbitMQ reconnected successfully")
								break
							}
						} else {
							conn.Close()
						}
					}
					r.logger.Error().Err(err).Msg("Reconnect failed, retrying...")
					time.Sleep(5 * time.Second)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (r *RabbitMQ) setupQueues() error {
	err := r.ch.ExchangeDeclare(
		r.dlxName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error declaring DLX: %w", err)
	}

	_, err = r.ch.QueueDeclare(
		r.dlqName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error declaring DLQ: %w", err)
	}

	err = r.ch.QueueBind(
		r.dlqName,
		r.dlqName,
		r.dlxName,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error binding DLQ: %w", err)
	}

	args := amqp.Table{
		"x-dead-letter-exchange":    r.dlxName,
		"x-dead-letter-routing-key": r.dlqName,
	}
	_, err = r.ch.QueueDeclare(
		r.queueName,
		true,
		false,
		false,
		false,
		args,
	)
	if err != nil {
		return fmt.Errorf("error declaring main queue: %w", err)
	}
	return nil
}

func (r *RabbitMQ) Publish(note *domain.Notification) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, err := json.Marshal(note)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to marshal notification")
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	err = r.ch.PublishWithContext(ctx,
		"",          // default exchange
		r.queueName, // routing key
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: 2, // persistent
		})
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to publish message")
		return fmt.Errorf("failed to publish message: %w", err)
	}

	r.logger.Info().Msgf("Sent message: %s", body)
	return nil
}

func (r *RabbitMQ) Consume(ctx context.Context, emailChannel, telegramChannel notificationService.NotificationChannel) error {
	msgs, err := r.ch.Consume(
		r.queueName,
		"",
		false, // autoAck disabled
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to register consumer")
		return err
	}

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		for {
			select {
			case <-ctx.Done():
				r.logger.Info().Msg("Consumer cancelled")
				return
			case d, ok := <-msgs:
				if !ok {
					r.logger.Info().Msg("Delivery channel closed")
					return
				}

				var notification domain.Notification
				if err := json.Unmarshal(d.Body, &notification); err != nil {
					r.logger.Error().Err(err).Msg("Failed to unmarshal notification")
					d.Nack(false, false)
					continue
				}

				timeToWait := notification.DataToSent
				if !timeToWait.IsZero() {
					duration := time.Until(timeToWait)
					if duration > 0 {
						timer := time.NewTimer(duration)
						select {
						case <-timer.C:
							r.logger.Info().Str("note_id", notification.ID.String()).Msg("Time to send notification")
						case <-ctx.Done():
							timer.Stop()
							r.logger.Info().Str("note_id", notification.ID.String()).Str("reason", ctx.Err().Error()).Msg("Context cancelled")
							d.Nack(false, true)
							return
						}
					}
				}

				var channel notificationService.NotificationChannel
				switch notification.Channel {
				case "email":
					channel = emailChannel
				case "telegram":
					channel = telegramChannel
				default:
					r.logger.Warn().Str("channel", notification.Channel.String()).Msg("Unknown notification channel")
					d.Nack(false, false)
					continue
				}

				if err := channel.Send(ctx, notification.Message, notification.Destination); err != nil {
					r.logger.Error().Str("channel", notification.Channel.String()).Err(err).Msg("Failed to send notification")
					if updateErr := r.notificationService.UpdateStatus(ctx, notification.ID, "failed"); updateErr != nil {
						r.logger.Error().Err(updateErr).Msg("Failed to update status to failed")
					}
					d.Nack(false, true)
					continue
				}

				d.Ack(false)
				if updateErr := r.notificationService.UpdateStatus(ctx, notification.ID, "sent"); updateErr != nil {
					r.logger.Error().Err(updateErr).Msg("Failed to update status to sent")
				}
			}
		}
	}()
	return nil
}

func (r *RabbitMQ) Stop() error {
	r.logger.Info().Msg("Stopping RabbitMQ...")
	defer r.logger.Info().Msg("RabbitMQ stopped")

	r.logger.Info().Msg("Closing RabbitMQ channel...")
	if err := r.ch.Close(); err != nil {
		r.logger.Error().Err(err).Msg("Error closing channel")
		return err
	}
	r.logger.Info().Msg("Closing RabbitMQ connection...")
	if err := r.conn.Close(); err != nil {
		r.logger.Error().Err(err).Msg("Error closing connection")
		return err
	}

	r.wg.Wait() // ожидание завершения горутин

	return nil
}

func (r *RabbitMQ) SetNotificationService(svc Rabbit) {
	r.notificationService = svc
}

func (r *RabbitMQ) UpdateStatus(ctx context.Context, noteID uuid.UUID, status string) error {
	// реализуйте обновление статуса уведомления в своей системе, если необходимо
	return nil
}
