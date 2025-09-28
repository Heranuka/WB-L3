package notificationService

import (
	"context"
	"fmt"
	"net/smtp"
)

type EmailChannel struct {
	smtpServer   string
	smtpPort     int
	smtpEmail    string
	smtpPassword string
}

func NewEmailChannel(smtpPort int, smtpServer, smtpEmail, smtpPassword string) *EmailChannel {
	return &EmailChannel{
		smtpServer:   smtpServer,
		smtpPort:     smtpPort,
		smtpEmail:    smtpEmail,
		smtpPassword: smtpPassword,
	}
}

func (e *EmailChannel) Send(ctx context.Context, message, destination string) error {
	from := e.smtpEmail
	pass := e.smtpPassword
	to := []string{destination}

	body := "To: " + destination + "\r\n" +
		"Subject: Notification\r\n" +
		"\r\n" +
		message + "\r\n"

	auth := smtp.PlainAuth("", from, pass, e.smtpServer)
	addr := fmt.Sprintf("%s:%d", e.smtpServer, e.smtpPort)

	done := make(chan error, 1)
	go func() {
		err := smtp.SendMail(addr, auth, from, to, []byte(body))
		done <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to send mail: %w", err)
		}
		return nil
	}
}
