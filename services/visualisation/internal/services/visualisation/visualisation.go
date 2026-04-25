package visualisation

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"log/slog"

	familytreepb "github.com/Bad-Utya/myforebears-backend/gen/go/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/storage"
	"github.com/google/uuid"
)

var (
	ErrInvalidVisualisationID   = errors.New("invalid visualisation id")
	ErrInvalidTreeID            = errors.New("invalid tree id")
	ErrInvalidRootPersonID      = errors.New("invalid root person id")
	ErrInvalidIncludedPersonID  = errors.New("invalid included person id")
	ErrIncludedPersonsRequired  = errors.New("included persons are required for full visualisation")
	ErrVisualisationNotFound    = errors.New("visualisation not found")
	ErrVisualisationNotReady    = errors.New("visualisation is not ready yet")
	ErrForbidden                = errors.New("forbidden")
	ErrGenerationNotImplemented = errors.New("visualisation algorithm is not implemented yet")
)

type FamilyTreeClient interface {
	GetPerson(ctx context.Context, treeID string, personID string) error
	GetTreeContent(ctx context.Context, treeID string) (*familytreepb.GetTreeContentResponse, error)
	GetTreeCreatorID(ctx context.Context, treeID string) (int, error)
}

type Service struct {
	log        *slog.Logger
	meta       storage.MetadataStorage
	objects    storage.ObjectStorage
	familyTree FamilyTreeClient
}

func New(log *slog.Logger, meta storage.MetadataStorage, objects storage.ObjectStorage, familyTree FamilyTreeClient) *Service {
	return &Service{log: log, meta: meta, objects: objects, familyTree: familyTree}
}

func (s *Service) CreateAncestorsVisualisation(ctx context.Context, treeID string, rootPersonID string) (models.Visualisation, error) {
	return s.createVisualisation(ctx, models.VisualisationTypeAncestors, treeID, rootPersonID, nil)
}

func (s *Service) CreateDescendantsVisualisation(ctx context.Context, treeID string, rootPersonID string) (models.Visualisation, error) {
	return s.createVisualisation(ctx, models.VisualisationTypeDescendants, treeID, rootPersonID, nil)
}

func (s *Service) CreateAncestorsAndDescendantsVisualisation(ctx context.Context, treeID string, rootPersonID string) (models.Visualisation, error) {
	return s.createVisualisation(ctx, models.VisualisationTypeAncestorsAndDescendants, treeID, rootPersonID, nil)
}

func (s *Service) CreateFullVisualisation(ctx context.Context, treeID string, rootPersonID string, includedPersonIDs []string) (models.Visualisation, error) {
	return s.createVisualisation(ctx, models.VisualisationTypeFull, treeID, rootPersonID, includedPersonIDs)
}

