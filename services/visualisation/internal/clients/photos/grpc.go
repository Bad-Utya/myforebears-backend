package photos

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	photospb "github.com/Bad-Utya/myforebears-backend/gen/go/photos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api  photospb.PhotosServiceClient
	conn *grpc.ClientConn
	log  *slog.Logger
}

func New(log *slog.Logger, addr string, timeout time.Duration, retriesCount int) (*Client, error) {
	const op = "clients.photos.New"

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		api:  photospb.NewPhotosServiceClient(conn),
		conn: conn,
		log:  log,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) GetPersonAvatar(ctx context.Context, req *photospb.GetPersonAvatarRequest) (*photospb.GetPhotoContentResponse, error) {
	const op = "clients.photos.GetPersonAvatar"

	resp, err := c.api.GetPersonAvatar(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}
