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

func New(ctx context.Context, log *slog.Logger, addr string, timeout time.Duration, retriesCount int) (*Client, error) {
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

func (c *Client) CreateEventType(ctx context.Context, req *eventspb.CreateEventTypeRequest) (*eventspb.CreateEventTypeResponse, error) {
	const op = "clients.events.CreateEventType"

	resp, err := c.api.CreateEventType(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) GetEventType(ctx context.Context, req *eventspb.GetEventTypeRequest) (*eventspb.GetEventTypeResponse, error) {
	const op = "clients.events.GetEventType"

	resp, err := c.api.GetEventType(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) ListEventTypes(ctx context.Context, req *eventspb.ListEventTypesRequest) (*eventspb.ListEventTypesResponse, error) {
	const op = "clients.events.ListEventTypes"

	resp, err := c.api.ListEventTypes(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) DeleteEventType(ctx context.Context, req *eventspb.DeleteEventTypeRequest) error {
	const op = "clients.events.DeleteEventType"

	_, err := c.api.DeleteEventType(ctx, req)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) CreateEvent(ctx context.Context, req *eventspb.CreateEventRequest) (*eventspb.CreateEventResponse, error) {
	const op = "clients.events.CreateEvent"

	resp, err := c.api.CreateEvent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) GetEvent(ctx context.Context, req *eventspb.GetEventRequest) (*eventspb.GetEventResponse, error) {
	const op = "clients.events.GetEvent"

	resp, err := c.api.GetEvent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) ListEventsByTree(ctx context.Context, req *eventspb.ListEventsByTreeRequest) (*eventspb.ListEventsByTreeResponse, error) {
	const op = "clients.events.ListEventsByTree"

	resp, err := c.api.ListEventsByTree(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) UpdateEvent(ctx context.Context, req *eventspb.UpdateEventRequest) (*eventspb.UpdateEventResponse, error) {
	const op = "clients.events.UpdateEvent"

	resp, err := c.api.UpdateEvent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) DeleteEvent(ctx context.Context, req *eventspb.DeleteEventRequest) error {
	const op = "clients.events.DeleteEvent"

	_, err := c.api.DeleteEvent(ctx, req)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
