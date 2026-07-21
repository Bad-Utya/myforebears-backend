package customtree

import (
	"context"
	"fmt"
	customtreepb "github.com/Bad-Utya/myforebears-backend/gen/go/customtree"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"time"
)

type Client struct {
	api  customtreepb.CustomTreeServiceClient
	conn *grpc.ClientConn
}

func New(ctx context.Context, log *slog.Logger, addr string, timeout time.Duration, retries int) (*Client, error) {
	_ = ctx
	_ = log
	_ = timeout
	_ = retries
	c, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("customtree client: %w", err)
	}
	return &Client{customtreepb.NewCustomTreeServiceClient(c), c}, nil
}
func (c *Client) Close() error { return c.conn.Close() }
func (c *Client) CreateTree(x context.Context, r *customtreepb.CreateTreeRequest) (*customtreepb.TreeResponse, error) {
	return c.api.CreateTree(x, r)
}
func (c *Client) GetTree(x context.Context, id string) (*customtreepb.TreeResponse, error) {
	return c.api.GetTree(x, &customtreepb.GetTreeRequest{TreeId: id})
}
func (c *Client) ListMine(x context.Context, user int) (*customtreepb.TreesResponse, error) {
	return c.api.ListTreesByOwner(x, &customtreepb.ListTreesByOwnerRequest{RequestUserId: int32(user)})
}
func (c *Client) ListByOwner(x context.Context, user int) (*customtreepb.TreesResponse, error) {
	return c.api.ListPublicTreesByOwner(x, &customtreepb.ListPublicTreesByOwnerRequest{OwnerUserId: int32(user)})
}
func (c *Client) Random(x context.Context, n int) (*customtreepb.TreesResponse, error) {
	return c.api.ListRandomPublicTrees(x, &customtreepb.ListRandomPublicTreesRequest{Limit: int32(n)})
}
func (c *Client) Search(x context.Context, q string, tags []string, n int) (*customtreepb.TreesResponse, error) {
	return c.api.SearchPublicTrees(x, &customtreepb.SearchPublicTreesRequest{Query: q, TagCodes: tags, Limit: int32(n)})
}
func (c *Client) SetTags(x context.Context, user int, id string, tags []string) (*customtreepb.TreeResponse, error) {
	return c.api.SetTreeTags(x, &customtreepb.SetTreeTagsRequest{RequestUserId: int32(user), TreeId: id, TagCodes: tags})
}
func (c *Client) UpdateTree(x context.Context, r *customtreepb.UpdateTreeRequest) (*customtreepb.TreeResponse, error) {
	return c.api.UpdateTree(x, r)
}
func (c *Client) DeleteTree(x context.Context, user int, id string) error {
	_, e := c.api.DeleteTree(x, &customtreepb.DeleteTreeRequest{RequestUserId: int32(user), TreeId: id})
	return e
}
func (c *Client) AddEmail(x context.Context, id, email string) error {
	_, e := c.api.AddAccessEmail(x, &customtreepb.AccessEmailRequest{TreeId: id, Email: email})
	return e
}
func (c *Client) ListEmails(x context.Context, id string) (*customtreepb.AccessEmailsResponse, error) {
	return c.api.ListAccessEmails(x, &customtreepb.GetTreeRequest{TreeId: id})
}
func (c *Client) DeleteEmail(x context.Context, id, email string) error {
	_, e := c.api.DeleteAccessEmail(x, &customtreepb.AccessEmailRequest{TreeId: id, Email: email})
	return e
}
func (c *Client) EmailAllowed(x context.Context, id, email string) (bool, error) {
	r, e := c.api.IsAccessEmailAllowed(x, &customtreepb.AccessEmailRequest{TreeId: id, Email: email})
	return r.GetAllowed(), e
}
func (c *Client) CreateEntity(x context.Context, r *customtreepb.CreateEntityRequest) (*customtreepb.EntityResponse, error) {
	return c.api.CreateEntity(x, r)
}
func (c *Client) AddParent(x context.Context, r *customtreepb.AddParentRequest) (*customtreepb.EntityResponse, error) {
	return c.api.AddParent(x, r)
}
func (c *Client) GetEntity(x context.Context, id, eid string) (*customtreepb.EntityResponse, error) {
	return c.api.GetEntity(x, &customtreepb.GetEntityRequest{TreeId: id, EntityId: eid})
}
func (c *Client) ListEntities(x context.Context, id string) (*customtreepb.EntitiesResponse, error) {
	return c.api.ListEntities(x, &customtreepb.GetTreeRequest{TreeId: id})
}
func (c *Client) UpdateEntity(x context.Context, r *customtreepb.UpdateEntityRequest) (*customtreepb.EntityResponse, error) {
	return c.api.UpdateEntity(x, r)
}
func (c *Client) DeleteEntity(x context.Context, id, eid string) error {
	_, e := c.api.DeleteEntity(x, &customtreepb.DeleteEntityRequest{TreeId: id, EntityId: eid})
	return e
}
func (c *Client) AddEdge(x context.Context, r *customtreepb.AddEdgeRequest) error {
	_, e := c.api.AddEdge(x, r)
	return e
}
func (c *Client) RemoveEdge(x context.Context, r *customtreepb.RemoveEdgeRequest) error {
	_, e := c.api.RemoveEdge(x, r)
	return e
}
func (c *Client) Content(x context.Context, id string) (*customtreepb.ContentResponse, error) {
	return c.api.GetContent(x, &customtreepb.GetTreeRequest{TreeId: id})
}
func (c *Client) UploadPhoto(x context.Context, r *customtreepb.UploadPhotoRequest) (*customtreepb.PhotoResponse, error) {
	return c.api.UploadPhoto(x, r)
}
func (c *Client) ListPhotos(x context.Context, id, eid string) (*customtreepb.PhotosResponse, error) {
	return c.api.ListPhotos(x, &customtreepb.GetEntityRequest{TreeId: id, EntityId: eid})
}
func (c *Client) GetPhoto(x context.Context, id, eid, pid string) (*customtreepb.PhotoContentResponse, error) {
	return c.api.GetPhoto(x, &customtreepb.GetPhotoRequest{TreeId: id, EntityId: eid, PhotoId: pid})
}
func (c *Client) DeletePhoto(x context.Context, id, eid, pid string) error {
	_, e := c.api.DeletePhoto(x, &customtreepb.DeletePhotoRequest{TreeId: id, EntityId: eid, PhotoId: pid})
	return e
}
func (c *Client) Coordinates(x context.Context, id, root string) (*customtreepb.CoordinatesResponse, error) {
	return c.api.RenderCoordinates(x, &customtreepb.RenderRequest{TreeId: id, RootEntityId: root})
}
func (c *Client) SVG(x context.Context, id, root string) (*customtreepb.SVGResponse, error) {
	return c.api.RenderSVG(x, &customtreepb.RenderRequest{TreeId: id, RootEntityId: root})
}
