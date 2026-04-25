package engine

import (
	"bytes"
	"errors"
	"testing"

	familytreepb "github.com/Bad-Utya/myforebears-backend/gen/go/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/domain/models"
	"github.com/google/uuid"
)

func TestRenderSVGBuildsDocument(t *testing.T) {
	rootID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	partnerID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	childID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	content := &familytreepb.GetTreeContentResponse{
		Persons: []*familytreepb.Person{
			{Id: rootID.String(), FirstName: "Root", Gender: familytreepb.Gender_GENDER_MALE},
			{Id: partnerID.String(), FirstName: "Partner", Gender: familytreepb.Gender_GENDER_FEMALE},
			{Id: childID.String(), FirstName: "Child", Gender: familytreepb.Gender_GENDER_UNSPECIFIED},
		},
		Relationships: []*familytreepb.Relationship{
			{PersonIdFrom: rootID.String(), PersonIdTo: partnerID.String(), Type: familytreepb.RelationshipType_RELATIONSHIP_PARTNER_MARRIED},
			{PersonIdFrom: rootID.String(), PersonIdTo: childID.String(), Type: familytreepb.RelationshipType_RELATIONSHIP_PARENT_CHILD},
			{PersonIdFrom: partnerID.String(), PersonIdTo: childID.String(), Type: familytreepb.RelationshipType_RELATIONSHIP_PARENT_CHILD},
		},
	}

	svg, err := RenderSVG(models.VisualisationTypeAncestorsAndDescendants, rootID, nil, content)
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}
	if len(svg) == 0 {
		t.Fatal("expected non-empty svg")
	}
	if !bytes.HasPrefix(svg, []byte("<?xml")) {
		t.Fatalf("expected SVG XML header, got %q", string(svg[:min(len(svg), 16)]))
	}
}

func TestRenderPDFFullPrunesNonIncludedPeople(t *testing.T) {
	rootID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	includedID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	excludedID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	people := map[uuid.UUID]personView{
		rootID:     {id: rootID, label: "Root"},
		includedID: {id: includedID, label: "Included"},
		excludedID: {id: excludedID, label: "Excluded"},
	}
	relations := []relationView{
		{from: rootID, to: includedID, rt: familytreepb.RelationshipType_RELATIONSHIP_PARTNER_MARRIED},
		{from: rootID, to: excludedID, rt: familytreepb.RelationshipType_RELATIONSHIP_PARTNER_MARRIED},
	}
	allowed := map[uuid.UUID]struct{}{
		rootID:     {},
		includedID: {},
	}

	filteredPeople, filteredRelations := filterByVisualisation(models.VisualisationTypeFull, rootID, allowed, people, relations)
	if len(filteredPeople) != 2 {
		t.Fatalf("expected 2 people after full filtering, got %d", len(filteredPeople))
	}
	if _, ok := filteredPeople[excludedID]; ok {
		t.Fatalf("excluded person %s must be removed", excludedID)
	}
	if len(filteredRelations) != 1 {
		t.Fatalf("expected 1 relation after full filtering, got %d", len(filteredRelations))
	}
}

func TestRenderSVGRootNotFound(t *testing.T) {
	rootID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	content := &familytreepb.GetTreeContentResponse{
		Persons: []*familytreepb.Person{{
			Id:        "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
			FirstName: "Someone",
		}},
	}

	_, err := RenderSVG(models.VisualisationTypeAncestorsAndDescendants, rootID, nil, content)
	if err == nil {
		t.Fatal("expected root not found error")
	}
	if !errors.Is(err, errRootNotFound) {
		t.Fatalf("expected errRootNotFound, got %v", err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
