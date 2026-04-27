package stage4_render

import (
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage3_ordering"
)

type Coord struct {
	X int
	Y int
}

type Edge struct {
	From     Coord
	To       Coord
	EdgeType string
}

type NodeInfo struct {
	Left            int
	Right           int
	Layer           int
	People          []*stage1_input.Person
	MergePartnerIdx int
	AddedLeft       bool
}

type EdgeInfo struct {
	FromNodeIdx int
	ToNodeIdx   int
	FromX       int
	FromY       int
	ToX         int
	ToY         int
	EdgeType    string

	ParentsAdjacent bool
	AdjacentCenterX int
	ParentAddedLeft bool
}

type RenderResult struct {
	Coords map[int]Coord
	Edges  []Edge
}

type CoordRenderResult struct {
	Nodes    []NodeInfo
	Edges    []EdgeInfo
	MinLayer int
	MaxLayer int
	MaxRight int
}

type PersonPosition struct {
	Person *stage1_input.Person
	X      int
	Y      int
}

type PeopleGrid struct {
	Layers    map[int][]PersonPosition
	Positions []PersonPosition
}

func isInParentNodes(childNode, parentNode *stage3_ordering.CoordNode) bool {

	parentIDs := make(map[int]bool)
	for _, person := range parentNode.People {
		parentIDs[person.ID] = true
	}

	if parentNode.MergePartner != nil {
		for _, person := range parentNode.MergePartner.People {
			parentIDs[person.ID] = true
		}
	}

	for _, pn := range childNode.ParentNodes {
		if pn == nil {
			continue
		}
		for _, person := range pn.People {
			if parentIDs[person.ID] {
				return true
			}
		}

		if pn.MergePartner != nil {
			for _, person := range pn.MergePartner.People {
				if parentIDs[person.ID] {
					return true
				}
			}
		}
	}
	return false
}

func addParentChildEdge(result *CoordRenderResult, parentNode, childNode *stage3_ordering.CoordNode, parentIdx, childIdx int, addedEdges *map[string]bool) {

	fromX := (parentNode.Left + parentNode.Right) / 2

	toX := (childNode.Left + childNode.Right) / 2

	parentsAdjacent := false
	adjacentCenterX := 0
	if parentNode.MergePartner != nil {
		parentsAdjacent = true

		left := parentNode.Left
		right := parentNode.Right
		if parentNode.MergePartner.Left < left {
			left = parentNode.MergePartner.Left
		}
		if parentNode.MergePartner.Right > right {
			right = parentNode.MergePartner.Right
		}
		adjacentCenterX = (left + right) / 2
		fromX = adjacentCenterX
	}

	result.Edges = append(result.Edges, EdgeInfo{
		FromNodeIdx:     parentIdx,
		ToNodeIdx:       childIdx,
		FromX:           fromX,
		FromY:           parentNode.Layer,
		ToX:             toX,
		ToY:             childNode.Layer,
		EdgeType:        "parent-child",
		ParentsAdjacent: parentsAdjacent,
		AdjacentCenterX: adjacentCenterX,
		ParentAddedLeft: parentNode.AddedLeft,
	})
}

func getChildXCoord(node *stage3_ordering.CoordNode, personID int) int {
	if len(node.People) == 1 {
		return (node.Left + node.Right) / 2
	}

	for i, p := range node.People {
		if p.ID == personID {
			if i == 0 {
				return node.Left + 1
			}
			return node.Right - 1
		}
	}

	return (node.Left + node.Right) / 2
}

func getPersonXCoord(node *stage3_ordering.CoordNode, personIndex int) int {
	if len(node.People) == 1 {
		return (node.Left + node.Right) / 2
	}

	if personIndex == 0 {
		return node.Left + 1
	}
	return node.Right - 1
}

func getPersonXCoordByIndex(node *stage3_ordering.CoordNode, index int) int {
	if len(node.People) == 1 {
		return (node.Left + node.Right) / 2
	}

	if index == 0 {
		return node.Left + 1
	}
	return node.Right - 1
}

func findPersonIndexInNode(node *stage3_ordering.CoordNode, personID int) int {
	for i, p := range node.People {
		if p.ID == personID {
			return i
		}
	}
	return 0
}

func containsSamePerson(people1, people2 []*stage1_input.Person) bool {
	for _, p1 := range people1 {
		for _, p2 := range people2 {
			if p1.ID == p2.ID {
				return true
			}
		}
	}
	return false
}
