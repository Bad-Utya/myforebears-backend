package familytree

import (
	"context"

	familytreepb "github.com/Bad-Utya/myforebears-backend/gen/go/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/lib/grpcerr"
)

func (h *Handler) AddRelationship(ctx context.Context, req *familytreepb.AddRelationshipRequest) (*familytreepb.AddRelationshipResponse, error) {
	err := h.service.AddRelationship(ctx, req.GetPersonIdFrom(), req.GetPersonIdTo(), toModelRelationshipType(req.GetType()))
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.AddRelationshipResponse{}, nil
}

func (h *Handler) RemoveRelationship(ctx context.Context, req *familytreepb.RemoveRelationshipRequest) (*familytreepb.RemoveRelationshipResponse, error) {
	err := h.service.RemoveRelationship(ctx, req.GetPersonIdFrom(), req.GetPersonIdTo(), toModelRelationshipType(req.GetType()))
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.RemoveRelationshipResponse{}, nil
}

func (h *Handler) GetRelatives(ctx context.Context, req *familytreepb.GetRelativesRequest) (*familytreepb.GetRelativesResponse, error) {
	relatives, err := h.service.GetRelatives(ctx, req.GetPersonId())
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
	persons, relationships, err := h.service.GetTreeForUser(ctx, int(req.GetRequestUserId()), req.GetTreeId())
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

	return &familytreepb.GetTreeResponse{Persons: protoPersons, Relationships: protoRelationships}, nil
}

func toModelRelationshipType(relType familytreepb.RelationshipType) models.RelationshipType {
	switch relType {
	case familytreepb.RelationshipType_RELATIONSHIP_PARENT_CHILD:
		return models.RelationshipParentChild
	case familytreepb.RelationshipType_RELATIONSHIP_PARTNER:
		return models.RelationshipPartner
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
	default:
		return familytreepb.RelationshipType_RELATIONSHIP_TYPE_UNSPECIFIED
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
