package app

import (
	"log/slog"

	"github.com/Bad-Utya/myforebears-backend/services/emailprovider/internal/consumer/rabbitmq"
)

// App wires up all components of the email-provider service.
type App struct {
	log      *slog.Logger
	consumer *rabbitmq.Consumer
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
) *App {
	consumer, err := rabbitmq.New(log, rabbitURL, exchange, queue, routingKey, consumerTag)
	if err != nil {
		panic(err)
	}

	return &App{
		log:      log,
		consumer: consumer,
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

// handleEmail is the business logic that processes an outbound email.
// Replace / extend this with your actual SMTP / SES / etc. sending logic.
func (a *App) handleEmail(msg rabbitmq.EmailMessage) error {
	a.log.Info("sending email",
		slog.String("to", msg.To),
		slog.String("subject", msg.Subject),
	)

	return nil
}
