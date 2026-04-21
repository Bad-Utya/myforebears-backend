package familytree

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	eventspb "github.com/Bad-Utya/myforebears-backend/gen/go/events"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/storage"
	"github.com/google/uuid"
)

var (
	ErrInvalidPersonID     = errors.New("invalid person id")
	ErrInvalidTreeID       = errors.New("invalid tree id")
	ErrInvalidName         = errors.New("first name and last name are required")
	ErrInvalidGender       = errors.New("invalid gender")
	ErrInvalidUserID       = errors.New("invalid user id")
	ErrForbidden           = errors.New("forbidden")
	ErrInvalidParentRole   = errors.New("invalid parent role")
	ErrParentExists        = errors.New("parent with this role already exists")
	ErrParentLimitReached  = errors.New("child already has two parents")
	ErrAtLeastOneParent    = errors.New("at least one parent id must be provided")
	ErrTreeMismatch        = errors.New("person does not belong to tree")
	ErrUnknownPersonGender = errors.New("person gender is unknown")
	ErrPersonNotFound      = errors.New("person not found")
	ErrTreeNotFound        = errors.New("tree not found")
	ErrDeleteNotAllowed    = errors.New("delete is not allowed for current relationship set")
	ErrCannotDeleteLast    = errors.New("cannot delete the only person in tree")
	ErrSelfRelationship    = errors.New("self relationship is not allowed")
	ErrRelationshipExists  = errors.New("relationship already exists")
	ErrRelationshipMissing = errors.New("relationship not found")
	ErrPersonNotInSameTree = errors.New("persons must belong to the same tree")
	ErrInvalidRelationType = errors.New("invalid relationship type")
)

type Service struct {
	log             *slog.Logger
	personStorage   storage.PersonStorage
	relationStorage storage.RelationshipStorage
	eventsClient    eventsClient
}

type eventsClient interface {
	CreateEvent(ctx context.Context, req *eventspb.CreateEventRequest) (*eventspb.CreateEventResponse, error)
}

func New(log *slog.Logger, personStorage storage.PersonStorage, relationStorage storage.RelationshipStorage, eventsClient eventsClient) *Service {
	return &Service{
		log:             log,
		personStorage:   personStorage,
		relationStorage: relationStorage,
		eventsClient:    eventsClient,
	}
}

func (s *Service) CreatePerson(
	ctx context.Context,
	treeID string,
	firstName string,
	lastName string,
	patronymic string,
	gender models.Gender,
) (models.Person, error) {
	const op = "service.familytree.CreatePerson"

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	if firstName == "" || lastName == "" {
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidName)
	}

	if !isValidGender(gender) {
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidGender)
	}

	tree, err := s.personStorage.GetTree(ctx, parsedTreeID)
	if err != nil {
		if errors.Is(err, storage.ErrTreeNotFound) {
			return models.Person{}, fmt.Errorf("%s: %w", op, ErrTreeNotFound)
		}
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	person, err := s.createPersonRecord(ctx, tree.CreatorID, parsedTreeID, gender, firstName, lastName, patronymic)
	if err != nil {
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	return person, nil
}

func (s *Service) GetPerson(ctx context.Context, personID string) (models.Person, error) {
	const op = "service.familytree.GetPerson"

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	person, err := s.personStorage.GetPerson(ctx, parsedPersonID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			return models.Person{}, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	return person, nil
}

func (s *Service) UpdatePerson(
	ctx context.Context,
	personID string,
	firstName string,
	lastName string,
	patronymic string,
	gender models.Gender,
) (models.Person, error) {
	const op = "service.familytree.UpdatePerson"

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	if firstName == "" || lastName == "" {
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidName)
	}

	if !isValidGender(gender) {
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidGender)
	}

	existing, err := s.personStorage.GetPerson(ctx, parsedPersonID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			return models.Person{}, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	existing.FirstName = firstName
	existing.LastName = lastName
	existing.Patronymic = patronymic
	existing.Gender = gender

	if err := s.personStorage.UpdatePerson(ctx, existing); err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			return models.Person{}, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	return existing, nil
}

func (s *Service) DeletePerson(ctx context.Context, personID string) error {
	const op = "service.familytree.DeletePerson"

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	person, err := s.personStorage.GetPerson(ctx, parsedPersonID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			return fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	personsInTree, err := s.personStorage.GetPersonsByTree(ctx, person.TreeID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if len(personsInTree) <= 1 {
		return fmt.Errorf("%s: %w", op, ErrCannotDeleteLast)
	}

	relatives, err := s.relationStorage.GetRelatives(ctx, parsedPersonID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !canDeleteByRelationshipRules(relatives) {
		return fmt.Errorf("%s: %w", op, ErrDeleteNotAllowed)
	}

	if err := s.relationStorage.DeletePersonNode(ctx, parsedPersonID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.personStorage.DeletePerson(ctx, parsedPersonID); err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			return fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func canDeleteByRelationshipRules(relatives []models.Relative) bool {
	if len(relatives) == 0 {
		return false
	}

	parentCount := 0
	childCount := 0
	partnerCount := 0

	for _, relative := range relatives {
		switch relative.RelationshipType {
		case models.RelationshipParentChild:
			if relative.Direction == models.DirectionIncoming {
				parentCount++
			} else if relative.Direction == models.DirectionOutgoing {
				childCount++
			}
		case models.RelationshipPartner:
			partnerCount++
		}
	}

	if childCount == 0 && partnerCount == 0 && parentCount >= 1 && parentCount <= 2 {
		return true
	}

	if parentCount == 0 && childCount == 0 && partnerCount == 1 {
		return true
	}

	if parentCount == 0 && partnerCount == 0 && childCount == 1 {
		return true
	}

	return false
}

func (s *Service) GetTree(ctx context.Context, treeID string) ([]models.Person, []models.Relationship, error) {
	const op = "service.familytree.GetTree"

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	persons, err := s.personStorage.GetPersonsByTree(ctx, parsedTreeID)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	relationships, err := s.relationStorage.GetTreeRelationships(ctx, parsedTreeID)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	return persons, relationships, nil
}

func isValidGender(gender models.Gender) bool {
	return gender == models.GenderMale || gender == models.GenderFemale
}
