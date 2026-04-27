package familytree

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	eventspb "github.com/Bad-Utya/myforebears-backend/gen/go/events"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/storage"
	"github.com/google/uuid"
)

type personStorageStub struct {
	t           *testing.T
	getTreeFn   func(ctx context.Context, treeID uuid.UUID) (models.Tree, error)
	getPersonFn func(ctx context.Context, personID uuid.UUID) (models.Person, error)
}

func (s *personStorageStub) CreateTree(ctx context.Context, tree models.Tree) error {
	s.t.Fatalf("unexpected CreateTree call")
	return nil
}

func (s *personStorageStub) GetTree(ctx context.Context, treeID uuid.UUID) (models.Tree, error) {
	if s.getTreeFn == nil {
		s.t.Fatalf("unexpected GetTree call")
	}
	return s.getTreeFn(ctx, treeID)
}

func (s *personStorageStub) UpdateTreeRootPersonID(ctx context.Context, treeID uuid.UUID, rootPersonID uuid.UUID) error {
	s.t.Fatalf("unexpected UpdateTreeRootPersonID call")
	return nil
}

func (s *personStorageStub) UpdateTreeSettings(ctx context.Context, treeID uuid.UUID, isViewRestricted bool, isPublicOnMainPage bool, name string) error {
	s.t.Fatalf("unexpected UpdateTreeSettings call")
	return nil
}

func (s *personStorageStub) AddTreeAccessEmail(ctx context.Context, treeID uuid.UUID, email string) error {
	s.t.Fatalf("unexpected AddTreeAccessEmail call")
	return nil
}

func (s *personStorageStub) ListTreeAccessEmails(ctx context.Context, treeID uuid.UUID) ([]string, error) {
	s.t.Fatalf("unexpected ListTreeAccessEmails call")
	return nil, nil
}

func (s *personStorageStub) IsTreeAccessEmailAllowed(ctx context.Context, treeID uuid.UUID, email string) (bool, error) {
	s.t.Fatalf("unexpected IsTreeAccessEmailAllowed call")
	return false, nil
}

func (s *personStorageStub) DeleteTreeAccessEmail(ctx context.Context, treeID uuid.UUID, email string) error {
	s.t.Fatalf("unexpected DeleteTreeAccessEmail call")
	return nil
}

func (s *personStorageStub) GetTreesByCreator(ctx context.Context, creatorID int) ([]models.Tree, error) {
	s.t.Fatalf("unexpected GetTreesByCreator call")
	return nil, nil
}

func (s *personStorageStub) GetPublicTreesByCreator(ctx context.Context, creatorID int) ([]models.Tree, error) {
	s.t.Fatalf("unexpected GetPublicTreesByCreator call")
	return nil, nil
}

func (s *personStorageStub) GetRandomPublicTrees(ctx context.Context, limit int) ([]models.Tree, error) {
	s.t.Fatalf("unexpected GetRandomPublicTrees call")
	return nil, nil
}

func (s *personStorageStub) CreatePerson(ctx context.Context, person models.Person) error {
	s.t.Fatalf("unexpected CreatePerson call")
	return nil
}

func (s *personStorageStub) GetPerson(ctx context.Context, personID uuid.UUID) (models.Person, error) {
	if s.getPersonFn == nil {
		s.t.Fatalf("unexpected GetPerson call")
	}
	return s.getPersonFn(ctx, personID)
}

func (s *personStorageStub) UpdatePerson(ctx context.Context, person models.Person) error {
	s.t.Fatalf("unexpected UpdatePerson call")
	return nil
}

func (s *personStorageStub) UpdatePersonAvatarPhoto(ctx context.Context, personID uuid.UUID, avatarPhotoID *uuid.UUID) error {
	s.t.Fatalf("unexpected UpdatePersonAvatarPhoto call")
	return nil
}

func (s *personStorageStub) DeletePerson(ctx context.Context, personID uuid.UUID) error {
	s.t.Fatalf("unexpected DeletePerson call")
	return nil
}

func (s *personStorageStub) GetPersonsByTree(ctx context.Context, treeID uuid.UUID) ([]models.Person, error) {
	s.t.Fatalf("unexpected GetPersonsByTree call")
	return nil, nil
}

func (s *personStorageStub) Close() {}

type relationStorageStub struct {
	t *testing.T
}

func (s *relationStorageStub) EnsurePersonNode(ctx context.Context, personID uuid.UUID, treeID uuid.UUID) error {
	s.t.Fatalf("unexpected EnsurePersonNode call")
	return nil
}

func (s *relationStorageStub) DeletePersonNode(ctx context.Context, personID uuid.UUID) error {
	s.t.Fatalf("unexpected DeletePersonNode call")
	return nil
}

func (s *relationStorageStub) CreateRelationship(ctx context.Context, personIDFrom uuid.UUID, personIDTo uuid.UUID, relType models.RelationshipType) error {
	s.t.Fatalf("unexpected CreateRelationship call")
	return nil
}

func (s *relationStorageStub) SetPartnerRelationshipStatus(ctx context.Context, personID1 uuid.UUID, personID2 uuid.UUID, status models.PartnerRelationshipStatus) error {
	s.t.Fatalf("unexpected SetPartnerRelationshipStatus call")
	return nil
}

