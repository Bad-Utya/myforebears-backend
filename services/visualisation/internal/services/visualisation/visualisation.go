package visualisation

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"log/slog"

	eventspb "github.com/Bad-Utya/myforebears-backend/gen/go/events"
	familytreepb "github.com/Bad-Utya/myforebears-backend/gen/go/familytree"
	photospb "github.com/Bad-Utya/myforebears-backend/gen/go/photos"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage4_render"
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
	GetTreeContentWithinDepth(ctx context.Context, treeID string, rootPersonID string, maxDepth int32) (*familytreepb.GetTreeContentResponse, error)
	GetTreeCreatorID(ctx context.Context, treeID string) (int, error)
}

type PhotosClient interface {
	GetPersonAvatar(ctx context.Context, req *photospb.GetPersonAvatarRequest) (*photospb.GetPhotoContentResponse, error)
}

type EventsClient interface {
	ListEventsByTree(ctx context.Context, req *eventspb.ListEventsByTreeRequest) (*eventspb.ListEventsByTreeResponse, error)
}

type Service struct {
	log        *slog.Logger
	meta       storage.MetadataStorage
	objects    storage.ObjectStorage
	familyTree FamilyTreeClient
	photos     PhotosClient
	events     EventsClient
	workerCtx  context.Context
	cancel     context.CancelFunc
	workers    sync.WaitGroup
}

func New(log *slog.Logger, meta storage.MetadataStorage, objects storage.ObjectStorage, familyTree FamilyTreeClient, photos PhotosClient, events EventsClient) *Service {
	workerCtx, cancel := context.WithCancel(context.Background())
	return &Service{log: log, meta: meta, objects: objects, familyTree: familyTree, photos: photos, events: events, workerCtx: workerCtx, cancel: cancel}
}

const generationTimeout = 2 * time.Minute

// Close cancels active generations and waits until their cleanup has finished.
func (s *Service) Close() {
	if s.cancel != nil {
		s.cancel()
	}
	s.workers.Wait()
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

	s.workers.Add(1)
	go func() {
		defer s.workers.Done()
		s.runGeneration(vis)
	}()

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
	ctx, cancel := context.WithTimeout(s.workerCtx, generationTimeout)
	defer cancel()

	if err := s.meta.SetVisualisationProcessing(ctx, vis.ID); err != nil {
		s.log.Error("failed to set visualisation processing", slog.String("visualisation_id", vis.ID.String()), slog.String("error", err.Error()))
		return
	}

	svgContent, err := s.generateVisualisationFile(ctx, vis)
	if err != nil {
		s.markGenerationFailed(vis.ID, err)
		return
	}

	if err := s.objects.PutObject(ctx, vis.ObjectKey, svgContent, vis.MIMEType); err != nil {
		s.markGenerationFailed(vis.ID, err)
		return
	}

	if err := s.meta.SetVisualisationReady(ctx, vis.ID, int64(len(svgContent))); err != nil {
		s.log.Error("failed to set visualisation ready", slog.String("visualisation_id", vis.ID.String()), slog.String("error", err.Error()))
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		if cleanupErr := s.objects.DeleteObject(cleanupCtx, vis.ObjectKey); cleanupErr != nil {
			s.log.Error("failed to remove orphaned visualisation", slog.String("visualisation_id", vis.ID.String()), slog.String("error", cleanupErr.Error()))
		}
		s.markGenerationFailed(vis.ID, err)
		return
	}
}

func (s *Service) markGenerationFailed(visualisationID uuid.UUID, generationErr error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.meta.SetVisualisationFailed(ctx, visualisationID, generationErr.Error()); err != nil {
		s.log.Error("failed to set visualisation failed", slog.String("visualisation_id", visualisationID.String()), slog.String("error", err.Error()))
	}
}

func (s *Service) generateVisualisationFile(ctx context.Context, vis models.Visualisation) ([]byte, error) {
	const op = "service.visualisation.generateVisualisationFile"

	content, err := s.familyTree.GetTreeContent(ctx, vis.TreeID.String())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	personData, err := s.buildPersonRenderData(ctx, vis.TreeID.String(), content.GetPersons())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	svgBytes, err := engine.RenderSVGWithPersonData(vis.Type, vis.RootPersonID, vis.IncludedPersonIDs, content, personData)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return svgBytes, nil
}

const (
	birthEventTypeID = "4af8f935-180f-4be6-8f7a-f6ecf90af4b2"
	deathEventTypeID = "7e92347c-b30d-474e-abdd-48f62cb0f6cf"
)

type personDates struct {
	birth string
	death string

	hasDeath bool
}

