package stage4_render

import (
	"strings"
	"testing"

	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage3_ordering"
)

func TestBuildCoordRenderResultRoutesSeparatedPartnersAboveHighestAncestor(t *testing.T) {
	left := stage1_input.NewPerson(1, "Left", stage1_input.Male, "left", "tree")
	middle := stage1_input.NewPerson(2, "Middle", stage1_input.Male, "middle", "tree")
	parent := stage1_input.NewPerson(3, "Parent", stage1_input.Female, "parent", "tree")
	right := stage1_input.NewPerson(4, "Right", stage1_input.Female, "right", "tree")
	top := stage1_input.NewPerson(5, "Top", stage1_input.Male, "top", "tree")

	tree := stage1_input.NewFamilyTree()
	for _, person := range []*stage1_input.Person{left, middle, parent, right, top} {
		tree.AddPerson(person)
	}
	if err := tree.AddPartnership(left.ID, right.ID); err != nil {
		t.Fatalf("AddPartnership() error = %v", err)
	}

	leftNode := coordNode(left, 0, 0, 2)
	middleNode := coordNode(middle, 0, 4, 6)
	rightNode := coordNode(right, 0, 8, 10)
	parentNode := coordNode(parent, 1, 4, 6)
	topNode := coordNode(top, 2, 4, 6)
	middleNode.Up = []*stage3_ordering.CoordNode{parentNode}
	parentNode.Up = []*stage3_ordering.CoordNode{topNode}
	// A malformed cycle must not make the BFS loop forever.
	topNode.Up = []*stage3_ordering.CoordNode{middleNode}

	cm := stage3_ordering.NewCoordMatrix(0, 2)
	for _, node := range []*stage3_ordering.CoordNode{leftNode, middleNode, rightNode, parentNode, topNode} {
		cm.AddNode(node)
	}

	result := BuildCoordRenderResult(cm, tree)
	var partnerEdge *EdgeInfo
	for i := range result.Edges {
		if result.Edges[i].EdgeType == "partner" && result.Edges[i].FromNodeIdx != result.Edges[i].ToNodeIdx {
			partnerEdge = &result.Edges[i]
			break
		}
	}
	if partnerEdge == nil {
		t.Fatal("separated partner edge was not built")
	}
	if !partnerEdge.RouteAbove {
		t.Fatal("separated partner edge must be routed above intermediate nodes")
	}
	if partnerEdge.RouteAboveLayer != 2 {
		t.Fatalf("RouteAboveLayer = %d, want 2", partnerEdge.RouteAboveLayer)
	}

	svg := NewCoordSVGRenderer(result, tree, nil).renderEdges()
	wantPath := `d="M 185 460 L 227 460 L 227 60 L 823 60 L 823 460 L 865 460"`
	if !strings.Contains(svg, wantPath) {
		t.Fatalf("routed partner path not found; want %s in %s", wantPath, svg)
	}
}

func TestHighestIntermediateUpLayerRoutesAdjacentPartnersAboveTheirLayer(t *testing.T) {
	left := coordNode(stage1_input.NewPerson(1, "Left", stage1_input.Male, "", ""), 0, 0, 2)
	right := coordNode(stage1_input.NewPerson(2, "Right", stage1_input.Female, "", ""), 0, 4, 6)
	cm := stage3_ordering.NewCoordMatrix(0, 0)
	cm.AddNode(left)
	cm.AddNode(right)

	if layer, ok := highestIntermediateUpLayer(cm, left, right); !ok || layer != 0 {
		t.Fatalf("highestIntermediateUpLayer() = (%d, %v), want (0, true)", layer, ok)
	}
}

func TestCoordSVGRendererKeepsDoubleNodePartnerEdgeStraight(t *testing.T) {
	left := stage1_input.NewPerson(1, "Left", stage1_input.Male, "", "")
	right := stage1_input.NewPerson(2, "Right", stage1_input.Female, "", "")
	result := &CoordRenderResult{
		Nodes:    []NodeInfo{{Left: 0, Right: 4, Layer: 0, People: []*stage1_input.Person{left, right}}},
		Edges:    []EdgeInfo{{FromNodeIdx: 0, ToNodeIdx: 0, FromY: 0, ToY: 0, EdgeType: "partner"}},
		MinLayer: 0,
		MaxLayer: 0,
		MaxRight: 4,
	}

	svg := NewCoordSVGRenderer(result, stage1_input.NewFamilyTree(), nil).renderEdges()
	if !strings.Contains(svg, `<line `) || strings.Contains(svg, `<path `) {
		t.Fatalf("double-node partner edge must remain straight, got %s", svg)
	}
}

func coordNode(person *stage1_input.Person, layer, left, right int) *stage3_ordering.CoordNode {
	return &stage3_ordering.CoordNode{
		Left:   left,
		Right:  right,
		Layer:  layer,
		People: []*stage1_input.Person{person},
		Up:     make([]*stage3_ordering.CoordNode, 0),
	}
}