func (s *relationStorageStub) RemoveRelationship(ctx context.Context, personIDFrom uuid.UUID, personIDTo uuid.UUID, relType models.RelationshipType) error {
	s.t.Fatalf("unexpected RemoveRelationship call")
	return nil
}

func (s *relationStorageStub) GetRelatives(ctx context.Context, personID uuid.UUID) ([]models.Relative, error) {
	s.t.Fatalf("unexpected GetRelatives call")
	return nil, nil
}

func (s *relationStorageStub) GetTreeRelationships(ctx context.Context, treeID uuid.UUID) ([]models.Relationship, error) {
	s.t.Fatalf("unexpected GetTreeRelationships call")
	return nil, nil
}

func (s *relationStorageStub) GetTreeRelationshipsWithinDepth(ctx context.Context, treeID uuid.UUID, rootPersonID uuid.UUID, maxDepth int) ([]models.Relationship, error) {
	s.t.Fatalf("unexpected GetTreeRelationshipsWithinDepth call")
	return nil, nil
}

func (s *relationStorageStub) Close(ctx context.Context) error {
	s.t.Fatalf("unexpected Close call")
	return nil
}

type eventsClientStub struct {
	t *testing.T
}

func (s *eventsClientStub) CreateEvent(ctx context.Context, req *eventspb.CreateEventRequest) (*eventspb.CreateEventResponse, error) {
	s.t.Fatalf("unexpected CreateEvent call")
	return nil, nil
}

func TestValidatePersonsInTreeInvalidID(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	treeID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	personStore := &personStorageStub{
		t: t,
		getTreeFn: func(ctx context.Context, id uuid.UUID) (models.Tree, error) {
			return models.Tree{ID: treeID}, nil
		},
	}

	svc := &Service{log: log, personStorage: personStore, relationStorage: &relationStorageStub{t: t}, eventsClient: &eventsClientStub{t: t}}
	if err := svc.ValidatePersonsInTree(ctx, treeID.String(), []string{"bad"}); !errors.Is(err, ErrInvalidPersonID) {
		t.Fatalf("expected ErrInvalidPersonID, got %v", err)
	}
}

func TestValidatePersonsInTreePersonNotFound(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	treeID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	personID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	personStore := &personStorageStub{
		t: t,
		getTreeFn: func(ctx context.Context, id uuid.UUID) (models.Tree, error) {
			return models.Tree{ID: treeID}, nil
		},
		getPersonFn: func(ctx context.Context, id uuid.UUID) (models.Person, error) {
			if id != personID {
				t.Fatalf("unexpected person id %s", id.String())
			}
			return models.Person{}, storage.ErrPersonNotFound
		},
	}

	svc := &Service{log: log, personStorage: personStore, relationStorage: &relationStorageStub{t: t}, eventsClient: &eventsClientStub{t: t}}
	if err := svc.ValidatePersonsInTree(ctx, treeID.String(), []string{personID.String()}); !errors.Is(err, ErrPersonNotFound) {
		t.Fatalf("expected ErrPersonNotFound, got %v", err)
	}
}

func TestValidatePersonsInTreeTreeMismatch(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	treeID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	otherTreeID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	personID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	personStore := &personStorageStub{
		t: t,
		getTreeFn: func(ctx context.Context, id uuid.UUID) (models.Tree, error) {
			return models.Tree{ID: treeID}, nil
		},
		getPersonFn: func(ctx context.Context, id uuid.UUID) (models.Person, error) {
			return models.Person{ID: personID, TreeID: otherTreeID}, nil
		},
	}

	svc := &Service{log: log, personStorage: personStore, relationStorage: &relationStorageStub{t: t}, eventsClient: &eventsClientStub{t: t}}
	if err := svc.ValidatePersonsInTree(ctx, treeID.String(), []string{personID.String()}); !errors.Is(err, ErrPersonNotInSameTree) {
		t.Fatalf("expected ErrPersonNotInSameTree, got %v", err)
	}
}

func TestNormalizeEmail(t *testing.T) {
	got, err := normalizeEmail(" Test@Example.com ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "test@example.com" {
		t.Fatalf("expected normalized email, got %q", got)
	}

	if _, err := normalizeEmail("bad@"); !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestValidatePersonsInTreeSuccessWithDuplicates(t *testing.T) {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	treeID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	personID1 := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	personID2 := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	callCount := map[uuid.UUID]int{}
	personStore := &personStorageStub{
		t: t,
		getTreeFn: func(ctx context.Context, id uuid.UUID) (models.Tree, error) {
			return models.Tree{ID: treeID}, nil
		},
		getPersonFn: func(ctx context.Context, id uuid.UUID) (models.Person, error) {
			callCount[id]++
			return models.Person{ID: id, TreeID: treeID}, nil
		},
	}

	svc := &Service{log: log, personStorage: personStore, relationStorage: &relationStorageStub{t: t}, eventsClient: &eventsClientStub{t: t}}
	if err := svc.ValidatePersonsInTree(ctx, treeID.String(), []string{personID1.String(), personID1.String(), personID2.String()}); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if callCount[personID1] != 1 || callCount[personID2] != 1 {
		t.Fatalf("expected single GetPerson call per unique id, got %v", callCount)
	}
}
