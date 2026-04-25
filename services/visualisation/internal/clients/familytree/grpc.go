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

func (c *Client) GetPerson(ctx context.Context, treeID string, personID string) error {
	const op = "clients.familytree.GetPerson"

	_, err := c.api.GetPerson(ctx, &familytreepb.GetPersonRequest{
		TreeId:   treeID,
		PersonId: personID,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) GetTreeContent(ctx context.Context, treeID string) (*familytreepb.GetTreeContentResponse, error) {
	const op = "clients.familytree.GetTreeContent"

	resp, err := c.api.GetTreeContent(ctx, &familytreepb.GetTreeContentRequest{TreeId: treeID})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) GetTreeCreatorID(ctx context.Context, treeID string) (int, error) {
	const op = "clients.familytree.GetTreeCreatorID"

	resp, err := c.api.GetTreeAccessInfo(ctx, &familytreepb.GetTreeAccessInfoRequest{TreeId: treeID})
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	tree := resp.GetTree()
	if tree == nil || tree.GetCreatorId() <= 0 {
		return 0, fmt.Errorf("%s: invalid tree access info", op)
	}

	return int(tree.GetCreatorId()), nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