func (s *Service) buildPersonRenderData(ctx context.Context, treeID string, persons []*familytreepb.Person) (map[string]stage4_render.PersonRenderData, error) {
	personData := make(map[string]stage4_render.PersonRenderData, len(persons))
	if len(persons) == 0 {
		return personData, nil
	}

	dates := make(map[string]personDates)
	if s.events != nil {
		eventsResp, err := s.events.ListEventsByTree(ctx, &eventspb.ListEventsByTreeRequest{TreeId: treeID})
		if err != nil {
			s.log.Warn("events lookup failed", slog.String("tree_id", treeID), slog.String("error", err.Error()))
		} else {
			dates = collectPersonDates(eventsResp.GetEvents())
		}
	}

	for _, person := range persons {
		if person == nil {
			continue
		}

		data := stage4_render.PersonRenderData{
			DisplayName: engine.BuildPersonDisplayName(person.GetFirstName(), person.GetLastName(), person.GetPatronymic(), person.GetId()),
			DateLine:    formatPersonDateLine(dates[person.GetId()]),
		}

		if s.photos != nil {
			avatar, err := s.photos.GetPersonAvatar(ctx, &photospb.GetPersonAvatarRequest{
				TreeId:   treeID,
				PersonId: person.GetId(),
			})
			if err == nil {
				mimeType := avatar.GetPhoto().GetMimeType()
				if mimeType != "" && len(avatar.GetContent()) > 0 {
					data.AvatarMime = mimeType
					data.AvatarData = base64.StdEncoding.EncodeToString(avatar.GetContent())
				}
			}
		}

		personData[person.GetId()] = data
	}

	return personData, nil
}

func collectPersonDates(events []*eventspb.Event) map[string]personDates {
	result := make(map[string]personDates)
	for _, event := range events {
		if event == nil {
			continue
		}

		if event.GetEventTypeId() != birthEventTypeID && event.GetEventTypeId() != deathEventTypeID {
			continue
		}

		formatted := formatEventDate(event)
		for _, personID := range event.GetPrimaryPersonIds() {
			dates := result[personID]
			if event.GetEventTypeId() == birthEventTypeID {
				if dates.birth == "" {
					dates.birth = formatted
				}
			} else {
				if dates.death == "" {
					dates.death = formatted
					dates.hasDeath = true
				}
			}
			result[personID] = dates
		}
	}

	return result
}

func formatPersonDateLine(dates personDates) string {
	birth := dates.birth
	death := dates.death
	if birth == "" {
		birth = "неиз"
	}

	if !dates.hasDeath {
		return birth
	}

	if death == "" {
		death = "неиз"
	}

	return fmt.Sprintf("%s-%s", birth, death)
}

func formatEventDate(event *eventspb.Event) string {
	if event == nil {
		return ""
	}
	if event.GetDateUnknown() {
		return "неиз"
	}
	dateISO := strings.TrimSpace(event.GetDateIso())
	if dateISO == "" {
		return "неиз"
	}
	if t, err := time.Parse("2006-01-02", dateISO); err == nil {
		return t.Format("02.01.2006")
	}
	if t, err := time.Parse(time.RFC3339, dateISO); err == nil {
		return t.Format("02.01.2006")
	}

	return "неиз"
}

// RenderCoordinatesForClient renders tree visualization for client-side rendering
// maxDepth: 0 = unlimited; N > 0 = include people up to N hops from root
// allowLayerShift: if false, disable layer shifting optimization
func (s *Service) RenderCoordinatesForClient(
	ctx context.Context,
	treeID string,
	rootPersonID string,
	maxDepth int,
) ([]byte, error) {
	const op = "service.visualisation.RenderCoordinatesForClient"

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	parsedRootPersonID, err := uuid.Parse(rootPersonID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidRootPersonID)
	}

	if err := s.familyTree.GetPerson(ctx, parsedTreeID.String(), parsedRootPersonID.String()); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if maxDepth < 0 {
		return nil, fmt.Errorf("%s: maxDepth must be >= 0", op)
	}

	content, err := s.familyTree.GetTreeContentWithinDepth(ctx, parsedTreeID.String(), parsedRootPersonID.String(), int32(maxDepth))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	coordBytes, err := engine.RenderCoordinates(models.VisualisationTypeFull, parsedRootPersonID, nil, content, 0, false)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return coordBytes, nil
}

func buildFileName(visType models.VisualisationType, visID uuid.UUID) string {
	normalizedType := strings.ReplaceAll(string(visType), "_", "-")
	return fmt.Sprintf("%s-%s.svg", normalizedType, visID.String())
}

func buildObjectKey(treeID uuid.UUID, visID uuid.UUID) string {
	return filepath.ToSlash(fmt.Sprintf("visualisations/%s/%s.svg", treeID.String(), visID.String()))
}
