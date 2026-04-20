package familytree

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	familytreepb "github.com/Bad-Utya/myforebears-backend/gen/go/familytree"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api  familytreepb.FamilyTreeServiceClient
	conn *grpc.ClientConn
	log  *slog.Logger
}

func New(log *slog.Logger, addr string, timeout time.Duration, retriesCount int) (*Client, error) {
	const op = "clients.familytree.New"

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		api:  familytreepb.NewFamilyTreeServiceClient(conn),
		conn: conn,
		log:  log,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) ValidatePersonsInTree(ctx context.Context, requestUserID int, treeID string, personIDs []string) error {
	const op = "clients.familytree.ValidatePersonsInTree"

	_, err := c.api.ValidatePersonsInTree(ctx, &familytreepb.ValidatePersonsInTreeRequest{
		RequestUserId: int32(requestUserID),
		TreeId:        treeID,
		PersonIds:     personIDs,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
