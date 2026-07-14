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

func (c *Client) CreateTree(ctx context.Context, requestUserID int, description *string, name string) (*familytreepb.CreateTreeResponse, error) {
	const op = "clients.familytree.CreateTree"

	resp, err := c.api.CreateTree(ctx, &familytreepb.CreateTreeRequest{
		RequestUserId: int32(requestUserID),
		Description:   stringOrEmpty(description),
		Name:          name,
	})
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

func (c *Client) ListPublicTreesByCreator(ctx context.Context, creatorID int) (*familytreepb.ListPublicTreesByCreatorResponse, error) {
	const op = "clients.familytree.ListPublicTreesByCreator"

	resp, err := c.api.ListPublicTreesByCreator(ctx, &familytreepb.ListPublicTreesByCreatorRequest{CreatorId: int32(creatorID)})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) ListRandomPublicTrees(ctx context.Context, limit int) (*familytreepb.ListRandomPublicTreesResponse, error) {
	const op = "clients.familytree.ListRandomPublicTrees"

	resp, err := c.api.ListRandomPublicTrees(ctx, &familytreepb.ListRandomPublicTreesRequest{Limit: int32(limit)})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) SearchPublicTrees(ctx context.Context, query string, tagCodes []string, limit int) (*familytreepb.SearchPublicTreesResponse, error) {
	const op = "clients.familytree.SearchPublicTrees"

	resp, err := c.api.SearchPublicTrees(ctx, &familytreepb.SearchPublicTreesRequest{Query: query, TagCodes: tagCodes, Limit: int32(limit)})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) ListTags(ctx context.Context) (*familytreepb.ListTagsResponse, error) {
	return c.api.ListTags(ctx, &familytreepb.ListTagsRequest{})
}
func (c *Client) SetTreeTags(ctx context.Context, userID int, treeID string, tags []string) (*familytreepb.GetTreeResponse, error) {
	return c.api.SetTreeTags(ctx, &familytreepb.SetTreeTagsRequest{RequestUserId: int32(userID), TreeId: treeID, TagCodes: tags})
}

func (c *Client) GetTree(ctx context.Context, treeID string) (*familytreepb.GetTreeResponse, error) {
	const op = "clients.familytree.GetTree"

	resp, err := c.api.GetTree(ctx, &familytreepb.GetTreeRequest{TreeId: treeID})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) GetTreeContent(ctx context.Context, treeID string) (*familytreepb.GetTreeContentResponse, error) {
	const op = "clients.familytree.GetTreeContent"

	resp, err := c.api.GetTreeContent(ctx, &familytreepb.GetTreeContentRequest{TreeId: treeID})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) GetTreeAccessInfo(ctx context.Context, treeID string) (*familytreepb.GetTreeAccessInfoResponse, error) {
	const op = "clients.familytree.GetTreeAccessInfo"

	resp, err := c.api.GetTreeAccessInfo(ctx, &familytreepb.GetTreeAccessInfoRequest{TreeId: treeID})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) IsTreeAccessEmailAllowed(ctx context.Context, treeID string, email string) (bool, error) {
	const op = "clients.familytree.IsTreeAccessEmailAllowed"

	resp, err := c.api.IsTreeAccessEmailAllowed(ctx, &familytreepb.IsTreeAccessEmailAllowedRequest{TreeId: treeID, Email: email})
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return resp.GetAllowed(), nil
}

func (c *Client) AddTreeAccessEmail(ctx context.Context, treeID string, email string) error {
	const op = "clients.familytree.AddTreeAccessEmail"

	_, err := c.api.AddTreeAccessEmail(ctx, &familytreepb.AddTreeAccessEmailRequest{
		TreeId: treeID,
		Email:  email,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) ListTreeAccessEmails(ctx context.Context, treeID string) (*familytreepb.ListTreeAccessEmailsResponse, error) {
	const op = "clients.familytree.ListTreeAccessEmails"

	resp, err := c.api.ListTreeAccessEmails(ctx, &familytreepb.ListTreeAccessEmailsRequest{
		TreeId: treeID,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) DeleteTreeAccessEmail(ctx context.Context, treeID string, email string) error {
	const op = "clients.familytree.DeleteTreeAccessEmail"

	_, err := c.api.DeleteTreeAccessEmail(ctx, &familytreepb.DeleteTreeAccessEmailRequest{
		TreeId: treeID,
		Email:  email,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) UpdateTreeSettings(ctx context.Context, treeID string, isViewRestricted bool, isPublicOnMainPage bool, name string, description *string) (*familytreepb.UpdateTreeSettingsResponse, error) {
	const op = "clients.familytree.UpdateTreeSettings"

	resp, err := c.api.UpdateTreeSettings(ctx, &familytreepb.UpdateTreeSettingsRequest{
		TreeId:             treeID,
		IsViewRestricted:   isViewRestricted,
		IsPublicOnMainPage: isPublicOnMainPage,
		Name:               name,
		Description:        stringOrEmpty(description),
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) UpdateTreeRootPerson(ctx context.Context, treeID string, rootPersonID string) (*familytreepb.UpdateTreeRootPersonResponse, error) {
	const op = "clients.familytree.UpdateTreeRootPerson"

	resp, err := c.api.UpdateTreeRootPerson(ctx, &familytreepb.UpdateTreeRootPersonRequest{
		TreeId:       treeID,
		RootPersonId: rootPersonID,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func stringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func (c *Client) ListPersonsByTree(ctx context.Context, treeID string) (*familytreepb.ListPersonsByTreeResponse, error) {
	const op = "clients.familytree.ListPersonsByTree"

	resp, err := c.api.ListPersonsByTree(ctx, &familytreepb.ListPersonsByTreeRequest{TreeId: treeID})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) GetPerson(ctx context.Context, treeID string, personID string) (*familytreepb.GetPersonResponse, error) {
	const op = "clients.familytree.GetPersonInTree"

	resp, err := c.api.GetPerson(ctx, &familytreepb.GetPersonRequest{
		TreeId:   treeID,
		PersonId: personID,
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

func (c *Client) UpdatePerson(ctx context.Context, req *familytreepb.UpdatePersonRequest) (*familytreepb.UpdatePersonResponse, error) {
	const op = "clients.familytree.UpdatePerson"

	resp, err := c.api.UpdatePerson(ctx, req)
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

func (c *Client) ExportTreeGEDCOM(ctx context.Context, requestUserID int, treeID string) (*familytreepb.ExportTreeGEDCOMResponse, error) {
	const op = "clients.familytree.ExportTreeGEDCOM"

	resp, err := c.api.ExportTreeGEDCOM(ctx, &familytreepb.ExportTreeGEDCOMRequest{
		TreeId:        treeID,
		RequestUserId: int32(requestUserID),
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) ImportTreeGEDCOM(ctx context.Context, requestUserID int, gedcomContent string) (*familytreepb.ImportTreeGEDCOMResponse, error) {
	const op = "clients.familytree.ImportTreeGEDCOM"

	resp, err := c.api.ImportTreeGEDCOM(ctx, &familytreepb.ImportTreeGEDCOMRequest{
		GedcomContent: gedcomContent,
		RequestUserId: int32(requestUserID),
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

func (c *Client) CreatePublicPerson(ctx context.Context, userID int) (*familytreepb.PublicPersonResponse, error) {
	return c.api.CreatePublicPerson(ctx, &familytreepb.CreatePublicPersonRequest{RequestUserId: int32(userID)})
}
func (c *Client) CreatePublicPersonSnapshot(ctx context.Context, req *familytreepb.CreatePublicPersonSnapshotRequest) (*familytreepb.PublicPersonResponse, error) {
	return c.api.CreatePublicPersonSnapshot(ctx, req)
}
func (c *Client) GetPublicPerson(ctx context.Context, id string) (*familytreepb.PublicPersonResponse, error) {
	return c.api.GetPublicPerson(ctx, &familytreepb.GetPublicPersonRequest{PublicPersonId: id})
}
func (c *Client) ListRandomPublicPersons(ctx context.Context, limit int) (*familytreepb.PublicPersonsResponse, error) {
	return c.api.ListRandomPublicPersons(ctx, &familytreepb.ListRandomPublicPersonsRequest{Limit: int32(limit)})
}
func (c *Client) ListPublicPersonsByOwner(ctx context.Context, ownerID, limit int) (*familytreepb.PublicPersonsResponse, error) {
	return c.api.ListPublicPersonsByOwner(ctx, &familytreepb.ListPublicPersonsByOwnerRequest{OwnerUserId: int32(ownerID), Limit: int32(limit)})
}
func (c *Client) SearchPublicPersons(ctx context.Context, q string, tags []string, limit int) (*familytreepb.PublicPersonsResponse, error) {
	return c.api.SearchPublicPersons(ctx, &familytreepb.SearchPublicPersonsRequest{Query: q, TagCodes: tags, Limit: int32(limit)})
}
func (c *Client) SetPublicPersonTags(ctx context.Context, userID int, id string, tags []string) (*familytreepb.PublicPersonResponse, error) {
	return c.api.SetPublicPersonTags(ctx, &familytreepb.SetPublicPersonTagsRequest{RequestUserId: int32(userID), PublicPersonId: id, TagCodes: tags})
}
func (c *Client) UpdatePublicPerson(ctx context.Context, req *familytreepb.UpdatePublicPersonRequest) (*familytreepb.PublicPersonResponse, error) {
	return c.api.UpdatePublicPerson(ctx, req)
}
func (c *Client) DeletePublicPerson(ctx context.Context, userID int, id string) error {
	_, err := c.api.DeletePublicPerson(ctx, &familytreepb.DeletePublicPersonRequest{RequestUserId: int32(userID), PublicPersonId: id})
	return err
}
func (c *Client) ImportPublicPersonIntoTree(ctx context.Context, req *familytreepb.ImportPublicPersonIntoTreeRequest) (*familytreepb.ImportPublicPersonIntoTreeResponse, error) {
	return c.api.ImportPublicPersonIntoTree(ctx, req)
}
func (c *Client) CreateTreeFromPublicPerson(ctx context.Context, req *familytreepb.CreateTreeFromPublicPersonRequest) (*familytreepb.CreateTreeFromPublicPersonResponse, error) {
	return c.api.CreateTreeFromPublicPerson(ctx, req)
}
