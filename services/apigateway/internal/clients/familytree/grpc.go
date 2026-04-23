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

func New(ctx context.Context, log *slog.Logger, addr string, timeout time.Duration, retriesCount int) (*Client, error) {
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

func (c *Client) CreateTree(ctx context.Context, requestUserID int) (*familytreepb.CreateTreeResponse, error) {
	const op = "clients.familytree.CreateTree"

	resp, err := c.api.CreateTree(ctx, &familytreepb.CreateTreeRequest{RequestUserId: int32(requestUserID)})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) ListTreesByCreator(ctx context.Context, requestUserID int) (*familytreepb.ListTreesByCreatorResponse, error) {
	const op = "clients.familytree.ListTreesByCreator"

	resp, err := c.api.ListTreesByCreator(ctx, &familytreepb.ListTreesByCreatorRequest{RequestUserId: int32(requestUserID)})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) GetTree(ctx context.Context, requestUserID int, treeID string) (*familytreepb.GetTreeResponse, error) {
	const op = "clients.familytree.GetTree"

	resp, err := c.api.GetTree(ctx, &familytreepb.GetTreeRequest{TreeId: treeID, RequestUserId: int32(requestUserID)})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) GetTreeContent(ctx context.Context, requestUserID int, treeID string) (*familytreepb.GetTreeContentResponse, error) {
	const op = "clients.familytree.GetTreeContent"

	resp, err := c.api.GetTreeContent(ctx, &familytreepb.GetTreeContentRequest{TreeId: treeID, RequestUserId: int32(requestUserID)})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) UpdateTreeSettings(ctx context.Context, requestUserID int, treeID string, isViewRestricted bool, isPublicOnMainPage bool) (*familytreepb.UpdateTreeSettingsResponse, error) {
	const op = "clients.familytree.UpdateTreeSettings"

	resp, err := c.api.UpdateTreeSettings(ctx, &familytreepb.UpdateTreeSettingsRequest{
		RequestUserId:      int32(requestUserID),
		TreeId:             treeID,
		IsViewRestricted:   isViewRestricted,
		IsPublicOnMainPage: isPublicOnMainPage,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) ListPersonsByTree(ctx context.Context, requestUserID int, treeID string) (*familytreepb.ListPersonsByTreeResponse, error) {
	const op = "clients.familytree.ListPersonsByTree"

	resp, err := c.api.ListPersonsByTree(ctx, &familytreepb.ListPersonsByTreeRequest{TreeId: treeID, RequestUserId: int32(requestUserID)})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) GetPersonInTree(ctx context.Context, requestUserID int, treeID string, personID string) (*familytreepb.GetPersonInTreeResponse, error) {
	const op = "clients.familytree.GetPersonInTree"

	resp, err := c.api.GetPersonInTree(ctx, &familytreepb.GetPersonInTreeRequest{
		RequestUserId: int32(requestUserID),
		TreeId:        treeID,
		PersonId:      personID,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) AddParent(ctx context.Context, req *familytreepb.AddParentRequest) (*familytreepb.AddParentResponse, error) {
	const op = "clients.familytree.AddParent"

	resp, err := c.api.AddParent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) AddChild(ctx context.Context, req *familytreepb.AddChildRequest) (*familytreepb.AddChildResponse, error) {
	const op = "clients.familytree.AddChild"

	resp, err := c.api.AddChild(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) AddPartner(ctx context.Context, req *familytreepb.AddPartnerRequest) (*familytreepb.AddPartnerResponse, error) {
	const op = "clients.familytree.AddPartner"

	resp, err := c.api.AddPartner(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) UpdatePersonName(ctx context.Context, req *familytreepb.UpdatePersonNameRequest) (*familytreepb.UpdatePersonNameResponse, error) {
	const op = "clients.familytree.UpdatePersonName"

	resp, err := c.api.UpdatePersonName(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) DeletePersonInTree(ctx context.Context, req *familytreepb.DeletePersonInTreeRequest) error {
	const op = "clients.familytree.DeletePersonInTree"

	_, err := c.api.DeletePersonInTree(ctx, req)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
