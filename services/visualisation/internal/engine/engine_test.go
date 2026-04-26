package engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	familytreepb "github.com/Bad-Utya/myforebears-backend/gen/go/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage2_layout"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage3_ordering"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage4_render"
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

func TestRenderSVGFullPrunesNonIncludedPeople(t *testing.T) {
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

func TestRenderCoordinatesWithDepthLimit(t *testing.T) {
	rootID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	partnerID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	childID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	grandchildID := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	content := &familytreepb.GetTreeContentResponse{
		Persons: []*familytreepb.Person{
			{Id: rootID.String(), FirstName: "Root", Gender: familytreepb.Gender_GENDER_MALE},
			{Id: partnerID.String(), FirstName: "Partner", Gender: familytreepb.Gender_GENDER_FEMALE},
			{Id: childID.String(), FirstName: "Child", Gender: familytreepb.Gender_GENDER_MALE},
			{Id: grandchildID.String(), FirstName: "Grandchild", Gender: familytreepb.Gender_GENDER_FEMALE},
		},
		Relationships: []*familytreepb.Relationship{
			{PersonIdFrom: rootID.String(), PersonIdTo: partnerID.String(), Type: familytreepb.RelationshipType_RELATIONSHIP_PARTNER_MARRIED},
			{PersonIdFrom: rootID.String(), PersonIdTo: childID.String(), Type: familytreepb.RelationshipType_RELATIONSHIP_PARENT_CHILD},
			{PersonIdFrom: partnerID.String(), PersonIdTo: childID.String(), Type: familytreepb.RelationshipType_RELATIONSHIP_PARENT_CHILD},
			{PersonIdFrom: childID.String(), PersonIdTo: grandchildID.String(), Type: familytreepb.RelationshipType_RELATIONSHIP_PARENT_CHILD},
		},
	}

	coordBytes, err := RenderCoordinates(models.VisualisationTypeAncestorsAndDescendants, rootID, nil, content, 0, true)
	if err != nil {
		t.Fatalf("RenderCoordinates with maxDepth=0 failed: %v", err)
	}
	if len(coordBytes) == 0 {
		t.Fatal("expected non-empty coordinates")
	}

	coordBytes, err = RenderCoordinates(models.VisualisationTypeAncestorsAndDescendants, rootID, nil, content, 1, true)
	if err != nil {
		t.Fatalf("RenderCoordinates with maxDepth=1 failed: %v", err)
	}
	if len(coordBytes) == 0 {
		t.Fatal("expected non-empty coordinates with depth limit")
	}
}

func TestRenderCoordinatesBuildsJSON(t *testing.T) {
	rootID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	partnerID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	content := &familytreepb.GetTreeContentResponse{
		Persons: []*familytreepb.Person{
			{Id: rootID.String(), FirstName: "Root", Gender: familytreepb.Gender_GENDER_MALE},
			{Id: partnerID.String(), FirstName: "Partner", Gender: familytreepb.Gender_GENDER_FEMALE},
		},
		Relationships: []*familytreepb.Relationship{
			{PersonIdFrom: rootID.String(), PersonIdTo: partnerID.String(), Type: familytreepb.RelationshipType_RELATIONSHIP_PARTNER_MARRIED},
		},
	}

	coordBytes, err := RenderCoordinates(models.VisualisationTypeAncestorsAndDescendants, rootID, nil, content, 0, true)
	if err != nil {
		t.Fatalf("RenderCoordinates failed: %v", err)
	}

	if !bytes.HasPrefix(coordBytes, []byte("{")) {
		t.Fatalf("expected JSON output starting with {, got %q", string(coordBytes[:min(len(coordBytes), 16)]))
	}

	jsonStr := string(coordBytes)
	if !bytes.Contains(coordBytes, []byte("\"nodes\"")) {
		t.Fatalf("expected 'nodes' key in JSON output")
	}
	if !bytes.Contains(coordBytes, []byte("\"edges\"")) {
		t.Fatalf("expected 'edges' key in JSON output")
	}

	t.Logf("Coordinates JSON sample:\n%s", jsonStr[:min(len(jsonStr), 200)])
}

func TestRenderCoordinatesIncludesFullPeopleAndMatchingEdges(t *testing.T) {
	rootID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	partnerID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	childID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	content := &familytreepb.GetTreeContentResponse{
		Persons: []*familytreepb.Person{
			{Id: rootID.String(), TreeId: "tree-1", FirstName: "Root", LastName: "Person", Patronymic: "Sr.", Gender: familytreepb.Gender_GENDER_MALE, AvatarPhotoId: "avatar-root"},
			{Id: partnerID.String(), TreeId: "tree-1", FirstName: "Partner", LastName: "Person", Patronymic: "", Gender: familytreepb.Gender_GENDER_FEMALE, AvatarPhotoId: "avatar-partner"},
			{Id: childID.String(), TreeId: "tree-1", FirstName: "Child", LastName: "Person", Patronymic: "Jr.", Gender: familytreepb.Gender_GENDER_UNSPECIFIED},
		},
		Relationships: []*familytreepb.Relationship{
			{PersonIdFrom: rootID.String(), PersonIdTo: partnerID.String(), Type: familytreepb.RelationshipType_RELATIONSHIP_PARTNER_MARRIED},
			{PersonIdFrom: rootID.String(), PersonIdTo: childID.String(), Type: familytreepb.RelationshipType_RELATIONSHIP_PARENT_CHILD},
			{PersonIdFrom: partnerID.String(), PersonIdTo: childID.String(), Type: familytreepb.RelationshipType_RELATIONSHIP_PARENT_CHILD},
		},
	}

	people, relations, err := normalizeInput(content)
	if err != nil {
		t.Fatalf("normalizeInput failed: %v", err)
	}

	allowed := map[uuid.UUID]struct{}{}
	for id := range people {
		allowed[id] = struct{}{}
	}

	filteredPeople, filteredRelations := filterByVisualisation(models.VisualisationTypeFull, rootID, allowed, people, relations)
	tree, idToInt, internalPersons, err := buildStageTree(filteredPeople, filteredRelations)
	if err != nil {
		t.Fatalf("buildStageTree failed: %v", err)
	}
	rootIntID := idToInt[rootID]
	history, err := stage2_layout.LayoutFromPerson(tree, rootIntID)
	if err != nil {
		t.Fatalf("LayoutFromPerson failed: %v", err)
	}
	layouts := make(map[int]*stage1_input.PersonLayout)
	for id, person := range tree.People {
		if person.Layout != nil {
			layouts[id] = person.Layout
		}
	}
	start := tree.People[rootIntID]
	om := stage3_ordering.ProcessPlacementHistory(history, start, start.Layout.Layer, layouts)
	renderResult := stage4_render.BuildCoordRenderResult(om.BuildCoordMatrix(), tree)
	expectedJSON := renderResultToJSON(renderResult, internalPersons)

	var actual CoordinateResultJSON
	if err := json.Unmarshal(expectedJSON, &actual); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if len(actual.Nodes) == 0 {
		t.Fatal("expected at least one node in JSON")
	}

	var rootPerson *PersonJSON
	for i := range actual.Nodes {
		for j := range actual.Nodes[i].People {
			if actual.Nodes[i].People[j].ID == rootID.String() {
				rootPerson = &actual.Nodes[i].People[j]
				break
			}
		}
		if rootPerson != nil {
			break
		}
	}
	if rootPerson == nil {
		t.Fatal("expected to find root person in JSON output")
	}
	if rootPerson.TreeID != "tree-1" || rootPerson.FirstName != "Root" || rootPerson.LastName != "Person" || rootPerson.Patronymic != "Sr." {
		t.Fatalf("expected full person fields, got %+v", rootPerson)
	}
	if rootPerson.AvatarPhotoID != "avatar-root" {
		t.Fatalf("expected avatar id to be preserved, got %+v", rootPerson)
	}

	if len(actual.Edges) != len(renderResult.Edges) {
		t.Fatalf("expected %d edges, got %d", len(renderResult.Edges), len(actual.Edges))
	}
	for i := range actual.Edges {
		expected := renderResult.Edges[i]
		got := actual.Edges[i]
		if got.FromNodeIdx != expected.FromNodeIdx || got.ToNodeIdx != expected.ToNodeIdx || got.FromX != expected.FromX || got.FromY != expected.FromY || got.ToX != expected.ToX || got.ToY != expected.ToY || got.EdgeType != expected.EdgeType {
			t.Fatalf("edge %d mismatch: got %+v, want %+v", i, got, expected)
		}
	}
}
