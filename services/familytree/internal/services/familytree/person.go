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
	ErrInvalidPersonID         = errors.New("invalid person id")
	ErrInvalidTreeID           = errors.New("invalid tree id")
	ErrInvalidName             = errors.New("first name and last name are required")
	ErrInvalidGender           = errors.New("invalid gender")
	ErrInvalidUserID           = errors.New("invalid user id")
	ErrInvalidLimit            = errors.New("invalid limit")
	ErrForbidden               = errors.New("forbidden")
	ErrInvalidParentRole       = errors.New("invalid parent role")
	ErrParentExists            = errors.New("parent with this role already exists")
	ErrParentLimitReached      = errors.New("child already has two parents")
	ErrAtLeastOneParent        = errors.New("at least one parent id must be provided")
	ErrTreeMismatch            = errors.New("person does not belong to tree")
	ErrUnknownPersonGender     = errors.New("person gender is unknown")
	ErrPersonNotFound          = errors.New("person not found")
	ErrTreeNotFound            = errors.New("tree not found")
	ErrDeleteNotAllowed        = errors.New("delete is not allowed for current relationship set")
	ErrCannotDeleteLast        = errors.New("cannot delete the only person in tree")
	ErrSelfRelationship        = errors.New("self relationship is not allowed")
	ErrRelationshipExists      = errors.New("relationship already exists")
	ErrRelationshipMissing     = errors.New("relationship not found")
	ErrPersonNotInSameTree     = errors.New("persons must belong to the same tree")
	ErrInvalidRelationType     = errors.New("invalid relationship type")
	ErrInvalidEmail            = errors.New("invalid email")
	ErrTreeAccessEmailExists   = errors.New("tree access email already exists")
	ErrTreeAccessEmailNotFound = errors.New("tree access email not found")
	ErrInvalidMaxDepth         = errors.New("invalid max depth")
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
	log := s.log.With(slog.String("op", op))

	log.Info("creating person", slog.String("tree_id", treeID))

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		log.Info("invalid tree id", slog.String("tree_id", treeID))
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	if firstName == "" || lastName == "" {
		log.Info("invalid person name")
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidName)
	}

	if !isValidGender(gender) {
		log.Info("invalid gender", slog.String("gender", string(gender)))
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidGender)
	}

	_, err = s.personStorage.GetTree(ctx, parsedTreeID)
	if err != nil {
		if errors.Is(err, storage.ErrTreeNotFound) {
			log.Info("tree not found", slog.String("tree_id", parsedTreeID.String()))
			return models.Person{}, fmt.Errorf("%s: %w", op, ErrTreeNotFound)
		}
		log.Error("failed to load tree", slog.String("error", err.Error()))
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	person, err := s.createPersonRecord(ctx, parsedTreeID, gender, firstName, lastName, patronymic)
	if err != nil {
		log.Error("failed to create person", slog.String("error", err.Error()))
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("person created", slog.String("person_id", person.ID.String()))

	return person, nil
}

func (s *Service) GetPerson(ctx context.Context, treeID string, personID string) (models.Person, error) {
	const op = "service.familytree.GetPerson"
	log := s.log.With(slog.String("op", op))

	log.Info("getting person", slog.String("tree_id", treeID), slog.String("person_id", personID))

	parsedTreeID, err := s.authorizeTree(ctx, treeID)
	if err != nil {
		log.Error("failed to validate tree", slog.String("error", err.Error()))
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		log.Info("invalid person id", slog.String("person_id", personID))
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	person, err := s.personStorage.GetPerson(ctx, parsedPersonID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Info("person not found", slog.String("person_id", parsedPersonID.String()))
			return models.Person{}, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		log.Error("failed to get person", slog.String("error", err.Error()))
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	if person.TreeID != parsedTreeID {
		log.Info("person tree mismatch", slog.String("person_tree_id", person.TreeID.String()), slog.String("requested_tree_id", parsedTreeID.String()))
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrTreeMismatch)
	}

	log.Info("person loaded", slog.String("person_id", person.ID.String()))

	return person, nil
}

func (s *Service) UpdatePerson(
	ctx context.Context,
	treeID string,
	personID string,
	firstName string,
	lastName string,
	patronymic string,
	gender models.Gender,
) (models.Person, error) {
	const op = "service.familytree.UpdatePerson"
	log := s.log.With(slog.String("op", op))

	log.Info("updating person", slog.String("tree_id", treeID), slog.String("person_id", personID))

	parsedTreeID, err := s.authorizeTree(ctx, treeID)
	if err != nil {
		log.Error("failed to validate tree", slog.String("error", err.Error()))
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		log.Info("invalid person id", slog.String("person_id", personID))
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	if firstName == "" || lastName == "" {
		log.Info("invalid person name")
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidName)
	}

	if !isValidGender(gender) {
		log.Info("invalid gender", slog.String("gender", string(gender)))
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidGender)
	}

	existing, err := s.personStorage.GetPerson(ctx, parsedPersonID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Info("person not found", slog.String("person_id", parsedPersonID.String()))
			return models.Person{}, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		log.Error("failed to load person", slog.String("error", err.Error()))
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	existing.FirstName = firstName
	existing.LastName = lastName
	existing.Patronymic = patronymic
	existing.Gender = gender

	if existing.TreeID != parsedTreeID {
		log.Info("person tree mismatch", slog.String("person_tree_id", existing.TreeID.String()), slog.String("requested_tree_id", parsedTreeID.String()))
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrTreeMismatch)
	}

	if err := s.personStorage.UpdatePerson(ctx, existing); err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Info("person not found during update", slog.String("person_id", parsedPersonID.String()))
			return models.Person{}, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		log.Error("failed to update person", slog.String("error", err.Error()))
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("person updated", slog.String("person_id", existing.ID.String()))

	return existing, nil
}

func (s *Service) DeletePerson(ctx context.Context, treeID string, personID string) error {
	const op = "service.familytree.DeletePerson"
	log := s.log.With(slog.String("op", op))

	log.Info("deleting person", slog.String("tree_id", treeID), slog.String("person_id", personID))

	parsedTreeID, err := s.authorizeTree(ctx, treeID)
	if err != nil {
		log.Error("failed to validate tree", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		log.Info("invalid person id", slog.String("person_id", personID))
		return fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	person, err := s.personStorage.GetPerson(ctx, parsedPersonID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Info("person not found", slog.String("person_id", parsedPersonID.String()))
			return fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		log.Error("failed to load person", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	if person.TreeID != parsedTreeID {
		log.Info("person tree mismatch", slog.String("person_tree_id", person.TreeID.String()), slog.String("requested_tree_id", parsedTreeID.String()))
		return fmt.Errorf("%s: %w", op, ErrTreeMismatch)
	}

	personsInTree, err := s.personStorage.GetPersonsByTree(ctx, person.TreeID)
	if err != nil {
		log.Error("failed to list persons in tree", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}
	if len(personsInTree) <= 1 {
		log.Info("cannot delete last person in tree", slog.String("tree_id", person.TreeID.String()))
		return fmt.Errorf("%s: %w", op, ErrCannotDeleteLast)
	}

	relatives, err := s.relationStorage.GetRelatives(ctx, parsedPersonID)
	if err != nil {
		log.Error("failed to get relatives", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	if !canDeleteByRelationshipRules(relatives) {
		log.Info("delete not allowed by relationship rules")
		return fmt.Errorf("%s: %w", op, ErrDeleteNotAllowed)
	}

	if err := s.relationStorage.DeletePersonNode(ctx, parsedPersonID); err != nil {
		log.Error("failed to delete person node", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.personStorage.DeletePerson(ctx, parsedPersonID); err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Info("person not found during delete", slog.String("person_id", parsedPersonID.String()))
			return fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		log.Error("failed to delete person", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("person deleted", slog.String("person_id", parsedPersonID.String()))

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

func (s *Service) getTreeContentRaw(ctx context.Context, treeID string) ([]models.Person, []models.Relationship, error) {
	const op = "service.familytree.getTreeContentRaw"
	log := s.log.With(slog.String("op", op))

	log.Info("getting tree content", slog.String("tree_id", treeID))

	parsedTreeID, err := s.authorizeTree(ctx, treeID)
	if err != nil {
		log.Error("failed to validate tree", slog.String("error", err.Error()))
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	persons, err := s.personStorage.GetPersonsByTree(ctx, parsedTreeID)
	if err != nil {
		log.Error("failed to load persons by tree", slog.String("error", err.Error()))
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	relationships, err := s.relationStorage.GetTreeRelationships(ctx, parsedTreeID)
	if err != nil {
		log.Error("failed to load tree relationships", slog.String("error", err.Error()))
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("tree loaded", slog.Int("persons_count", len(persons)), slog.Int("relationships_count", len(relationships)))

	return persons, relationships, nil
}

func isValidGender(gender models.Gender) bool {
	return gender == models.GenderMale || gender == models.GenderFemale
}
