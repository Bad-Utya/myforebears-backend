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

func (c *Client) GetEventTreeID(ctx context.Context, requestUserID int, eventID string) (string, error) {
	const op = "clients.events.GetEventTreeID"

	resp, err := c.api.GetEvent(ctx, &eventspb.GetEventRequest{
		RequestUserId: int32(requestUserID),
		EventId:       eventID,
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return resp.GetEvent().GetTreeId(), nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
