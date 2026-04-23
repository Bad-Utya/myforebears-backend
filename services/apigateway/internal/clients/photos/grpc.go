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

func New(ctx context.Context, log *slog.Logger, addr string, timeout time.Duration, retriesCount int) (*Client, error) {
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

func (c *Client) UploadUserAvatar(ctx context.Context, req *photospb.UploadUserAvatarRequest) (*photospb.UploadUserAvatarResponse, error) {
	const op = "clients.photos.UploadUserAvatar"

	resp, err := c.api.UploadUserAvatar(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) GetUserAvatar(ctx context.Context, req *photospb.GetUserAvatarRequest) (*photospb.GetPhotoContentResponse, error) {
	const op = "clients.photos.GetUserAvatar"

	resp, err := c.api.GetUserAvatar(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) UploadPersonAvatar(ctx context.Context, req *photospb.UploadPersonAvatarRequest) (*photospb.UploadPersonAvatarResponse, error) {
	const op = "clients.photos.UploadPersonAvatar"

	resp, err := c.api.UploadPersonAvatar(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) GetPersonAvatar(ctx context.Context, req *photospb.GetPersonAvatarRequest) (*photospb.GetPhotoContentResponse, error) {
	const op = "clients.photos.GetPersonAvatar"

	resp, err := c.api.GetPersonAvatar(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) UploadPersonPhoto(ctx context.Context, req *photospb.UploadPersonPhotoRequest) (*photospb.UploadPersonPhotoResponse, error) {
	const op = "clients.photos.UploadPersonPhoto"

	resp, err := c.api.UploadPersonPhoto(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) ListPersonPhotos(ctx context.Context, req *photospb.ListPersonPhotosRequest) (*photospb.ListPersonPhotosResponse, error) {
	const op = "clients.photos.ListPersonPhotos"

	resp, err := c.api.ListPersonPhotos(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) UploadEventPhoto(ctx context.Context, req *photospb.UploadEventPhotoRequest) (*photospb.UploadEventPhotoResponse, error) {
	const op = "clients.photos.UploadEventPhoto"

	resp, err := c.api.UploadEventPhoto(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) ListEventPhotos(ctx context.Context, req *photospb.ListEventPhotosRequest) (*photospb.ListEventPhotosResponse, error) {
	const op = "clients.photos.ListEventPhotos"

	resp, err := c.api.ListEventPhotos(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) GetPhotoByID(ctx context.Context, req *photospb.GetPhotoByIDRequest) (*photospb.GetPhotoContentResponse, error) {
	const op = "clients.photos.GetPhotoByID"

	resp, err := c.api.GetPhotoByID(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) DeletePhotoByID(ctx context.Context, req *photospb.DeletePhotoByIDRequest) error {
	const op = "clients.photos.DeletePhotoByID"

	_, err := c.api.DeletePhotoByID(ctx, req)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
