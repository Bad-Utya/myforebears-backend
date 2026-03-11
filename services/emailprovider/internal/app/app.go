package app

import (
	"log/slog"

	"github.com/Bad-Utya/myforebears-backend/services/emailprovider/internal/consumer/rabbitmq"
	"github.com/Bad-Utya/myforebears-backend/services/emailprovider/internal/sender"
)

// App wires up all components of the email-provider service.
type App struct {
	log      *slog.Logger
	consumer *rabbitmq.Consumer
	sender   *sender.Sender
	stopCh   chan struct{}
}

// New creates the App, connecting to RabbitMQ and preparing the consumer.
func New(
	log *slog.Logger,
	rabbitURL string,
	exchange string,
	queue string,
	routingKey string,
	consumerTag string,
	smtpHost string,
	smtpPort int,
	smtpUsername string,
	smtpPassword string,
	smtpFrom string,
) *App {
	consumer, err := rabbitmq.New(log, rabbitURL, exchange, queue, routingKey, consumerTag)
	if err != nil {
		panic(err)
	}

	emailSender := sender.New(log, smtpHost, smtpPort, smtpUsername, smtpPassword, smtpFrom)

	return &App{
		log:      log,
		consumer: consumer,
		sender:   emailSender,
		stopCh:   make(chan struct{}),
	}
}

// MustRun starts consuming email messages. Panics if the consumer errors out.
func (a *App) MustRun() {
	if err := a.consumer.Consume(a.stopCh, a.handleEmail); err != nil {
		panic(err)
	}
}

// Stop signals the consumer to stop and closes the connection.
func (a *App) Stop() {
	close(a.stopCh)
	a.consumer.Close()
}

// handleEmail sends the email via SMTP.
func (a *App) handleEmail(msg rabbitmq.EmailMessage) error {
	return a.sender.Send(msg.To, msg.Subject, msg.Body)
}
