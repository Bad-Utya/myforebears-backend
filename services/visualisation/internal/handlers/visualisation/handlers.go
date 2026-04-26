package visualisation

import (
	"context"

	visualisationpb "github.com/Bad-Utya/myforebears-backend/gen/go/visualisation"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/lib/grpcerr"
	"google.golang.org/grpc"
)

type VisualisationService interface {
	CreateAncestorsVisualisation(ctx context.Context, treeID string, rootPersonID string) (models.Visualisation, error)
	CreateDescendantsVisualisation(ctx context.Context, treeID string, rootPersonID string) (models.Visualisation, error)
	CreateAncestorsAndDescendantsVisualisation(ctx context.Context, treeID string, rootPersonID string) (models.Visualisation, error)
	CreateFullVisualisation(ctx context.Context, treeID string, rootPersonID string, includedPersonIDs []string) (models.Visualisation, error)
	ListTreeVisualisations(ctx context.Context, treeID string) ([]models.Visualisation, error)
	GetVisualisationByID(ctx context.Context, treeID string, visualisationID string) (models.Visualisation, []byte, error)
	DeleteVisualisationByID(ctx context.Context, treeID string, visualisationID string) error
	RenderCoordinatesForClient(ctx context.Context, treeID string, rootPersonID string, maxDepth int) ([]byte, error)
}

type Handler struct {
	visualisationpb.UnimplementedVisualisationServiceServer
	service VisualisationService
}

func New(service VisualisationService) *Handler {
	return &Handler{service: service}
}

func Register(gRPC *grpc.Server, service VisualisationService) {
	visualisationpb.RegisterVisualisationServiceServer(gRPC, New(service))
}

