package notificationService

import (
	"context"
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramChannel struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

func NewTelegramChannel(botToken string, chatID int64) (*TelegramChannel, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}
	return &TelegramChannel{
		bot:    bot,
		chatID: chatID,
	}, nil
}

func (t *TelegramChannel) Send(ctx context.Context, message, destination string) error {
	chatID, err := strconv.ParseInt(destination, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid chat ID: %w", err)
	}
	msg := tgbotapi.NewMessage(chatID, message)
	done := make(chan error, 1)
	go func() {
		_, err := t.bot.Send(msg)
		done <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return err
		}
		return nil
	}
}
