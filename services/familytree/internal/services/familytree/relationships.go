package familytree

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/storage"
	"github.com/google/uuid"
)

func (s *Service) AddRelationship(ctx context.Context, personIDFrom string, personIDTo string, relType models.RelationshipType) error {
	const op = "service.familytree.AddRelationship"
	log := s.log.With(slog.String("op", op))

	log.Info("adding relationship", slog.String("person_id_from", personIDFrom), slog.String("person_id_to", personIDTo), slog.String("relationship_type", string(relType)))

	if !isValidRelationshipType(relType) {
		log.Info("invalid relationship type", slog.String("relationship_type", string(relType)))
		return fmt.Errorf("%s: %w", op, ErrInvalidRelationType)
	}

	fromID, toID, err := parseRelationshipIDs(personIDFrom, personIDTo)
	if err != nil {
		log.Info("invalid relationship person ids", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	if fromID == toID {
		log.Info("self relationship is not allowed", slog.String("person_id", fromID.String()))
		return fmt.Errorf("%s: %w", op, ErrSelfRelationship)
	}

	fromPerson, err := s.personStorage.GetPerson(ctx, fromID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Info("from person not found", slog.String("person_id", fromID.String()))
			return fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		log.Error("failed to load from person", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	toPerson, err := s.personStorage.GetPerson(ctx, toID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Info("to person not found", slog.String("person_id", toID.String()))
			return fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		log.Error("failed to load to person", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	if fromPerson.TreeID != toPerson.TreeID {
		log.Info("persons belong to different trees", slog.String("from_tree_id", fromPerson.TreeID.String()), slog.String("to_tree_id", toPerson.TreeID.String()))
		return fmt.Errorf("%s: %w", op, ErrPersonNotInSameTree)
	}

	if isPartnerRelationshipType(relType) && fromID.String() > toID.String() {
		fromID, toID = toID, fromID
	}

	if err := s.relationStorage.CreateRelationship(ctx, fromID, toID, relType); err != nil {
		if errors.Is(err, storage.ErrRelationshipExists) {
			log.Info("relationship already exists")
			return fmt.Errorf("%s: %w", op, ErrRelationshipExists)
		}
		log.Error("failed to add relationship", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("relationship added", slog.String("person_id_from", fromID.String()), slog.String("person_id_to", toID.String()))

	return nil
}

func (s *Service) RemoveRelationship(ctx context.Context, personIDFrom string, personIDTo string, relType models.RelationshipType) error {
	const op = "service.familytree.RemoveRelationship"
	log := s.log.With(slog.String("op", op))

	log.Info("removing relationship", slog.String("person_id_from", personIDFrom), slog.String("person_id_to", personIDTo), slog.String("relationship_type", string(relType)))

	if !isValidRelationshipType(relType) {
		log.Info("invalid relationship type", slog.String("relationship_type", string(relType)))
		return fmt.Errorf("%s: %w", op, ErrInvalidRelationType)
	}

	fromID, toID, err := parseRelationshipIDs(personIDFrom, personIDTo)
	if err != nil {
		log.Info("invalid relationship person ids", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	if isPartnerRelationshipType(relType) && fromID.String() > toID.String() {
		fromID, toID = toID, fromID
	}

	if err := s.relationStorage.RemoveRelationship(ctx, fromID, toID, relType); err != nil {
		if errors.Is(err, storage.ErrRelationshipMissing) {
			log.Info("relationship not found")
			return fmt.Errorf("%s: %w", op, ErrRelationshipMissing)
		}
		log.Error("failed to remove relationship", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("relationship removed", slog.String("person_id_from", fromID.String()), slog.String("person_id_to", toID.String()))

	return nil
}

func (s *Service) GetRelatives(ctx context.Context, personID string) ([]models.Relative, error) {
	const op = "service.familytree.GetRelatives"
	log := s.log.With(slog.String("op", op))

	log.Info("getting relatives", slog.String("person_id", personID))

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		log.Info("invalid person id", slog.String("person_id", personID))
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	if _, err := s.personStorage.GetPerson(ctx, parsedPersonID); err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Info("person not found", slog.String("person_id", parsedPersonID.String()))
			return nil, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		log.Error("failed to get person", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	relatives, err := s.relationStorage.GetRelatives(ctx, parsedPersonID)
	if err != nil {
		log.Error("failed to get relatives", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	for i := range relatives {
		fullPerson, err := s.personStorage.GetPerson(ctx, relatives[i].Person.ID)
		if err != nil {
			log.Error("failed to expand relative person", slog.String("error", err.Error()))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		relatives[i].Person = fullPerson
	}

	log.Info("relatives loaded", slog.Int("count", len(relatives)))

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
	return relType == models.RelationshipParentChild || isPartnerRelationshipType(relType)
}

func isPartnerRelationshipType(relType models.RelationshipType) bool {
	return relType == models.RelationshipPartner ||
		relType == models.RelationshipPartnerUnmarried ||
		relType == models.RelationshipPartnerMarried ||
		relType == models.RelationshipPartnerDivorced
}
