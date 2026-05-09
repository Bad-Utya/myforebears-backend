package visualisation

import (
	"context"
	"errors"
	"testing"

	familytreepb "github.com/Bad-Utya/myforebears-backend/gen/go/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/domain/models"
	"github.com/google/uuid"
)

type familyTreeStub struct {
	t *testing.T
}

func (f *familyTreeStub) GetPerson(ctx context.Context, treeID string, personID string) error {
	return nil
}

func (f *familyTreeStub) GetTreeContent(ctx context.Context, treeID string) (*familytreepb.GetTreeContentResponse, error) {
	f.t.Fatalf("unexpected GetTreeContent call")
	return nil, nil
}

func (f *familyTreeStub) GetTreeContentWithinDepth(ctx context.Context, treeID string, rootPersonID string, maxDepth int32) (*familytreepb.GetTreeContentResponse, error) {
	f.t.Fatalf("unexpected GetTreeContentWithinDepth call")
	return nil, nil
}

func (f *familyTreeStub) GetTreeCreatorID(ctx context.Context, treeID string) (int, error) {
	f.t.Fatalf("unexpected GetTreeCreatorID call")
	return 0, nil
}

func TestValidateIncludedPersonIDsRequiresList(t *testing.T) {
	svc := &Service{familyTree: &familyTreeStub{t: t}}
	_, err := svc.validateIncludedPersonIDs(context.Background(), models.VisualisationTypeFull, uuid.New(), nil)
	if !errors.Is(err, ErrIncludedPersonsRequired) {
		t.Fatalf("expected ErrIncludedPersonsRequired, got %v", err)
	}
}

func TestValidateIncludedPersonIDsDeduplicates(t *testing.T) {
	treeID := uuid.New()
	personID := uuid.New()
	stub := &familyTreeStub{t: t}

	svc := &Service{familyTree: stub}
	included, err := svc.validateIncludedPersonIDs(context.Background(), models.VisualisationTypeFull, treeID, []string{personID.String(), personID.String()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(included) != 1 {
		t.Fatalf("expected 1 unique person id, got %d", len(included))
	}
}

func TestBuildFileNameAndObjectKey(t *testing.T) {
	visID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	treeID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	name := buildFileName(models.VisualisationTypeAncestorsAndDescendants, visID)
	if name != "ancestors-and-descendants-11111111-1111-1111-1111-111111111111.svg" {
		t.Fatalf("unexpected file name %q", name)
	}

	key := buildObjectKey(treeID, visID)
	expected := "visualisations/22222222-2222-2222-2222-222222222222/11111111-1111-1111-1111-111111111111.svg"
	if key != expected {
		t.Fatalf("unexpected object key %q", key)
	}
}
