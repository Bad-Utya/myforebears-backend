package stage4_render

import (
	"fmt"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage3_ordering"
)

func BuildCoordRenderResult(cm *stage3_ordering.CoordMatrix, tree *stage1_input.FamilyTree) *CoordRenderResult {
	result := &CoordRenderResult{
		Nodes:    []NodeInfo{},
		Edges:    []EdgeInfo{},
		MinLayer: cm.MinLayer,
		MaxLayer: cm.MaxLayer,
		MaxRight: 0,
	}

	nodeToIndex := make(map[*stage3_ordering.CoordNode]int)

	personToNode := make(map[int]*stage3_ordering.CoordNode)

	for layerNum := cm.MinLayer; layerNum <= cm.MaxLayer; layerNum++ {
		for _, node := range cm.Layers[layerNum] {
			if node.IsPseudo {
				continue
			}

			idx := len(result.Nodes)
			nodeToIndex[node] = idx

			for _, person := range node.People {
				personToNode[person.ID] = node
			}

			info := NodeInfo{
				Left:            node.Left,
				Right:           node.Right,
				Layer:           layerNum,
				People:          node.People,
				MergePartnerIdx: -1,
				AddedLeft:       node.AddedLeft,
			}
			result.Nodes = append(result.Nodes, info)

			if node.Right > result.MaxRight {
				result.MaxRight = node.Right
			}
		}
	}

	for layerNum := cm.MinLayer; layerNum <= cm.MaxLayer; layerNum++ {
		for _, node := range cm.Layers[layerNum] {
			if node.IsPseudo {
				continue
			}
			nodeIdx, ok := nodeToIndex[node]
			if !ok {
				continue
			}
			if node.MergePartner != nil {
				partnerIdx, ok := nodeToIndex[node.MergePartner]
				if ok {
					result.Nodes[nodeIdx].MergePartnerIdx = partnerIdx
				}
			}
		}
	}

	addedEdges := make(map[string]bool)

	for _, person := range tree.People {
		childNode := personToNode[person.ID]
		if childNode == nil {
			continue
		}

		childIdx, ok := nodeToIndex[childNode]
		if !ok {
			continue
		}

		if person.Mother != nil {
			motherNode := personToNode[person.Mother.ID]
			if motherNode != nil {
				motherIdx, ok := nodeToIndex[motherNode]
				if ok {

					isParent := isInParentNodes(childNode, motherNode)
					if isParent {

						keyIdx := motherIdx
						if motherNode.MergePartner != nil {
							partnerIdx, ok := nodeToIndex[motherNode.MergePartner]
							if ok && partnerIdx < keyIdx {
								keyIdx = partnerIdx
							}
						}
						key := fmt.Sprintf("pc-%d-%d", keyIdx, childIdx)
						if !addedEdges[key] {
							addedEdges[key] = true
							addParentChildEdge(result, motherNode, childNode, motherIdx, childIdx, &addedEdges)
						}
					}
				}
			}
		}

		if person.Father != nil {
			fatherNode := personToNode[person.Father.ID]
			if fatherNode != nil {
				fatherIdx, ok := nodeToIndex[fatherNode]
				if ok {

					isParent := isInParentNodes(childNode, fatherNode)
					if isParent {

						keyIdx := fatherIdx
						if fatherNode.MergePartner != nil {
							partnerIdx, ok := nodeToIndex[fatherNode.MergePartner]
							if ok && partnerIdx < keyIdx {
								keyIdx = partnerIdx
							}
						}
						key := fmt.Sprintf("pc-%d-%d", keyIdx, childIdx)
						if !addedEdges[key] {
							addedEdges[key] = true
							addParentChildEdge(result, fatherNode, childNode, fatherIdx, childIdx, &addedEdges)
						}
					}
				}
			}
		}
	}

	for layerNum := cm.MinLayer; layerNum <= cm.MaxLayer; layerNum++ {
		for _, node := range cm.Layers[layerNum] {
			if node.IsPseudo {
				continue
			}

			nodeIdx, ok := nodeToIndex[node]
			if !ok {
				continue
			}

			if len(node.People) == 2 {
				key := fmt.Sprintf("partner-%d", nodeIdx)
				if !addedEdges[key] {
					addedEdges[key] = true
					result.Edges = append(result.Edges, EdgeInfo{
						FromNodeIdx: nodeIdx,
						ToNodeIdx:   nodeIdx,
						FromX:       node.Left + 1,
						FromY:       node.Layer,
						ToX:         node.Right - 1,
						ToY:         node.Layer,
						EdgeType:    "partner",
					})
				}
			}

			if node.MergePartner != nil {
				partnerIdx, ok := nodeToIndex[node.MergePartner]
				if ok {
					if nodeIdx > partnerIdx {
						continue
					}

					if node.Right <= node.MergePartner.Left {
						key := fmt.Sprintf("merge-partner-%d-%d", nodeIdx, partnerIdx)
						if !addedEdges[key] {
							addedEdges[key] = true

							fromCenter := (node.Left + node.Right) / 2
							toCenter := (node.MergePartner.Left + node.MergePartner.Right) / 2
							result.Edges = append(result.Edges, EdgeInfo{
								FromNodeIdx: nodeIdx,
								ToNodeIdx:   partnerIdx,
								FromX:       fromCenter,
								FromY:       node.Layer,
								ToX:         toCenter,
								ToY:         node.Layer,
								EdgeType:    "partner",
							})
						}
					}
				}
			}
		}
	}

	for layerNum := cm.MinLayer; layerNum <= cm.MaxLayer; layerNum++ {
		for _, childNode := range cm.Layers[layerNum] {
			if childNode.IsPseudo {
				continue
			}

			if len(childNode.ParentNodes) >= 2 && childNode.ParentNodes[0] != nil && childNode.ParentNodes[1] != nil {
				parent1 := childNode.ParentNodes[0]
				parent2 := childNode.ParentNodes[1]

				if parent1 == parent2 {
					continue
				}

				if parent1.MergePartner == parent2 {
					continue
				}

				parent1Idx, ok1 := nodeToIndex[parent1]
				parent2Idx, ok2 := nodeToIndex[parent2]
				if !ok1 || !ok2 {
					continue
				}

				var leftIdx, rightIdx int
				var leftNode, rightNode *stage3_ordering.CoordNode
				if parent1.Right <= parent2.Left {
					leftIdx, rightIdx = parent1Idx, parent2Idx
					leftNode, rightNode = parent1, parent2
				} else {
					leftIdx, rightIdx = parent2Idx, parent1Idx
					leftNode, rightNode = parent2, parent1
				}

				key := fmt.Sprintf("child-partner-%d-%d", leftIdx, rightIdx)
				if !addedEdges[key] {
					addedEdges[key] = true
					fromCenter := (leftNode.Left + leftNode.Right) / 2
					toCenter := (rightNode.Left + rightNode.Right) / 2
					result.Edges = append(result.Edges, EdgeInfo{
						FromNodeIdx: leftIdx,
						ToNodeIdx:   rightIdx,
						FromX:       fromCenter,
						FromY:       leftNode.Layer,
						ToX:         toCenter,
						ToY:         leftNode.Layer,
						EdgeType:    "partner",
					})
				}
			}
		}
	}

	for _, person := range tree.People {
		personNode := personToNode[person.ID]
		if personNode == nil {
			continue
		}

		personIdx, ok := nodeToIndex[personNode]
		if !ok {
			continue
		}

		for _, partner := range person.Partners {
			partnerNode := personToNode[partner.ID]
			if partnerNode == nil {
				continue
			}

			if partnerNode == personNode {
				continue
			}

			if personNode.MergePartner == partnerNode {
				continue
			}

			partnerIdx, ok := nodeToIndex[partnerNode]
			if !ok {
				continue
			}

			if person.ID > partner.ID {
				continue
			}

			var leftIdx, rightIdx int
			var leftNode, rightNode *stage3_ordering.CoordNode
			if personNode.Right <= partnerNode.Left {
				leftIdx, rightIdx = personIdx, partnerIdx
				leftNode, rightNode = personNode, partnerNode
			} else {
				leftIdx, rightIdx = partnerIdx, personIdx
				leftNode, rightNode = partnerNode, personNode
			}

			key := fmt.Sprintf("family-partner-%d-%d", leftIdx, rightIdx)
			if !addedEdges[key] {
				addedEdges[key] = true
				fromCenter := (leftNode.Left + leftNode.Right) / 2
				toCenter := (rightNode.Left + rightNode.Right) / 2
				result.Edges = append(result.Edges, EdgeInfo{
					FromNodeIdx: leftIdx,
					ToNodeIdx:   rightIdx,
					FromX:       fromCenter,
					FromY:       leftNode.Layer,
					ToX:         toCenter,
					ToY:         leftNode.Layer,
					EdgeType:    "partner",
				})
			}
		}
	}

	return result
}

