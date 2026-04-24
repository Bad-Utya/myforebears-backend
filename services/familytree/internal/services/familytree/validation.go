package familytree

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/storage"
	"github.com/google/uuid"
)

func (s *Service) ValidatePersonsInTree(ctx context.Context, treeID string, personIDs []string) error {
	const op = "service.familytree.ValidatePersonsInTree"
	log := s.log.With(slog.String("op", op))

	log.Info("validating persons in tree", slog.String("tree_id", treeID), slog.Int("person_ids_count", len(personIDs)))

	parsedTreeID, err := s.authorizeTree(ctx, treeID)
	if err != nil {
		log.Error("failed to validate tree", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	seen := make(map[uuid.UUID]struct{}, len(personIDs))
	for _, rawID := range personIDs {
		parsedID, err := uuid.Parse(rawID)
		if err != nil {
			log.Info("invalid person id", slog.String("person_id", rawID))
			return fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
		}
		if _, ok := seen[parsedID]; ok {
			continue
		}
		seen[parsedID] = struct{}{}

		person, err := s.personStorage.GetPerson(ctx, parsedID)
		if err != nil {
			if errors.Is(err, storage.ErrPersonNotFound) {
				log.Info("person not found", slog.String("person_id", parsedID.String()))
				return fmt.Errorf("%s: %w", op, ErrPersonNotFound)
			}
			log.Error("failed to load person", slog.String("error", err.Error()))
			return fmt.Errorf("%s: %w", op, err)
		}
		if person.TreeID != parsedTreeID {
			log.Info("person tree mismatch", slog.String("person_id", parsedID.String()), slog.String("person_tree_id", person.TreeID.String()), slog.String("requested_tree_id", parsedTreeID.String()))
			return fmt.Errorf("%s: %w", op, ErrPersonNotInSameTree)
		}
	}

	log.Info("persons in tree validated")

	return nil
}
