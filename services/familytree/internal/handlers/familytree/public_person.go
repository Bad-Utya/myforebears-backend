package familytree

import (
	"context"

	familytreepb "github.com/Bad-Utya/myforebears-backend/gen/go/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/lib/grpcerr"
	personsvc "github.com/Bad-Utya/myforebears-backend/services/familytree/internal/services/familytree"
	"github.com/google/uuid"
)

func (h *Handler) CreatePublicPerson(ctx context.Context, req *familytreepb.CreatePublicPersonRequest) (*familytreepb.PublicPersonResponse, error) {
	p, err := h.service.CreatePublicPerson(ctx, int(req.GetRequestUserId()))
	if err != nil {
		return nil, grpcerr.Map(err)
	}
	return &familytreepb.PublicPersonResponse{Person: toProtoPublicPerson(p)}, nil
}

func (h *Handler) CreatePublicPersonSnapshot(ctx context.Context, req *familytreepb.CreatePublicPersonSnapshotRequest) (*familytreepb.PublicPersonResponse, error) {
	p, err := h.service.CreatePublicPersonSnapshot(ctx, int(req.GetRequestUserId()), req.GetFirstName(), req.GetLastName(), req.GetPatronymic(), toModelGender(req.GetGender()), req.GetBiography(), toModelPublicEvents(req.GetEvents()))
	if err != nil {
		return nil, grpcerr.Map(err)
	}
	return &familytreepb.PublicPersonResponse{Person: toProtoPublicPerson(p)}, nil
}

func (h *Handler) GetPublicPerson(ctx context.Context, req *familytreepb.GetPublicPersonRequest) (*familytreepb.PublicPersonResponse, error) {
	p, err := h.service.GetPublicPerson(ctx, req.GetPublicPersonId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}
	return &familytreepb.PublicPersonResponse{Person: toProtoPublicPerson(p)}, nil
}

func (h *Handler) ListRandomPublicPersons(ctx context.Context, req *familytreepb.ListRandomPublicPersonsRequest) (*familytreepb.PublicPersonsResponse, error) {
	items, err := h.service.ListRandomPublicPersons(ctx, int(req.GetLimit()))
	if err != nil {
		return nil, grpcerr.Map(err)
	}
	return toProtoPublicPersons(items), nil
}

func (h *Handler) ListPublicPersonsByOwner(ctx context.Context, req *familytreepb.ListPublicPersonsByOwnerRequest) (*familytreepb.PublicPersonsResponse, error) {
	items, err := h.service.ListPublicPersonsByOwner(ctx, int(req.GetOwnerUserId()), int(req.GetLimit()))
	if err != nil {
		return nil, grpcerr.Map(err)
	}
	return toProtoPublicPersons(items), nil
}

func (h *Handler) SearchPublicPersons(ctx context.Context, req *familytreepb.SearchPublicPersonsRequest) (*familytreepb.PublicPersonsResponse, error) {
	items, err := h.service.SearchPublicPersons(ctx, req.GetQuery(), req.GetTagCodes(), int(req.GetLimit()))
	if err != nil {
		return nil, grpcerr.Map(err)
	}
	return toProtoPublicPersons(items), nil
}

func (h *Handler) SetPublicPersonTags(ctx context.Context, req *familytreepb.SetPublicPersonTagsRequest) (*familytreepb.PublicPersonResponse, error) {
	person, err := h.service.SetPublicPersonTags(ctx, int(req.GetRequestUserId()), req.GetPublicPersonId(), req.GetTagCodes())
	if err != nil {
		return nil, grpcerr.Map(err)
	}
	return &familytreepb.PublicPersonResponse{Person: toProtoPublicPerson(person)}, nil
}

func (h *Handler) UpdatePublicPerson(ctx context.Context, req *familytreepb.UpdatePublicPersonRequest) (*familytreepb.PublicPersonResponse, error) {
	id, err := uuid.Parse(req.GetPublicPersonId())
	if err != nil {
		return nil, grpcerr.Map(personsvc.ErrInvalidPersonID)
	}
	p, err := h.service.UpdatePublicPerson(ctx, int(req.GetRequestUserId()), models.PublicPerson{ID: id, FirstName: req.GetFirstName(), LastName: req.GetLastName(), Patronymic: req.GetPatronymic(), Gender: toModelGender(req.GetGender()), Biography: req.GetBiography(), Events: toModelPublicEvents(req.GetEvents())})
	if err != nil {
		return nil, grpcerr.Map(err)
	}
	return &familytreepb.PublicPersonResponse{Person: toProtoPublicPerson(p)}, nil
}

func (h *Handler) SetPublicPersonAvatarPhoto(ctx context.Context, req *familytreepb.SetPublicPersonAvatarPhotoRequest) (*familytreepb.PublicPersonResponse, error) {
	p, err := h.service.SetPublicPersonAvatarPhoto(ctx, int(req.GetRequestUserId()), req.GetPublicPersonId(), req.GetAvatarPhotoId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}
	return &familytreepb.PublicPersonResponse{Person: toProtoPublicPerson(p)}, nil
}

