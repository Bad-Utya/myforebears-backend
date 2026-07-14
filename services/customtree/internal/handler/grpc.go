package handler

import (
	"context"
	"errors"
	customtreepb "github.com/Bad-Utya/myforebears-backend/gen/go/customtree"
	"github.com/Bad-Utya/myforebears-backend/services/customtree/internal/domain"
	"github.com/Bad-Utya/myforebears-backend/services/customtree/internal/service"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	customtreepb.UnimplementedCustomTreeServiceServer
	s *service.Service
}

func Register(g *grpc.Server, s *service.Service) {
	customtreepb.RegisterCustomTreeServiceServer(g, &Server{s: s})
}
func mapErr(err error) error {
	switch {
	case errors.Is(err, service.ErrInvalid):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, service.ErrConflict), errors.Is(err, service.ErrDeleteForbidden):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
func tree(t domain.Tree) *customtreepb.Tree {
	tags := make([]*customtreepb.Tag, 0, len(t.Tags))
	for _, x := range t.Tags {
		tags = append(tags, &customtreepb.Tag{Code: x.Code, Name: x.Name, Description: x.Description})
	}
	return &customtreepb.Tree{Id: t.ID.String(), CreatorId: int32(t.CreatorID), Name: t.Name, Description: t.Description, RelationDown: t.RelationDown, RelationUp: t.RelationUp, RootEntityId: t.RootEntityID.String(), IsViewRestricted: t.IsViewRestricted, IsPublicOnMainPage: t.IsPublicOnMainPage, CreatedAtUnix: t.CreatedAt.Unix(), Tags: tags, SimilarityScore: t.SimilarityScore}
}
func entity(e domain.Entity) *customtreepb.Entity {
	avatar := ""
	if e.AvatarPhotoID != nil {
		avatar = e.AvatarPhotoID.String()
	}
	return &customtreepb.Entity{Id: e.ID.String(), TreeId: e.TreeID.String(), Name: e.Name, Description: e.Description, AvatarPhotoId: avatar, CreatedAtUnix: e.CreatedAt.Unix()}
}
func photo(p domain.Photo) *customtreepb.Photo {
	return &customtreepb.Photo{Id: p.ID.String(), EntityId: p.EntityID.String(), FileName: p.FileName, MimeType: p.MIMEType, SizeBytes: p.SizeBytes, IsAvatar: p.IsAvatar, CreatedAtUnix: p.CreatedAt.Unix()}
}
func (s *Server) CreateTree(ctx context.Context, r *customtreepb.CreateTreeRequest) (*customtreepb.TreeResponse, error) {
	t, e, err := s.s.CreateTree(ctx, int(r.GetRequestUserId()), r.GetName(), r.GetDescription(), r.GetRelationDown(), r.GetRelationUp(), r.GetRootEntityName())
	if err != nil {
		return nil, mapErr(err)
	}
	return &customtreepb.TreeResponse{Tree: tree(t), RootEntity: entity(e)}, nil
}
func (s *Server) GetTree(ctx context.Context, r *customtreepb.GetTreeRequest) (*customtreepb.TreeResponse, error) {
	t, err := s.s.GetTree(ctx, r.GetTreeId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &customtreepb.TreeResponse{Tree: tree(t)}, nil
}
func trees(items []domain.Tree) *customtreepb.TreesResponse {
	out := make([]*customtreepb.Tree, 0, len(items))
	for _, t := range items {
		out = append(out, tree(t))
	}
	return &customtreepb.TreesResponse{Trees: out}
}
func (s *Server) ListTreesByOwner(ctx context.Context, r *customtreepb.ListTreesByOwnerRequest) (*customtreepb.TreesResponse, error) {
	x, e := s.s.ListTreesByOwner(ctx, int(r.GetRequestUserId()))
	if e != nil {
		return nil, mapErr(e)
	}
	return trees(x), nil
}
func (s *Server) ListPublicTreesByOwner(ctx context.Context, r *customtreepb.ListPublicTreesByOwnerRequest) (*customtreepb.TreesResponse, error) {
	x, e := s.s.ListPublicTreesByOwner(ctx, int(r.GetOwnerUserId()))
	if e != nil {
		return nil, mapErr(e)
	}
	return trees(x), nil
}
func (s *Server) ListRandomPublicTrees(ctx context.Context, r *customtreepb.ListRandomPublicTreesRequest) (*customtreepb.TreesResponse, error) {
	x, e := s.s.ListRandomPublicTrees(ctx, int(r.GetLimit()))
	if e != nil {
		return nil, mapErr(e)
	}
	return trees(x), nil
}
func (s *Server) SearchPublicTrees(ctx context.Context, r *customtreepb.SearchPublicTreesRequest) (*customtreepb.TreesResponse, error) {
	x, e := s.s.SearchPublicTrees(ctx, r.GetQuery(), r.GetTagCodes(), int(r.GetLimit()))
	if e != nil {
		return nil, mapErr(e)
	}
	return trees(x), nil
}
func (s *Server) SetTreeTags(ctx context.Context, r *customtreepb.SetTreeTagsRequest) (*customtreepb.TreeResponse, error) {
	t, e := s.s.SetTreeTags(ctx, int(r.GetRequestUserId()), r.GetTreeId(), r.GetTagCodes())
	if e != nil {
		return nil, mapErr(e)
	}
	return &customtreepb.TreeResponse{Tree: tree(t)}, nil
}
func (s *Server) UpdateTree(ctx context.Context, r *customtreepb.UpdateTreeRequest) (*customtreepb.TreeResponse, error) {
	id, e := uuid.Parse(r.GetTreeId())
	if e != nil {
		return nil, mapErr(service.ErrInvalid)
	}
	root, e := uuid.Parse(r.GetRootEntityId())
	if e != nil {
		return nil, mapErr(service.ErrInvalid)
	}
	t, e := s.s.UpdateTree(ctx, int(r.GetRequestUserId()), domain.Tree{ID: id, Name: r.GetName(), Description: r.GetDescription(), RelationDown: r.GetRelationDown(), RelationUp: r.GetRelationUp(), RootEntityID: root, IsViewRestricted: r.GetIsViewRestricted(), IsPublicOnMainPage: r.GetIsPublicOnMainPage()})
	if e != nil {
		return nil, mapErr(e)
	}
	return &customtreepb.TreeResponse{Tree: tree(t)}, nil
}
func (s *Server) DeleteTree(ctx context.Context, r *customtreepb.DeleteTreeRequest) (*customtreepb.Empty, error) {
	if e := s.s.DeleteTree(ctx, int(r.GetRequestUserId()), r.GetTreeId()); e != nil {
		return nil, mapErr(e)
	}
	return &customtreepb.Empty{}, nil
}
func (s *Server) GetTreeAccess(ctx context.Context, r *customtreepb.GetTreeRequest) (*customtreepb.TreeResponse, error) {
	return s.GetTree(ctx, r)
}
func (s *Server) AddAccessEmail(ctx context.Context, r *customtreepb.AccessEmailRequest) (*customtreepb.Empty, error) {
	if e := s.s.AddAccessEmail(ctx, r.GetTreeId(), r.GetEmail()); e != nil {
		return nil, mapErr(e)
	}
	return &customtreepb.Empty{}, nil
}
func (s *Server) ListAccessEmails(ctx context.Context, r *customtreepb.GetTreeRequest) (*customtreepb.AccessEmailsResponse, error) {
	x, e := s.s.ListAccessEmails(ctx, r.GetTreeId())
	if e != nil {
		return nil, mapErr(e)
	}
	return &customtreepb.AccessEmailsResponse{Emails: x}, nil
}
func (s *Server) DeleteAccessEmail(ctx context.Context, r *customtreepb.AccessEmailRequest) (*customtreepb.Empty, error) {
	if e := s.s.DeleteAccessEmail(ctx, r.GetTreeId(), r.GetEmail()); e != nil {
		return nil, mapErr(e)
	}
	return &customtreepb.Empty{}, nil
}
func (s *Server) IsAccessEmailAllowed(ctx context.Context, r *customtreepb.AccessEmailRequest) (*customtreepb.AccessAllowedResponse, error) {
	x, e := s.s.IsAccessEmailAllowed(ctx, r.GetTreeId(), r.GetEmail())
	if e != nil {
		return nil, mapErr(e)
	}
	return &customtreepb.AccessAllowedResponse{Allowed: x}, nil
}
func (s *Server) CreateEntity(ctx context.Context, r *customtreepb.CreateEntityRequest) (*customtreepb.EntityResponse, error) {
	e, err := s.s.CreateEntity(ctx, r.GetTreeId(), r.GetParentId(), r.GetName(), r.GetDescription())
	if err != nil {
		return nil, mapErr(err)
	}
	return &customtreepb.EntityResponse{Entity: entity(e)}, nil
}
func (s *Server) GetEntity(ctx context.Context, r *customtreepb.GetEntityRequest) (*customtreepb.EntityResponse, error) {
	e, err := s.s.GetEntity(ctx, r.GetTreeId(), r.GetEntityId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &customtreepb.EntityResponse{Entity: entity(e)}, nil
}
func (s *Server) ListEntities(ctx context.Context, r *customtreepb.GetTreeRequest) (*customtreepb.EntitiesResponse, error) {
	x, e := s.s.ListEntities(ctx, r.GetTreeId())
	if e != nil {
		return nil, mapErr(e)
	}
	out := make([]*customtreepb.Entity, 0, len(x))
	for _, v := range x {
		out = append(out, entity(v))
	}
	return &customtreepb.EntitiesResponse{Entities: out}, nil
}
func (s *Server) UpdateEntity(ctx context.Context, r *customtreepb.UpdateEntityRequest) (*customtreepb.EntityResponse, error) {
	e, err := s.s.UpdateEntity(ctx, r.GetTreeId(), r.GetEntityId(), r.GetName(), r.GetDescription())
	if err != nil {
		return nil, mapErr(err)
	}
	return &customtreepb.EntityResponse{Entity: entity(e)}, nil
}
func (s *Server) DeleteEntity(ctx context.Context, r *customtreepb.DeleteEntityRequest) (*customtreepb.Empty, error) {
	if e := s.s.DeleteEntity(ctx, r.GetTreeId(), r.GetEntityId()); e != nil {
		return nil, mapErr(e)
	}
	return &customtreepb.Empty{}, nil
}
func (s *Server) AddEdge(ctx context.Context, r *customtreepb.AddEdgeRequest) (*customtreepb.Empty, error) {
	if e := s.s.AddEdge(ctx, r.GetTreeId(), r.GetParentId(), r.GetChildId()); e != nil {
		return nil, mapErr(e)
	}
	return &customtreepb.Empty{}, nil
}
func (s *Server) RemoveEdge(ctx context.Context, r *customtreepb.RemoveEdgeRequest) (*customtreepb.Empty, error) {
	if e := s.s.RemoveEdge(ctx, r.GetTreeId(), r.GetParentId(), r.GetChildId()); e != nil {
		return nil, mapErr(e)
	}
	return &customtreepb.Empty{}, nil
}
func (s *Server) GetContent(ctx context.Context, r *customtreepb.GetTreeRequest) (*customtreepb.ContentResponse, error) {
	t, es, edges, e := s.s.GetContent(ctx, r.GetTreeId())
	if e != nil {
		return nil, mapErr(e)
	}
	outE := make([]*customtreepb.Entity, 0, len(es))
	for _, x := range es {
		outE = append(outE, entity(x))
	}
	outR := make([]*customtreepb.Edge, 0, len(edges))
	for _, x := range edges {
		outR = append(outR, &customtreepb.Edge{ParentId: x.ParentID.String(), ChildId: x.ChildID.String()})
	}
	return &customtreepb.ContentResponse{Tree: tree(t), Entities: outE, Edges: outR}, nil
}
func (s *Server) UploadPhoto(ctx context.Context, r *customtreepb.UploadPhotoRequest) (*customtreepb.PhotoResponse, error) {
	p, e := s.s.UploadPhoto(ctx, r.GetTreeId(), r.GetEntityId(), r.GetFileName(), r.GetMimeType(), r.GetContent(), r.GetIsAvatar())
	if e != nil {
		return nil, mapErr(e)
	}
	return &customtreepb.PhotoResponse{Photo: photo(p)}, nil
}
func (s *Server) ListPhotos(ctx context.Context, r *customtreepb.GetEntityRequest) (*customtreepb.PhotosResponse, error) {
	x, e := s.s.ListPhotos(ctx, r.GetTreeId(), r.GetEntityId())
	if e != nil {
		return nil, mapErr(e)
	}
	out := make([]*customtreepb.Photo, 0, len(x))
	for _, v := range x {
		out = append(out, photo(v))
	}
	return &customtreepb.PhotosResponse{Photos: out}, nil
}
func (s *Server) GetPhoto(ctx context.Context, r *customtreepb.GetPhotoRequest) (*customtreepb.PhotoContentResponse, error) {
	p, b, e := s.s.GetPhoto(ctx, r.GetTreeId(), r.GetEntityId(), r.GetPhotoId())
	if e != nil {
		return nil, mapErr(e)
	}
	return &customtreepb.PhotoContentResponse{Photo: photo(p), Content: b}, nil
}
func (s *Server) DeletePhoto(ctx context.Context, r *customtreepb.DeletePhotoRequest) (*customtreepb.Empty, error) {
	if e := s.s.DeletePhoto(ctx, r.GetTreeId(), r.GetEntityId(), r.GetPhotoId()); e != nil {
		return nil, mapErr(e)
	}
	return &customtreepb.Empty{}, nil
}
func (s *Server) RenderCoordinates(ctx context.Context, r *customtreepb.RenderRequest) (*customtreepb.CoordinatesResponse, error) {
	x, e := s.s.Render(ctx, r.GetTreeId(), r.GetRootEntityId())
	if e != nil {
		return nil, mapErr(e)
	}
	nodes := make([]*customtreepb.CoordinateNode, 0, len(x.Nodes))
	for _, n := range x.Nodes {
		a := ""
		if n.AvatarPhotoID != nil {
			a = n.AvatarPhotoID.String()
		}
		nodes = append(nodes, &customtreepb.CoordinateNode{EntityId: n.EntityID.String(), Name: n.Name, AvatarPhotoId: a, Layer: int32(n.Layer), X: n.X, Y: n.Y})
	}
	edges := make([]*customtreepb.CoordinateEdge, 0, len(x.Edges))
	for _, v := range x.Edges {
		edges = append(edges, &customtreepb.CoordinateEdge{ParentId: v.ParentID.String(), ChildId: v.ChildID.String(), LabelDown: v.LabelDown, LabelUp: v.LabelUp})
	}
	return &customtreepb.CoordinatesResponse{Nodes: nodes, Edges: edges, Width: x.Width, Height: x.Height}, nil
}
func (s *Server) RenderSVG(ctx context.Context, r *customtreepb.RenderRequest) (*customtreepb.SVGResponse, error) {
	x, e := s.s.Render(ctx, r.GetTreeId(), r.GetRootEntityId())
	if e != nil {
		return nil, mapErr(e)
	}
	return &customtreepb.SVGResponse{Content: x.SVG(), MimeType: "image/svg+xml"}, nil
}
