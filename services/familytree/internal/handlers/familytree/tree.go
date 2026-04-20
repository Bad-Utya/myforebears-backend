package familytree

import (
	"context"

	familytreepb "github.com/Bad-Utya/myforebears-backend/gen/go/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/lib/grpcerr"
	personsvc "github.com/Bad-Utya/myforebears-backend/services/familytree/internal/services/familytree"
)

func (h *Handler) CreateTree(ctx context.Context, req *familytreepb.CreateTreeRequest) (*familytreepb.CreateTreeResponse, error) {
	tree, root, err := h.service.CreateTree(ctx, int(req.GetRequestUserId()))
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.CreateTreeResponse{
		Tree:       toProtoTree(tree),
		RootPerson: toProtoPerson(root),
	}, nil
}

func (h *Handler) ListTreesByCreator(ctx context.Context, req *familytreepb.ListTreesByCreatorRequest) (*familytreepb.ListTreesByCreatorResponse, error) {
	trees, err := h.service.ListTreesByCreator(ctx, int(req.GetRequestUserId()))
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	out := make([]*familytreepb.Tree, 0, len(trees))
	for _, tree := range trees {
		out = append(out, toProtoTree(tree))
	}

	return &familytreepb.ListTreesByCreatorResponse{Trees: out}, nil
}

func (h *Handler) AddParent(ctx context.Context, req *familytreepb.AddParentRequest) (*familytreepb.AddParentResponse, error) {
	parent, autoParent, err := h.service.AddParent(
		ctx,
		int(req.GetRequestUserId()),
		req.GetTreeId(),
		req.GetChildId(),
		toModelParentRole(req.GetRole()),
		req.GetFirstName(),
		req.GetLastName(),
		req.GetPatronymic(),
	)
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	resp := &familytreepb.AddParentResponse{Parent: toProtoPerson(parent)}
	if autoParent != nil {
		resp.AutoCreatedSecondParent = toProtoPerson(*autoParent)
	}

	return resp, nil
}

func (h *Handler) AddChild(ctx context.Context, req *familytreepb.AddChildRequest) (*familytreepb.AddChildResponse, error) {
	child, autoParent, err := h.service.AddChild(
		ctx,
		int(req.GetRequestUserId()),
		req.GetTreeId(),
		req.GetParent1Id(),
		req.GetParent2Id(),
		req.GetFirstName(),
		req.GetLastName(),
		req.GetPatronymic(),
		toModelGender(req.GetGender()),
	)
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	resp := &familytreepb.AddChildResponse{Child: toProtoPerson(child)}
	if autoParent != nil {
		resp.AutoCreatedParent = toProtoPerson(*autoParent)
	}

	return resp, nil
}

func (h *Handler) AddPartner(ctx context.Context, req *familytreepb.AddPartnerRequest) (*familytreepb.AddPartnerResponse, error) {
	partner, err := h.service.AddPartner(
		ctx,
		int(req.GetRequestUserId()),
		req.GetTreeId(),
		req.GetPersonId(),
		req.GetFirstName(),
		req.GetLastName(),
		req.GetPatronymic(),
	)
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.AddPartnerResponse{Partner: toProtoPerson(partner)}, nil
}

func (h *Handler) ValidatePersonsInTree(ctx context.Context, req *familytreepb.ValidatePersonsInTreeRequest) (*familytreepb.ValidatePersonsInTreeResponse, error) {
	err := h.service.ValidatePersonsInTree(ctx, int(req.GetRequestUserId()), req.GetTreeId(), req.GetPersonIds())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.ValidatePersonsInTreeResponse{}, nil
}
func (h *Handler) UpdatePersonName(ctx context.Context, req *familytreepb.UpdatePersonNameRequest) (*familytreepb.UpdatePersonNameResponse, error) {
	person, err := h.service.UpdatePersonName(
		ctx,
		int(req.GetRequestUserId()),
		req.GetTreeId(),
		req.GetPersonId(),
		req.GetFirstName(),
		req.GetLastName(),
		req.GetPatronymic(),
	)
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.UpdatePersonNameResponse{Person: toProtoPerson(person)}, nil
}

func (h *Handler) DeletePersonInTree(ctx context.Context, req *familytreepb.DeletePersonInTreeRequest) (*familytreepb.DeletePersonInTreeResponse, error) {
	if err := h.service.DeletePersonInTree(ctx, int(req.GetRequestUserId()), req.GetTreeId(), req.GetPersonId()); err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.DeletePersonInTreeResponse{}, nil
}

func toModelParentRole(role familytreepb.ParentRole) personsvc.ParentRole {
	switch role {
	case familytreepb.ParentRole_PARENT_ROLE_FATHER:
		return personsvc.ParentRoleFather
	case familytreepb.ParentRole_PARENT_ROLE_MOTHER:
		return personsvc.ParentRoleMother
	default:
		return ""
	}
}

func toProtoTree(tree models.Tree) *familytreepb.Tree {
	return &familytreepb.Tree{
		Id:            tree.ID.String(),
		CreatorId:     int32(tree.CreatorID),
		CreatedAtUnix: tree.CreatedAt.Unix(),
	}
}
