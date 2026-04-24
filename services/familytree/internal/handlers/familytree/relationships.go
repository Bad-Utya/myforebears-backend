package familytree

import (
	"context"

	familytreepb "github.com/Bad-Utya/myforebears-backend/gen/go/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/lib/grpcerr"
)

func (h *Handler) AddRelationship(ctx context.Context, req *familytreepb.AddRelationshipRequest) (*familytreepb.AddRelationshipResponse, error) {
	err := h.service.AddRelationship(ctx, req.GetTreeId(), req.GetPersonIdFrom(), req.GetPersonIdTo(), toModelRelationshipType(req.GetType()))
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.AddRelationshipResponse{}, nil
}

func (h *Handler) RemoveRelationship(ctx context.Context, req *familytreepb.RemoveRelationshipRequest) (*familytreepb.RemoveRelationshipResponse, error) {
	err := h.service.RemoveRelationship(ctx, req.GetTreeId(), req.GetPersonIdFrom(), req.GetPersonIdTo(), toModelRelationshipType(req.GetType()))
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.RemoveRelationshipResponse{}, nil
}

func (h *Handler) GetRelatives(ctx context.Context, req *familytreepb.GetRelativesRequest) (*familytreepb.GetRelativesResponse, error) {
	relatives, err := h.service.GetRelatives(ctx, req.GetTreeId(), req.GetPersonId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	out := make([]*familytreepb.Relative, 0, len(relatives))
	for _, relative := range relatives {
		out = append(out, &familytreepb.Relative{
			Person:           toProtoPerson(relative.Person),
			RelationshipType: toProtoRelationshipType(relative.RelationshipType),
			Direction:        toProtoDirection(relative.Direction),
		})
	}

	return &familytreepb.GetRelativesResponse{Relatives: out}, nil
}

func (h *Handler) GetTree(ctx context.Context, req *familytreepb.GetTreeRequest) (*familytreepb.GetTreeResponse, error) {
	tree, err := h.service.GetTree(ctx, req.GetTreeId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.GetTreeResponse{Tree: toProtoTree(tree)}, nil
}

func (h *Handler) GetTreeContent(ctx context.Context, req *familytreepb.GetTreeContentRequest) (*familytreepb.GetTreeContentResponse, error) {
	persons, relationships, err := h.service.GetTreeContent(ctx, req.GetTreeId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	protoPersons := make([]*familytreepb.Person, 0, len(persons))
	for _, person := range persons {
		protoPersons = append(protoPersons, toProtoPerson(person))
	}

	protoRelationships := make([]*familytreepb.Relationship, 0, len(relationships))
	for _, rel := range relationships {
		protoRelationships = append(protoRelationships, &familytreepb.Relationship{
			PersonIdFrom: rel.PersonIDFrom.String(),
			PersonIdTo:   rel.PersonIDTo.String(),
			Type:         toProtoRelationshipType(rel.Type),
		})
	}

	return &familytreepb.GetTreeContentResponse{Persons: protoPersons, Relationships: protoRelationships}, nil
}

func (h *Handler) UpdatePartnerRelationshipStatus(ctx context.Context, req *familytreepb.UpdatePartnerRelationshipStatusRequest) (*familytreepb.UpdatePartnerRelationshipStatusResponse, error) {
	err := h.service.UpdatePartnerRelationshipStatus(
		ctx,
		req.GetTreeId(),
		req.GetPersonId1(),
		req.GetPersonId2(),
		toModelPartnerRelationshipStatus(req.GetStatus()),
	)
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.UpdatePartnerRelationshipStatusResponse{}, nil
}

func (h *Handler) UpdateTreeSettings(ctx context.Context, req *familytreepb.UpdateTreeSettingsRequest) (*familytreepb.UpdateTreeSettingsResponse, error) {
	tree, err := h.service.UpdateTreeSettings(
		ctx,
		req.GetTreeId(),
		req.GetIsViewRestricted(),
		req.GetIsPublicOnMainPage(),
	)
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.UpdateTreeSettingsResponse{Tree: toProtoTree(tree)}, nil
}

func toModelRelationshipType(relType familytreepb.RelationshipType) models.RelationshipType {
	switch relType {
	case familytreepb.RelationshipType_RELATIONSHIP_PARENT_CHILD:
		return models.RelationshipParentChild
	case familytreepb.RelationshipType_RELATIONSHIP_PARTNER:
		return models.RelationshipPartner
	case familytreepb.RelationshipType_RELATIONSHIP_PARTNER_UNMARRIED:
		return models.RelationshipPartnerUnmarried
	case familytreepb.RelationshipType_RELATIONSHIP_PARTNER_MARRIED:
		return models.RelationshipPartnerMarried
	case familytreepb.RelationshipType_RELATIONSHIP_PARTNER_DIVORCED:
		return models.RelationshipPartnerDivorced
	default:
		return ""
	}
}

func toProtoRelationshipType(relType models.RelationshipType) familytreepb.RelationshipType {
	switch relType {
	case models.RelationshipParentChild:
		return familytreepb.RelationshipType_RELATIONSHIP_PARENT_CHILD
	case models.RelationshipPartner:
		return familytreepb.RelationshipType_RELATIONSHIP_PARTNER
	case models.RelationshipPartnerUnmarried:
		return familytreepb.RelationshipType_RELATIONSHIP_PARTNER_UNMARRIED
	case models.RelationshipPartnerMarried:
		return familytreepb.RelationshipType_RELATIONSHIP_PARTNER_MARRIED
	case models.RelationshipPartnerDivorced:
		return familytreepb.RelationshipType_RELATIONSHIP_PARTNER_DIVORCED
	default:
		return familytreepb.RelationshipType_RELATIONSHIP_TYPE_UNSPECIFIED
	}
}

func toModelPartnerRelationshipStatus(status familytreepb.PartnerRelationshipStatus) models.PartnerRelationshipStatus {
	switch status {
	case familytreepb.PartnerRelationshipStatus_PARTNER_RELATIONSHIP_STATUS_UNMARRIED:
		return models.PartnerRelationshipStatusUnmarried
	case familytreepb.PartnerRelationshipStatus_PARTNER_RELATIONSHIP_STATUS_MARRIED:
		return models.PartnerRelationshipStatusMarried
	case familytreepb.PartnerRelationshipStatus_PARTNER_RELATIONSHIP_STATUS_DIVORCED:
		return models.PartnerRelationshipStatusDivorced
	default:
		return ""
	}
}

func toProtoDirection(direction models.RelationDirection) familytreepb.RelationDirection {
	switch direction {
	case models.DirectionOutgoing:
		return familytreepb.RelationDirection_RELATION_DIRECTION_OUTGOING
	case models.DirectionIncoming:
		return familytreepb.RelationDirection_RELATION_DIRECTION_INCOMING
	default:
		return familytreepb.RelationDirection_RELATION_DIRECTION_UNSPECIFIED
	}
}
