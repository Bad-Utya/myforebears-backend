package familytree

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/storage"
	"github.com/google/uuid"
)

func (s *Service) ValidatePersonsInTree(ctx context.Context, requestUserID int, treeID string, personIDs []string) error {
	const op = "service.familytree.ValidatePersonsInTree"

	parsedTreeID, err := s.authorizeTree(ctx, requestUserID, treeID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	seen := make(map[uuid.UUID]struct{}, len(personIDs))
	for _, rawID := range personIDs {
		parsedID, err := uuid.Parse(rawID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
		}
		if _, ok := seen[parsedID]; ok {
			continue
		}
		seen[parsedID] = struct{}{}

		person, err := s.personStorage.GetPerson(ctx, parsedID)
		if err != nil {
			if errors.Is(err, storage.ErrPersonNotFound) {
				return fmt.Errorf("%s: %w", op, ErrPersonNotFound)
			}
			return fmt.Errorf("%s: %w", op, err)
		}
		if person.TreeID != parsedTreeID {
			return fmt.Errorf("%s: %w", op, ErrPersonNotInSameTree)
		}
	}

	return nil
}
