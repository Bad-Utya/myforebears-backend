package layout

import (
	"bytes"
	"testing"

	"github.com/Bad-Utya/myforebears-backend/services/customtree/internal/domain"
	"github.com/google/uuid"
)

func TestBuildPlacesParentAboveAndChildrenBelowSelectedRoot(t *testing.T) {
	parent, root, child := uuid.New(), uuid.New(), uuid.New()
	entities := []domain.Entity{{ID: parent, Name: "Supervisor"}, {ID: root, Name: "Student"}, {ID: child, Name: "Junior"}}
	edges := []domain.Edge{{ParentID: parent, ChildID: root}, {ParentID: root, ChildID: child}}
	result := Build(root, entities, edges, "student", "supervisor")
	layers := map[uuid.UUID]int{}
	for _, node := range result.Nodes {
		layers[node.EntityID] = node.Layer
	}
	if layers[parent] != -1 || layers[root] != 0 || layers[child] != 1 {
		t.Fatalf("unexpected layers: %#v", layers)
	}
	if len(result.Edges) != 2 || result.Edges[0].LabelDown != "student" {
		t.Fatalf("relation labels were not preserved")
	}
}

func TestBuildIsDeterministic(t *testing.T) {
	root := uuid.New()
	entities := []domain.Entity{{ID: root, Name: "root"}}
	edges := []domain.Edge{}
	for i := 0; i < 8; i++ {
		id := uuid.New()
		entities = append(entities, domain.Entity{ID: id, Name: id.String()})
		edges = append(edges, domain.Edge{ParentID: root, ChildID: id})
	}
	a := Build(root, entities, edges, "child", "parent")
	b := Build(root, entities, edges, "child", "parent")
	if !bytes.Equal(a.SVG(), b.SVG()) {
		t.Fatal("layout must be deterministic")
	}
}

func TestSVGContainsEscapedEntityName(t *testing.T) {
	id := uuid.New()
	result := Build(id, []domain.Entity{{ID: id, Name: "A < B"}}, nil, "child", "parent")
	if bytes.Contains(result.SVG(), []byte("A < B")) || !bytes.Contains(result.SVG(), []byte("A &lt; B")) {
		t.Fatal("entity name was not escaped")
	}
}
