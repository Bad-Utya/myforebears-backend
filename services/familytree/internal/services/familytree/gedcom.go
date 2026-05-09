package familytree

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/lib/gedcom"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/storage"
	"github.com/google/uuid"
)

// ExportTreeGEDCOM exports a family tree to GEDCOM format
func (s *Service) ExportTreeGEDCOM(ctx context.Context, requestUserID int, treeID string) (string, error) {
	const op = "service.familytree.ExportTreeGEDCOM"
	log := s.log.With(slog.String("op", op))

	log.Info("exporting tree to GEDCOM", slog.String("tree_id", treeID), slog.Int("request_user_id", requestUserID))

	// Parse and validate tree ID
	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		log.Info("invalid tree id", slog.String("tree_id", treeID))
		return "", fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	// Verify user has access to the tree
	tree, err := s.personStorage.GetTree(ctx, parsedTreeID)
	if err != nil {
		if err == storage.ErrTreeNotFound {
			log.Info("tree not found", slog.String("tree_id", treeID))
			return "", fmt.Errorf("%s: %w", op, ErrTreeNotFound)
		}
		log.Error("failed to load tree", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	// Only creator can export
	if tree.CreatorID != requestUserID {
		log.Warn("user does not have permission to export tree", slog.String("tree_id", treeID), slog.Int("request_user_id", requestUserID))
		return "", fmt.Errorf("%s: %w", op, ErrForbidden)
	}

	// Get all persons and relationships
	persons, err := s.personStorage.GetPersonsByTree(ctx, parsedTreeID)
	if err != nil {
		log.Error("failed to load persons", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	relationships, err := s.relationStorage.GetTreeRelationships(ctx, parsedTreeID)
	if err != nil {
		log.Error("failed to load relationships", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	// Convert to GEDCOM
	gedcomContent := gedcom.ExportToGEDCOM(persons, relationships)
	log.Info("tree exported to GEDCOM", slog.String("tree_id", treeID), slog.Int("persons_count", len(persons)), slog.Int("relationships_count", len(relationships)))

	return gedcomContent, nil
}

// ImportTreeGEDCOM imports a GEDCOM file and creates a new family tree
func (s *Service) ImportTreeGEDCOM(ctx context.Context, requestUserID int, gedcomContent string) (models.Tree, int, int, []string, error) {
	const op = "service.familytree.ImportTreeGEDCOM"
	log := s.log.With(slog.String("op", op))

	log.Info("importing GEDCOM content", slog.Int("request_user_id", requestUserID), slog.Int("content_length", len(gedcomContent)))

	if requestUserID <= 0 {
		return models.Tree{}, 0, 0, nil, fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}

	// Parse GEDCOM content
	parsedData := gedcom.ImportFromGEDCOM(gedcomContent)

	log.Info("GEDCOM parsed",
		slog.Int("persons_count", len(parsedData.Persons)),
		slog.Int("relationships_count", len(parsedData.Relationships)),
		slog.Int("errors_count", len(parsedData.Errors)))

	if len(parsedData.Persons) == 0 {
		return models.Tree{}, 0, 0, parsedData.Errors, fmt.Errorf("%s: no persons found in GEDCOM", op)
	}

	// Create a new tree
	treeID := uuid.New()
	tree := models.Tree{
		ID:        treeID,
		CreatorID: requestUserID,
		CreatedAt: time.Now(),
	}

	// Create tree in storage
	if err := s.personStorage.CreateTree(ctx, tree); err != nil {
		log.Error("failed to create tree", slog.String("error", err.Error()))
		return models.Tree{}, 0, 0, parsedData.Errors, fmt.Errorf("%s: %w", op, err)
	}

	// Update persons with the new tree ID
	personIDMap := make(map[string]uuid.UUID) // Old UUID -> New UUID
	for i, person := range parsedData.Persons {
		newID := uuid.New()
		personIDMap[person.ID.String()] = newID
		parsedData.Persons[i].ID = newID
		parsedData.Persons[i].TreeID = treeID
	}

	// Create all persons
	for _, person := range parsedData.Persons {
		if err := s.personStorage.CreatePerson(ctx, person); err != nil {
			log.Error("failed to create person", slog.String("error", err.Error()))
			parsedData.Errors = append(parsedData.Errors, fmt.Sprintf("failed to create person %s %s: %v", person.FirstName, person.LastName, err))
		}
	}

	// Update relationship IDs to match new person IDs
	for i, rel := range parsedData.Relationships {
		oldFromID := rel.PersonIDFrom.String()
		oldToID := rel.PersonIDTo.String()

		if newFromID, ok := personIDMap[oldFromID]; ok {
			parsedData.Relationships[i].PersonIDFrom = newFromID
		}
		if newToID, ok := personIDMap[oldToID]; ok {
			parsedData.Relationships[i].PersonIDTo = newToID
		}
	}

	// Create all relationships
	successRelCount := 0
	for _, rel := range parsedData.Relationships {
		if err := s.relationStorage.CreateRelationship(ctx, rel.PersonIDFrom, rel.PersonIDTo, rel.Type); err != nil {
			log.Error("failed to create relationship", slog.String("error", err.Error()))
			parsedData.Errors = append(parsedData.Errors, fmt.Sprintf("failed to create relationship: %v", err))
			continue
		}
		successRelCount++
	}

	log.Info("tree imported from GEDCOM",
		slog.String("tree_id", tree.ID.String()),
		slog.Int("persons_created", len(parsedData.Persons)),
		slog.Int("relationships_created", successRelCount),
		slog.Int("errors_count", len(parsedData.Errors)))

	return tree, len(parsedData.Persons), successRelCount, parsedData.Errors, nil
}
