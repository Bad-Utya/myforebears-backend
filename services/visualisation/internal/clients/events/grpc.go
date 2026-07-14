package events

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	eventspb "github.com/Bad-Utya/myforebears-backend/gen/go/events"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api  eventspb.EventsServiceClient
	conn *grpc.ClientConn
	log  *slog.Logger
}

func New(log *slog.Logger, addr string, timeout time.Duration, retriesCount int) (*Client, error) {
	const op = "clients.events.New"

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		api:  eventspb.NewEventsServiceClient(conn),
		conn: conn,
		log:  log,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) ListEventsByTree(ctx context.Context, req *eventspb.ListEventsByTreeRequest) (*eventspb.ListEventsByTreeResponse, error) {
	const op = "clients.events.ListEventsByTree"

	resp, err := c.api.ListEventsByTree(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}
