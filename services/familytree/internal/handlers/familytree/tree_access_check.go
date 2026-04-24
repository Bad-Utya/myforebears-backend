package familytree

import (
	"context"

	familytreepb "github.com/Bad-Utya/myforebears-backend/gen/go/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/lib/grpcerr"
)

func (h *Handler) GetTreeAccessInfo(ctx context.Context, req *familytreepb.GetTreeAccessInfoRequest) (*familytreepb.GetTreeAccessInfoResponse, error) {
	tree, err := h.service.GetTreeAccessInfo(ctx, req.GetTreeId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.GetTreeAccessInfoResponse{Tree: toProtoTree(tree)}, nil
}

func (h *Handler) IsTreeAccessEmailAllowed(ctx context.Context, req *familytreepb.IsTreeAccessEmailAllowedRequest) (*familytreepb.IsTreeAccessEmailAllowedResponse, error) {
	allowed, err := h.service.IsTreeAccessEmailAllowed(ctx, req.GetTreeId(), req.GetEmail())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.IsTreeAccessEmailAllowedResponse{Allowed: allowed}, nil
}
