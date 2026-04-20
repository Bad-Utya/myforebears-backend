package familytree

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/storage"
	"github.com/google/uuid"
)

func (s *Service) AddRelationship(ctx context.Context, personIDFrom string, personIDTo string, relType models.RelationshipType) error {
	const op = "service.familytree.AddRelationship"

	if !isValidRelationshipType(relType) {
		return fmt.Errorf("%s: %w", op, ErrInvalidRelationType)
	}

	fromID, toID, err := parseRelationshipIDs(personIDFrom, personIDTo)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if fromID == toID {
		return fmt.Errorf("%s: %w", op, ErrSelfRelationship)
	}

	fromPerson, err := s.personStorage.GetPerson(ctx, fromID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			return fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	toPerson, err := s.personStorage.GetPerson(ctx, toID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			return fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if fromPerson.TreeID != toPerson.TreeID {
		return fmt.Errorf("%s: %w", op, ErrPersonNotInSameTree)
	}

	if relType == models.RelationshipPartner && fromID.String() > toID.String() {
		fromID, toID = toID, fromID
	}

	if err := s.relationStorage.CreateRelationship(ctx, fromID, toID, relType); err != nil {
		if errors.Is(err, storage.ErrRelationshipExists) {
			return fmt.Errorf("%s: %w", op, ErrRelationshipExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) RemoveRelationship(ctx context.Context, personIDFrom string, personIDTo string, relType models.RelationshipType) error {
	const op = "service.familytree.RemoveRelationship"

	if !isValidRelationshipType(relType) {
		return fmt.Errorf("%s: %w", op, ErrInvalidRelationType)
	}

	fromID, toID, err := parseRelationshipIDs(personIDFrom, personIDTo)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if relType == models.RelationshipPartner && fromID.String() > toID.String() {
		fromID, toID = toID, fromID
	}

	if err := s.relationStorage.RemoveRelationship(ctx, fromID, toID, relType); err != nil {
		if errors.Is(err, storage.ErrRelationshipMissing) {
			return fmt.Errorf("%s: %w", op, ErrRelationshipMissing)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) GetRelatives(ctx context.Context, personID string) ([]models.Relative, error) {
	const op = "service.familytree.GetRelatives"

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	if _, err := s.personStorage.GetPerson(ctx, parsedPersonID); err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			return nil, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	relatives, err := s.relationStorage.GetRelatives(ctx, parsedPersonID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	for i := range relatives {
		fullPerson, err := s.personStorage.GetPerson(ctx, relatives[i].Person.ID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		relatives[i].Person = fullPerson
	}

	return relatives, nil
}

func parseRelationshipIDs(personIDFrom string, personIDTo string) (uuid.UUID, uuid.UUID, error) {
	fromID, err := uuid.Parse(personIDFrom)
	if err != nil {
		return uuid.Nil, uuid.Nil, ErrInvalidPersonID
	}

	toID, err := uuid.Parse(personIDTo)
	if err != nil {
		return uuid.Nil, uuid.Nil, ErrInvalidPersonID
	}

	return fromID, toID, nil
}

func isValidRelationshipType(relType models.RelationshipType) bool {
	return relType == models.RelationshipParentChild || relType == models.RelationshipPartner
}