func (h *Handler) DeletePublicPerson(ctx context.Context, req *familytreepb.DeletePublicPersonRequest) (*familytreepb.DeletePublicPersonResponse, error) {
	if err := h.service.DeletePublicPerson(ctx, int(req.GetRequestUserId()), req.GetPublicPersonId()); err != nil {
		return nil, grpcerr.Map(err)
	}
	return &familytreepb.DeletePublicPersonResponse{}, nil
}

func (h *Handler) ImportPublicPersonIntoTree(ctx context.Context, req *familytreepb.ImportPublicPersonIntoTreeRequest) (*familytreepb.ImportPublicPersonIntoTreeResponse, error) {
	p, err := h.service.ImportPublicPersonIntoTree(ctx, int(req.GetRequestUserId()), req.GetPublicPersonId(), req.GetTreeId(), req.GetAttachToPersonId(), toAttachment(req.GetAttachment()))
	if err != nil {
		return nil, grpcerr.Map(err)
	}
	return &familytreepb.ImportPublicPersonIntoTreeResponse{Person: toProtoPerson(p)}, nil
}

func (h *Handler) CreateTreeFromPublicPerson(ctx context.Context, req *familytreepb.CreateTreeFromPublicPersonRequest) (*familytreepb.CreateTreeFromPublicPersonResponse, error) {
	tree, p, err := h.service.CreateTreeFromPublicPerson(ctx, int(req.GetRequestUserId()), req.GetPublicPersonId(), req.GetTreeName())
	if err != nil {
		return nil, grpcerr.Map(err)
	}
	return &familytreepb.CreateTreeFromPublicPersonResponse{Tree: toProtoTree(tree), Person: toProtoPerson(p)}, nil
}

func toModelPublicEvents(items []*familytreepb.PublicPersonEventInput) []models.PublicPersonEvent {
	result := make([]models.PublicPersonEvent, 0, len(items))
	for _, item := range items {
		e := models.PublicPersonEvent{EventTypeName: item.GetEventTypeName(), DateISO: item.GetDateIso(), DatePrecision: item.GetDatePrecision(), DateBound: item.GetDateBound(), DateUnknown: item.GetDateUnknown()}
		if id, err := uuid.Parse(item.GetId()); err == nil {
			e.ID = id
		}
		if id, err := uuid.Parse(item.GetSourceEventId()); err == nil {
			e.SourceEventID = &id
		}
		if id, err := uuid.Parse(item.GetEventTypeId()); err == nil {
			e.EventTypeID = &id
		}
		result = append(result, e)
	}
	return result
}

func toProtoPublicPerson(p models.PublicPerson) *familytreepb.PublicPerson {
	events := make([]*familytreepb.PublicPersonEvent, 0, len(p.Events))
	for _, e := range p.Events {
		pe := &familytreepb.PublicPersonEvent{Id: e.ID.String(), PublicPersonId: p.ID.String(), EventTypeName: e.EventTypeName, DateIso: e.DateISO, DatePrecision: e.DatePrecision, DateBound: e.DateBound, DateUnknown: e.DateUnknown}
		if e.SourceEventID != nil {
			pe.SourceEventId = e.SourceEventID.String()
		}
		if e.EventTypeID != nil {
			pe.EventTypeId = e.EventTypeID.String()
		}
		events = append(events, pe)
	}
	tags := make([]*familytreepb.Tag, 0, len(p.Tags))
	for _, tag := range p.Tags {
		tags = append(tags, toProtoTag(tag))
	}
	return &familytreepb.PublicPerson{Id: p.ID.String(), OwnerUserId: int32(p.OwnerUserID), FirstName: p.FirstName, LastName: p.LastName, Patronymic: p.Patronymic, Gender: toProtoGender(p.Gender), Biography: p.Biography, AvatarPhotoId: toProtoAvatarPhotoID(p.AvatarPhotoID), CreatedAtUnix: p.CreatedAt.Unix(), UpdatedAtUnix: p.UpdatedAt.Unix(), Events: events, Tags: tags, SimilarityScore: p.SimilarityScore}
}

func toProtoPublicPersons(items []models.PublicPerson) *familytreepb.PublicPersonsResponse {
	out := make([]*familytreepb.PublicPerson, 0, len(items))
	for _, p := range items {
		out = append(out, toProtoPublicPerson(p))
	}
	return &familytreepb.PublicPersonsResponse{Persons: out}
}

func toAttachment(v familytreepb.PublicPersonAttachment) personsvc.PublicPersonAttachment {
	switch v {
	case familytreepb.PublicPersonAttachment_PUBLIC_PERSON_ATTACHMENT_AS_PARENT:
		return personsvc.AttachmentAsParent
	case familytreepb.PublicPersonAttachment_PUBLIC_PERSON_ATTACHMENT_AS_CHILD:
		return personsvc.AttachmentAsChild
	case familytreepb.PublicPersonAttachment_PUBLIC_PERSON_ATTACHMENT_AS_PARTNER:
		return personsvc.AttachmentAsPartner
	default:
		return ""
	}
}
