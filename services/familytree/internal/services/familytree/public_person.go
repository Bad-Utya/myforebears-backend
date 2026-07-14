package familytree

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/storage"
	"github.com/google/uuid"
)

var (
	ErrPublicPersonNotFound = errors.New("public person not found")
	ErrInvalidAttachment    = errors.New("invalid public person attachment")
)

type PublicPersonAttachment string

const (
	AttachmentAsParent  PublicPersonAttachment = "AS_PARENT"
	AttachmentAsChild   PublicPersonAttachment = "AS_CHILD"
	AttachmentAsPartner PublicPersonAttachment = "AS_PARTNER"
)

func (s *Service) CreatePublicPerson(ctx context.Context, userID int) (models.PublicPerson, error) {
	return s.CreatePublicPersonSnapshot(ctx, userID, "", "", "", "", "", nil)
}

func (s *Service) CreatePublicPersonSnapshot(ctx context.Context, userID int, firstName, lastName, patronymic string, gender models.Gender, biography string, events []models.PublicPersonEvent) (models.PublicPerson, error) {
	if userID <= 0 {
		return models.PublicPerson{}, ErrInvalidUserID
	}
	if err := validatePublicEvents(events); err != nil {
		return models.PublicPerson{}, err
	}
	now := time.Now()
	p := models.PublicPerson{ID: uuid.New(), OwnerUserID: userID, FirstName: strings.TrimSpace(firstName), LastName: strings.TrimSpace(lastName), Patronymic: strings.TrimSpace(patronymic), Gender: gender, Biography: strings.TrimSpace(biography), CreatedAt: now, UpdatedAt: now}
	p.Events = normalizePublicEvents(p.ID, events)
	if err := s.publicStorage.CreatePublicPerson(ctx, p); err != nil {
		return models.PublicPerson{}, fmt.Errorf("create public person: %w", err)
	}
	return p, nil
}

func (s *Service) GetPublicPerson(ctx context.Context, id string) (models.PublicPerson, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return models.PublicPerson{}, ErrInvalidPersonID
	}
	p, err := s.publicStorage.GetPublicPerson(ctx, parsed)
	if errors.Is(err, storage.ErrPublicPersonNotFound) {
		return models.PublicPerson{}, ErrPublicPersonNotFound
	}
	return p, err
}

func (s *Service) ListRandomPublicPersons(ctx context.Context, limit int) ([]models.PublicPerson, error) {
	if limit <= 0 || limit > 100 {
		return nil, ErrInvalidLimit
	}
	return s.publicStorage.ListRandomPublicPersons(ctx, limit)
}

func (s *Service) ListPublicPersonsByOwner(ctx context.Context, ownerUserID, limit int) ([]models.PublicPerson, error) {
	if ownerUserID <= 0 {
		return nil, ErrInvalidUserID
	}
	if limit <= 0 || limit > 100 {
		return nil, ErrInvalidLimit
	}
	return s.publicStorage.ListPublicPersonsByOwner(ctx, ownerUserID, limit)
}

func (s *Service) SearchPublicPersons(ctx context.Context, query string, tagCodes []string, limit int) ([]models.PublicPerson, error) {
	query = strings.TrimSpace(query)
	tagCodes = normalizeTagCodes(tagCodes)
	if query == "" && len(tagCodes) == 0 {
		return nil, ErrInvalidQuery
	}
	if limit <= 0 || limit > 100 {
		return nil, ErrInvalidLimit
	}
	items, err := s.tagStorage.SearchPublicPersonsByTags(ctx, query, tagCodes, limit)
	if errors.Is(err, storage.ErrUnknownTag) {
		return nil, ErrUnknownTag
	}
	return items, err
}

func (s *Service) UpdatePublicPerson(ctx context.Context, userID int, person models.PublicPerson) (models.PublicPerson, error) {
	if err := validatePublicEvents(person.Events); err != nil {
		return models.PublicPerson{}, err
	}
	existing, err := s.GetPublicPerson(ctx, person.ID.String())
	if err != nil {
		return models.PublicPerson{}, err
	}
	if existing.OwnerUserID != userID {
		return models.PublicPerson{}, ErrForbidden
	}
	existing.FirstName = strings.TrimSpace(person.FirstName)
	existing.LastName = strings.TrimSpace(person.LastName)
	existing.Patronymic = strings.TrimSpace(person.Patronymic)
	existing.Gender = person.Gender
	existing.Biography = strings.TrimSpace(person.Biography)
	existing.UpdatedAt = time.Now()
	existing.Events = normalizePublicEvents(existing.ID, person.Events)
	if err := s.publicStorage.UpdatePublicPerson(ctx, existing); err != nil {
		return models.PublicPerson{}, err
	}
	return existing, nil
}

