package familytree

import (
	"context"

	familytreepb "github.com/Bad-Utya/myforebears-backend/gen/go/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/lib/grpcerr"
)

func (h *Handler) AddTreeAccessEmail(ctx context.Context, req *familytreepb.AddTreeAccessEmailRequest) (*familytreepb.AddTreeAccessEmailResponse, error) {
	err := h.service.AddTreeAccessEmail(ctx, int(req.GetRequestUserId()), req.GetTreeId(), req.GetEmail())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.AddTreeAccessEmailResponse{}, nil
}

func (h *Handler) ListTreeAccessEmails(ctx context.Context, req *familytreepb.ListTreeAccessEmailsRequest) (*familytreepb.ListTreeAccessEmailsResponse, error) {
	emails, err := h.service.ListTreeAccessEmails(ctx, int(req.GetRequestUserId()), req.GetTreeId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.ListTreeAccessEmailsResponse{Emails: emails}, nil
}

func (h *Handler) DeleteTreeAccessEmail(ctx context.Context, req *familytreepb.DeleteTreeAccessEmailRequest) (*familytreepb.DeleteTreeAccessEmailResponse, error) {
	err := h.service.DeleteTreeAccessEmail(ctx, int(req.GetRequestUserId()), req.GetTreeId(), req.GetEmail())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.DeleteTreeAccessEmailResponse{}, nil
}
