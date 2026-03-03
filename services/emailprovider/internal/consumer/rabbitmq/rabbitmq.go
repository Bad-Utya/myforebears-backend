package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

// EmailMessage is the payload published to the email queue.
type EmailMessage struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// Consumer holds an open AMQP connection/channel and is responsible for
// declaring the topology and delivering messages to a handler function.
type Consumer struct {
	log         *slog.Logger
	conn        *amqp.Connection
	ch          *amqp.Channel
	exchange    string
	queue       string
	routingKey  string
	consumerTag string
}

// New connects to RabbitMQ, declares the exchange + queue, and binds them.
// The topology is idempotent – safe to call on every start-up.
func New(
	log *slog.Logger,
	url string,
	exchange string,
	queue string,
	routingKey string,
	consumerTag string,
) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq: open channel: %w", err)
	}

	// Declare a durable direct exchange.
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
		return nil, fmt.Errorf("rabbitmq: declare exchange: %w", err)
	}

	// Declare a durable queue.
	if _, err = ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("rabbitmq: declare queue: %w", err)
	}

	// Bind the queue to the exchange with the routing key.
	if err = ch.QueueBind(queue, routingKey, exchange, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("rabbitmq: bind queue: %w", err)
	}

	log.Info("rabbitmq: connected",
		slog.String("exchange", exchange),
		slog.String("queue", queue),
	)

	return &Consumer{
		log:         log,
		conn:        conn,
		ch:          ch,
		exchange:    exchange,
		queue:       queue,
		routingKey:  routingKey,
		consumerTag: consumerTag,
	}, nil
}

// Consume starts blocking delivery processing and stops when stopCh is closed.
// Each delivery is acknowledged after the handler returns without an error; on
// error the message is negatively acknowledged and requeued.
func (c *Consumer) Consume(stopCh <-chan struct{}, handler func(msg EmailMessage) error) error {
	// Prefetch one message at a time so failed messages don't pile up.
	if err := c.ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("rabbitmq: set qos: %w", err)
	}

	deliveries, err := c.ch.Consume(
		c.queue,
		c.consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("rabbitmq: start consume: %w", err)
	}

	c.log.Info("rabbitmq: waiting for messages", slog.String("queue", c.queue))

	for {
		select {
		case <-stopCh:
			c.log.Info("rabbitmq: consumer stopped")
			return nil
		case d, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("rabbitmq: deliveries channel closed")
			}

			var msg EmailMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				c.log.Error("rabbitmq: unmarshal message", slog.String("error", err.Error()))
				_ = d.Nack(false, false)
				continue
			}

			if err := handler(msg); err != nil {
				c.log.Error("rabbitmq: handler error",
					slog.String("error", err.Error()),
					slog.String("to", msg.To),
				)
				_ = d.Nack(false, true)
			} else {
				_ = d.Ack(false)
			}
		}
	}
}

// Close gracefully shuts down the channel and connection.
func (c *Consumer) Close() {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.log.Info("rabbitmq: connection closed")
}