func validatePublicEvents(events []models.PublicPersonEvent) error {
	for _, event := range events {
		if strings.TrimSpace(event.EventTypeName) == "" {
			return ErrInvalidName
		}
		if !event.DateUnknown && strings.TrimSpace(event.DateISO) == "" {
			return ErrInvalidQuery
		}
	}
	return nil
}

func (s *Service) SetPublicPersonAvatarPhoto(ctx context.Context, userID int, personID, photoID string) (models.PublicPerson, error) {
	p, err := s.GetPublicPerson(ctx, personID)
	if err != nil {
		return p, err
	}
	if p.OwnerUserID != userID {
		return p, ErrForbidden
	}
	var avatar *uuid.UUID
	if strings.TrimSpace(photoID) != "" {
		id, err := uuid.Parse(photoID)
		if err != nil {
			return p, ErrInvalidPersonID
		}
		avatar = &id
	}
	if err := s.publicStorage.SetPublicPersonAvatarPhoto(ctx, p.ID, avatar); err != nil {
		return p, err
	}
	p.AvatarPhotoID = avatar
	return p, nil
}

func (s *Service) DeletePublicPerson(ctx context.Context, userID int, personID string) error {
	p, err := s.GetPublicPerson(ctx, personID)
	if err != nil {
		return err
	}
	if p.OwnerUserID != userID {
		return ErrForbidden
	}
	if err := s.publicStorage.DeletePublicPerson(ctx, p.ID); errors.Is(err, storage.ErrPublicPersonNotFound) {
		return ErrPublicPersonNotFound
	} else {
		return err
	}
}

func (s *Service) ImportPublicPersonIntoTree(ctx context.Context, userID int, publicPersonID, treeID, attachToID string, attachment PublicPersonAttachment) (models.Person, error) {
	if attachment != AttachmentAsParent && attachment != AttachmentAsChild && attachment != AttachmentAsPartner {
		return models.Person{}, ErrInvalidAttachment
	}
	pub, err := s.GetPublicPerson(ctx, publicPersonID)
	if err != nil {
		return models.Person{}, err
	}
	if !isValidGender(pub.Gender) {
		return models.Person{}, ErrInvalidGender
	}
	tree, err := s.GetTree(ctx, treeID)
	if err != nil {
		return models.Person{}, err
	}
	if tree.CreatorID != userID {
		return models.Person{}, ErrForbidden
	}
	anchor, err := s.GetPerson(ctx, treeID, attachToID)
	if err != nil {
		return models.Person{}, err
	}
	p, err := s.createPersonRecord(ctx, tree.ID, pub.Gender, pub.FirstName, pub.LastName, pub.Patronymic)
	if err != nil {
		return models.Person{}, err
	}
	p.Biography = pub.Biography
	if err := s.personStorage.UpdatePerson(ctx, p); err != nil {
		return models.Person{}, err
	}
	var from, to uuid.UUID
	var rt models.RelationshipType
	switch attachment {
	case AttachmentAsParent:
		from, to, rt = p.ID, anchor.ID, models.RelationshipParentChild
	case AttachmentAsChild:
		from, to, rt = anchor.ID, p.ID, models.RelationshipParentChild
	case AttachmentAsPartner:
		from, to, rt = anchor.ID, p.ID, models.RelationshipPartner
	default:
		return models.Person{}, ErrInvalidAttachment
	}
	if _, err := s.ensureRelationship(ctx, from, to, rt); err != nil {
		return models.Person{}, err
	}
	return p, nil
}

func (s *Service) CreateTreeFromPublicPerson(ctx context.Context, userID int, publicPersonID, treeName string) (models.Tree, models.Person, error) {
	pub, err := s.GetPublicPerson(ctx, publicPersonID)
	if err != nil {
		return models.Tree{}, models.Person{}, err
	}
	if !isValidGender(pub.Gender) {
		return models.Tree{}, models.Person{}, ErrInvalidGender
	}
	tree, root, err := s.CreateTree(ctx, userID, "", treeName)
	if err != nil {
		return tree, root, err
	}
	root.FirstName, root.LastName, root.Patronymic, root.Gender, root.Biography = pub.FirstName, pub.LastName, pub.Patronymic, pub.Gender, pub.Biography
	if err := s.personStorage.UpdatePerson(ctx, root); err != nil {
		return models.Tree{}, models.Person{}, err
	}
	return tree, root, nil
}

func normalizePublicEvents(personID uuid.UUID, events []models.PublicPersonEvent) []models.PublicPersonEvent {
	result := make([]models.PublicPersonEvent, 0, len(events))
	for _, event := range events {
		if event.ID == uuid.Nil {
			event.ID = uuid.New()
		}
		event.PublicPersonID = personID
		if event.DatePrecision == "" {
			event.DatePrecision = "DAY"
		}
		if event.DateBound == "" {
			event.DateBound = "EXACT"
		}
		event.EventTypeName = strings.TrimSpace(event.EventTypeName)
		result = append(result, event)
	}
	return result
}
