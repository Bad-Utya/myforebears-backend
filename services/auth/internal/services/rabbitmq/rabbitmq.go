package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	sendCodeText = `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="margin:0;padding:0;background:#f4f4f7;font-family:Arial,Helvetica,sans-serif">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background:#f4f4f7;padding:40px 0">
    <tr><td align="center">
      <table width="480" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:8px;overflow:hidden">
        <tr>
          <td style="background:#2d6a4f;padding:24px;text-align:center;color:#ffffff;font-size:22px;font-weight:bold">
            Rooots
          </td>
        </tr>
        <tr>
          <td style="padding:32px 32px 16px;text-align:center;color:#333333;font-size:16px">
            Your verification code
          </td>
        </tr>
        <tr>
          <td style="padding:0 32px 32px;text-align:center">
            <div style="display:inline-block;background:#f0faf4;border:2px dashed #2d6a4f;border-radius:8px;padding:16px 32px;font-size:32px;font-weight:bold;letter-spacing:6px;color:#2d6a4f">
              %s
            </div>
          </td>
        </tr>
        <tr>
          <td style="padding:0 32px 32px;text-align:center;color:#888888;font-size:13px">
            If you didn't request this code, just ignore this email.
          </td>
        </tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`
)

// emailMessage is the wire format consumed by the email-provider service.
type emailMessage struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// Publisher holds an open AMQP connection/channel and publishes messages to an
// exchange.  It declares the exchange on construction so startup fails fast if
// RabbitMQ is unreachable.
type Publisher struct {
	log        *slog.Logger
	conn       *amqp.Connection
	ch         *amqp.Channel
	exchange   string
	routingKey string
}

// New connects to RabbitMQ and declares the exchange.
func New(log *slog.Logger, url string, exchange string, routingKey string) (*Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq publisher: dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq publisher: open channel: %w", err)
	}

	// Declare the same durable direct exchange that the consumer expects.
	if err = ch.ExchangeDeclare(
		exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("rabbitmq publisher: declare exchange: %w", err)
	}

	log.Info("rabbitmq publisher: ready",
		slog.String("exchange", exchange),
		slog.String("routing_key", routingKey),
	)

	return &Publisher{
		log:        log,
		conn:       conn,
		ch:         ch,
		exchange:   exchange,
		routingKey: routingKey,
	}, nil
}

func (p *Publisher) PublishCode(ctx context.Context, to string, code string) error {
	body := fmt.Sprintf(sendCodeText, code)
	return p.publishEmail(ctx, to, "Rooots — Verification Code", body)
}

// PublishEmail serializes the email task and publishes it as a persistent
// message so it survives a RabbitMQ restart.
// It satisfies the auth.EmailPublisher interface.
func (p *Publisher) publishEmail(ctx context.Context, to string, subject string, body string) error {
	msg := emailMessage{To: to, Subject: subject, Body: body}
	raw, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("rabbitmq publisher: marshal: %w", err)
	}

	err = p.ch.PublishWithContext(
		ctx,
		p.exchange,
		p.routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         raw,
		},
	)
	if err != nil {
		return fmt.Errorf("rabbitmq publisher: publish: %w", err)
	}

	p.log.Info("rabbitmq publisher: message published",
		slog.String("to", to),
		slog.String("subject", subject),
	)

	return nil
}

// Close gracefully shuts down the channel and connection.
func (p *Publisher) Close() {
	if p.ch != nil {
		_ = p.ch.Close()
	}
	if p.conn != nil {
		_ = p.conn.Close()
	}
	p.log.Info("rabbitmq publisher: connection closed")
}
