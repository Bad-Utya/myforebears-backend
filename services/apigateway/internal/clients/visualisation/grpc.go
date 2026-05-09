package visualisation

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	visualisationpb "github.com/Bad-Utya/myforebears-backend/gen/go/visualisation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api  visualisationpb.VisualisationServiceClient
	conn *grpc.ClientConn
	log  *slog.Logger
}

func New(ctx context.Context, log *slog.Logger, addr string, timeout time.Duration, retriesCount int) (*Client, error) {
	const op = "clients.visualisation.New"

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		api:  visualisationpb.NewVisualisationServiceClient(conn),
		conn: conn,
		log:  log,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) CreateAncestorsVisualisation(ctx context.Context, req *visualisationpb.CreateLineageVisualisationRequest) (*visualisationpb.CreateVisualisationResponse, error) {
	const op = "clients.visualisation.CreateAncestorsVisualisation"

	resp, err := c.api.CreateAncestorsVisualisation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) CreateDescendantsVisualisation(ctx context.Context, req *visualisationpb.CreateLineageVisualisationRequest) (*visualisationpb.CreateVisualisationResponse, error) {
	const op = "clients.visualisation.CreateDescendantsVisualisation"

	resp, err := c.api.CreateDescendantsVisualisation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) CreateAncestorsAndDescendantsVisualisation(ctx context.Context, req *visualisationpb.CreateLineageVisualisationRequest) (*visualisationpb.CreateVisualisationResponse, error) {
	const op = "clients.visualisation.CreateAncestorsAndDescendantsVisualisation"

	resp, err := c.api.CreateAncestorsAndDescendantsVisualisation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) CreateFullVisualisation(ctx context.Context, req *visualisationpb.CreateFullVisualisationRequest) (*visualisationpb.CreateVisualisationResponse, error) {
	const op = "clients.visualisation.CreateFullVisualisation"

	resp, err := c.api.CreateFullVisualisation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) ListTreeVisualisations(ctx context.Context, req *visualisationpb.ListTreeVisualisationsRequest) (*visualisationpb.ListTreeVisualisationsResponse, error) {
	const op = "clients.visualisation.ListTreeVisualisations"

	resp, err := c.api.ListTreeVisualisations(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) GetVisualisationByID(ctx context.Context, req *visualisationpb.GetVisualisationByIDRequest) (*visualisationpb.GetVisualisationContentResponse, error) {
	const op = "clients.visualisation.GetVisualisationByID"

	resp, err := c.api.GetVisualisationByID(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) DeleteVisualisationByID(ctx context.Context, req *visualisationpb.DeleteVisualisationByIDRequest) error {
	const op = "clients.visualisation.DeleteVisualisationByID"

	_, err := c.api.DeleteVisualisationByID(ctx, req)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) RenderCoordinatesForClient(ctx context.Context, req *visualisationpb.RenderCoordinatesForClientRequest) (*visualisationpb.RenderCoordinatesForClientResponse, error) {
	const op = "clients.visualisation.RenderCoordinatesForClient"

	resp, err := c.api.RenderCoordinatesForClient(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}