func (h *Handler) CreateAncestorsVisualisation(ctx context.Context, req *visualisationpb.CreateLineageVisualisationRequest) (*visualisationpb.CreateVisualisationResponse, error) {
	vis, err := h.service.CreateAncestorsVisualisation(ctx, req.GetTreeId(), req.GetRootPersonId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &visualisationpb.CreateVisualisationResponse{Visualisation: toProtoVisualisation(vis), Status: "queued"}, nil
}

func (h *Handler) CreateDescendantsVisualisation(ctx context.Context, req *visualisationpb.CreateLineageVisualisationRequest) (*visualisationpb.CreateVisualisationResponse, error) {
	vis, err := h.service.CreateDescendantsVisualisation(ctx, req.GetTreeId(), req.GetRootPersonId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &visualisationpb.CreateVisualisationResponse{Visualisation: toProtoVisualisation(vis), Status: "queued"}, nil
}

func (h *Handler) CreateAncestorsAndDescendantsVisualisation(ctx context.Context, req *visualisationpb.CreateLineageVisualisationRequest) (*visualisationpb.CreateVisualisationResponse, error) {
	vis, err := h.service.CreateAncestorsAndDescendantsVisualisation(ctx, req.GetTreeId(), req.GetRootPersonId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &visualisationpb.CreateVisualisationResponse{Visualisation: toProtoVisualisation(vis), Status: "queued"}, nil
}

func (h *Handler) CreateFullVisualisation(ctx context.Context, req *visualisationpb.CreateFullVisualisationRequest) (*visualisationpb.CreateVisualisationResponse, error) {
	vis, err := h.service.CreateFullVisualisation(ctx, req.GetTreeId(), req.GetRootPersonId(), req.GetIncludedPersonIds())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &visualisationpb.CreateVisualisationResponse{Visualisation: toProtoVisualisation(vis), Status: "queued"}, nil
}

func (h *Handler) ListTreeVisualisations(ctx context.Context, req *visualisationpb.ListTreeVisualisationsRequest) (*visualisationpb.ListTreeVisualisationsResponse, error) {
	items, err := h.service.ListTreeVisualisations(ctx, req.GetTreeId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	out := make([]*visualisationpb.Visualisation, 0, len(items))
	for _, item := range items {
		out = append(out, toProtoVisualisation(item))
	}

	return &visualisationpb.ListTreeVisualisationsResponse{Visualisations: out}, nil
}

func (h *Handler) GetVisualisationByID(ctx context.Context, req *visualisationpb.GetVisualisationByIDRequest) (*visualisationpb.GetVisualisationContentResponse, error) {
	vis, content, err := h.service.GetVisualisationByID(ctx, req.GetTreeId(), req.GetVisualisationId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &visualisationpb.GetVisualisationContentResponse{Visualisation: toProtoVisualisation(vis), Content: content}, nil
}

func (h *Handler) DeleteVisualisationByID(ctx context.Context, req *visualisationpb.DeleteVisualisationByIDRequest) (*visualisationpb.DeleteVisualisationByIDResponse, error) {
	if err := h.service.DeleteVisualisationByID(ctx, req.GetTreeId(), req.GetVisualisationId()); err != nil {
		return nil, grpcerr.Map(err)
	}

	return &visualisationpb.DeleteVisualisationByIDResponse{}, nil
}

func (h *Handler) RenderCoordinatesForClient(ctx context.Context, req *visualisationpb.RenderCoordinatesForClientRequest) (*visualisationpb.RenderCoordinatesForClientResponse, error) {
	coordBytes, err := h.service.RenderCoordinatesForClient(
		ctx,
		req.GetTreeId(),
		req.GetRootPersonId(),
		int(req.GetMaxDepth()),
	)
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &visualisationpb.RenderCoordinatesForClientResponse{CoordinatesJson: coordBytes}, nil
}

func toProtoVisualisation(vis models.Visualisation) *visualisationpb.Visualisation {
	included := make([]string, 0, len(vis.IncludedPersonIDs))
	for _, id := range vis.IncludedPersonIDs {
		included = append(included, id.String())
	}

	completedAtUnix := int64(0)
	if vis.CompletedAt != nil {
		completedAtUnix = vis.CompletedAt.Unix()
	}

	return &visualisationpb.Visualisation{
		Id:                vis.ID.String(),
		OwnerUserId:       int32(vis.OwnerUserID),
		TreeId:            vis.TreeID.String(),
		RootPersonId:      vis.RootPersonID.String(),
		IncludedPersonIds: included,
		Type:              toProtoType(vis.Type),
		Status:            toProtoStatus(vis.Status),
		FileName:          vis.FileName,
		MimeType:          vis.MIMEType,
		SizeBytes:         vis.SizeBytes,
		ErrorMessage:      vis.ErrorMessage,
		CreatedAtUnix:     vis.CreatedAt.Unix(),
		UpdatedAtUnix:     vis.UpdatedAt.Unix(),
		CompletedAtUnix:   completedAtUnix,
	}
}

func toProtoType(visType models.VisualisationType) visualisationpb.VisualisationType {
	switch visType {
	case models.VisualisationTypeAncestors:
		return visualisationpb.VisualisationType_VISUALISATION_TYPE_ANCESTORS
	case models.VisualisationTypeDescendants:
		return visualisationpb.VisualisationType_VISUALISATION_TYPE_DESCENDANTS
	case models.VisualisationTypeAncestorsAndDescendants:
		return visualisationpb.VisualisationType_VISUALISATION_TYPE_ANCESTORS_AND_DESCENDANTS
	case models.VisualisationTypeFull:
		return visualisationpb.VisualisationType_VISUALISATION_TYPE_FULL
	default:
		return visualisationpb.VisualisationType_VISUALISATION_TYPE_UNSPECIFIED
	}
}

func toProtoStatus(status models.VisualisationStatus) visualisationpb.VisualisationStatus {
	switch status {
	case models.VisualisationStatusPending:
		return visualisationpb.VisualisationStatus_VISUALISATION_STATUS_PENDING
	case models.VisualisationStatusProcessing:
		return visualisationpb.VisualisationStatus_VISUALISATION_STATUS_PROCESSING
	case models.VisualisationStatusReady:
		return visualisationpb.VisualisationStatus_VISUALISATION_STATUS_READY
	case models.VisualisationStatusFailed:
		return visualisationpb.VisualisationStatus_VISUALISATION_STATUS_FAILED
	default:
		return visualisationpb.VisualisationStatus_VISUALISATION_STATUS_UNSPECIFIED
	}
}