func (s *Service) ListTreeVisualisations(ctx context.Context, treeID string) ([]models.Visualisation, error) {
	const op = "service.visualisation.ListTreeVisualisations"

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	items, err := s.meta.ListTreeVisualisations(ctx, parsedTreeID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return items, nil
}

func (s *Service) GetVisualisationByID(ctx context.Context, treeID string, visualisationID string) (models.Visualisation, []byte, error) {
	const op = "service.visualisation.GetVisualisationByID"

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return models.Visualisation{}, nil, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	parsedVisualisationID, err := uuid.Parse(visualisationID)
	if err != nil {
		return models.Visualisation{}, nil, fmt.Errorf("%s: %w", op, ErrInvalidVisualisationID)
	}

	vis, err := s.meta.GetVisualisationByID(ctx, parsedVisualisationID)
	if err != nil {
		if errors.Is(err, storage.ErrVisualisationNotFound) {
			return models.Visualisation{}, nil, fmt.Errorf("%s: %w", op, ErrVisualisationNotFound)
		}
		return models.Visualisation{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	if vis.TreeID != parsedTreeID {
		return models.Visualisation{}, nil, fmt.Errorf("%s: %w", op, ErrForbidden)
	}

	if vis.Status != models.VisualisationStatusReady {
		return models.Visualisation{}, nil, fmt.Errorf("%s: %w", op, ErrVisualisationNotReady)
	}

	content, err := s.objects.GetObject(ctx, vis.ObjectKey)
	if err != nil {
		return models.Visualisation{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	return vis, content, nil
}

func (s *Service) DeleteVisualisationByID(ctx context.Context, treeID string, visualisationID string) error {
	const op = "service.visualisation.DeleteVisualisationByID"

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	parsedVisualisationID, err := uuid.Parse(visualisationID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, ErrInvalidVisualisationID)
	}

	vis, err := s.meta.GetVisualisationByID(ctx, parsedVisualisationID)
	if err != nil {
		if errors.Is(err, storage.ErrVisualisationNotFound) {
			return fmt.Errorf("%s: %w", op, ErrVisualisationNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if vis.TreeID != parsedTreeID {
		return fmt.Errorf("%s: %w", op, ErrForbidden)
	}

	deleted, err := s.meta.DeleteVisualisationByID(ctx, parsedVisualisationID)
	if err != nil {
		if errors.Is(err, storage.ErrVisualisationNotFound) {
			return fmt.Errorf("%s: %w", op, ErrVisualisationNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if deleted.Status == models.VisualisationStatusReady && strings.TrimSpace(deleted.ObjectKey) != "" {
		if err := s.objects.DeleteObject(ctx, deleted.ObjectKey); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (s *Service) createVisualisation(ctx context.Context, visType models.VisualisationType, treeID string, rootPersonID string, includedPersonIDs []string) (models.Visualisation, error) {
	const op = "service.visualisation.createVisualisation"

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return models.Visualisation{}, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	parsedRootPersonID, err := uuid.Parse(rootPersonID)
	if err != nil {
		return models.Visualisation{}, fmt.Errorf("%s: %w", op, ErrInvalidRootPersonID)
	}

	if err := s.familyTree.GetPerson(ctx, parsedTreeID.String(), parsedRootPersonID.String()); err != nil {
		return models.Visualisation{}, fmt.Errorf("%s: %w", op, err)
	}

	validatedIncludedIDs, err := s.validateIncludedPersonIDs(ctx, visType, parsedTreeID, includedPersonIDs)
	if err != nil {
		return models.Visualisation{}, fmt.Errorf("%s: %w", op, err)
	}

	ownerUserID, err := s.familyTree.GetTreeCreatorID(ctx, parsedTreeID.String())
	if err != nil {
		return models.Visualisation{}, fmt.Errorf("%s: %w", op, err)
	}

	now := time.Now()
	id := uuid.New()
	vis := models.Visualisation{
		ID:                id,
		OwnerUserID:       ownerUserID,
		TreeID:            parsedTreeID,
		RootPersonID:      parsedRootPersonID,
		IncludedPersonIDs: validatedIncludedIDs,
		Type:              visType,
		Status:            models.VisualisationStatusPending,
		FileName:          buildFileName(visType, id),
		MIMEType:          "image/svg+xml",
		SizeBytes:         0,
		ObjectKey:         buildObjectKey(parsedTreeID, id),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.meta.CreateVisualisation(ctx, vis); err != nil {
		return models.Visualisation{}, fmt.Errorf("%s: %w", op, err)
	}

	go s.runGeneration(vis)

	return vis, nil
}

func (s *Service) validateIncludedPersonIDs(ctx context.Context, visType models.VisualisationType, treeID uuid.UUID, includedPersonIDs []string) ([]uuid.UUID, error) {
	if visType != models.VisualisationTypeFull {
		return []uuid.UUID{}, nil
	}

	if len(includedPersonIDs) == 0 {
		return nil, ErrIncludedPersonsRequired
	}

	out := make([]uuid.UUID, 0, len(includedPersonIDs))
	seen := make(map[uuid.UUID]struct{}, len(includedPersonIDs))

	for _, personID := range includedPersonIDs {
		parsedPersonID, err := uuid.Parse(personID)
		if err != nil {
			return nil, ErrInvalidIncludedPersonID
		}

		if _, ok := seen[parsedPersonID]; ok {
			continue
		}

		if err := s.familyTree.GetPerson(ctx, treeID.String(), parsedPersonID.String()); err != nil {
			return nil, err
		}

		seen[parsedPersonID] = struct{}{}
		out = append(out, parsedPersonID)
	}

	return out, nil
}

func (s *Service) runGeneration(vis models.Visualisation) {
	ctx := context.Background()

	if err := s.meta.SetVisualisationProcessing(ctx, vis.ID); err != nil {
		s.log.Error("failed to set visualisation processing", slog.String("visualisation_id", vis.ID.String()), slog.String("error", err.Error()))
		return
	}

	svgContent, err := s.generateVisualisationFile(vis)
	if err != nil {
		if updateErr := s.meta.SetVisualisationFailed(ctx, vis.ID, err.Error()); updateErr != nil {
			s.log.Error("failed to set visualisation failed", slog.String("visualisation_id", vis.ID.String()), slog.String("error", updateErr.Error()))
		}
		return
	}

	if err := s.objects.PutObject(ctx, vis.ObjectKey, svgContent, vis.MIMEType); err != nil {
		if updateErr := s.meta.SetVisualisationFailed(ctx, vis.ID, err.Error()); updateErr != nil {
			s.log.Error("failed to set visualisation failed", slog.String("visualisation_id", vis.ID.String()), slog.String("error", updateErr.Error()))
		}
		return
	}

	if err := s.meta.SetVisualisationReady(ctx, vis.ID, int64(len(svgContent))); err != nil {
		s.log.Error("failed to set visualisation ready", slog.String("visualisation_id", vis.ID.String()), slog.String("error", err.Error()))
		return
	}
}

func (s *Service) generateVisualisationFile(vis models.Visualisation) ([]byte, error) {
	const op = "service.visualisation.generateVisualisationFile"

	content, err := s.familyTree.GetTreeContent(context.Background(), vis.TreeID.String())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	svgBytes, err := engine.RenderSVG(vis.Type, vis.RootPersonID, vis.IncludedPersonIDs, content)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return svgBytes, nil
}

func buildFileName(visType models.VisualisationType, visID uuid.UUID) string {
	normalizedType := strings.ReplaceAll(string(visType), "_", "-")
	return fmt.Sprintf("%s-%s.svg", normalizedType, visID.String())
}

func buildObjectKey(treeID uuid.UUID, visID uuid.UUID) string {
	return filepath.ToSlash(fmt.Sprintf("visualisations/%s/%s.svg", treeID.String(), visID.String()))
}