func BuildRenderResult(grid *PeopleGrid, tree *stage1_input.FamilyTree) *RenderResult {
	result := &RenderResult{
		Coords: make(map[int]Coord),
		Edges:  []Edge{},
	}

	for _, pos := range grid.Positions {
		result.Coords[pos.Person.ID] = Coord{
			X: pos.X,
			Y: pos.Y,
		}
	}

	for _, person := range tree.People {
		if person.Layout == nil {
			continue
		}
		personCoord, ok := result.Coords[person.ID]
		if !ok {
			continue
		}

		if person.Mother != nil {
			if motherCoord, ok := result.Coords[person.Mother.ID]; ok {
				result.Edges = append(result.Edges, Edge{
					From:     motherCoord,
					To:       personCoord,
					EdgeType: "parent-child",
				})
			}
		}

		if person.Father != nil {
			if fatherCoord, ok := result.Coords[person.Father.ID]; ok {
				result.Edges = append(result.Edges, Edge{
					From:     fatherCoord,
					To:       personCoord,
					EdgeType: "parent-child",
				})
			}
		}
	}

	addedPartnerEdges := make(map[string]bool)
	for _, person := range tree.People {
		if person.Layout == nil {
			continue
		}
		personCoord, ok := result.Coords[person.ID]
		if !ok {
			continue
		}

		for _, partner := range person.Partners {
			if partner.Layout == nil {
				continue
			}
			partnerCoord, ok := result.Coords[partner.ID]
			if !ok {
				continue
			}

			if personCoord.Y != partnerCoord.Y {
				continue
			}

			if personCoord.X == partnerCoord.X {
				continue
			}

			var key string
			if person.ID < partner.ID {
				key = fmt.Sprintf("%d-%d", person.ID, partner.ID)
			} else {
				key = fmt.Sprintf("%d-%d", partner.ID, person.ID)
			}

			if !addedPartnerEdges[key] {
				addedPartnerEdges[key] = true
				result.Edges = append(result.Edges, Edge{
					From:     personCoord,
					To:       partnerCoord,
					EdgeType: "partner",
				})
			}
		}
	}

	return result
}

func (r *RenderResult) Print(tree *stage1_input.FamilyTree) {
	fmt.Println("\n=== РљРѕРѕСЂРґРёРЅР°С‚С‹ РІРµСЂС€РёРЅ ===")
	for id, coord := range r.Coords {
		person := tree.People[id]
		fmt.Printf("%s (ID=%d): X=%d, Y=%d\n", person.Name, id, coord.X, coord.Y)
	}

	fmt.Println("\n=== РЎРІСЏР·Рё ===")
	for _, edge := range r.Edges {
		fmt.Printf("(%d,%d) -> (%d,%d) [%s]\n", edge.From.X, edge.From.Y, edge.To.X, edge.To.Y, edge.EdgeType)
	}
}
