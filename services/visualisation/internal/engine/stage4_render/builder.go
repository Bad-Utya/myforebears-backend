package stage4_render

import (
	"fmt"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage3_ordering"
)

// BuildCoordRenderResult СЃС‚СЂРѕРёС‚ СЂРµР·СѓР»СЊС‚Р°С‚ РІРёР·СѓР°Р»РёР·Р°С†РёРё РёР· CoordMatrix
func BuildCoordRenderResult(cm *stage3_ordering.CoordMatrix, tree *stage1_input.FamilyTree) *CoordRenderResult {
	result := &CoordRenderResult{
		Nodes:    []NodeInfo{},
		Edges:    []EdgeInfo{},
		MinLayer: cm.MinLayer,
		MaxLayer: cm.MaxLayer,
		MaxRight: 0,
	}

	// РљР°СЂС‚Р° РґР»СЏ РїРѕРёСЃРєР° РёРЅРґРµРєСЃР° СѓР·Р»Р° РїРѕ СѓРєР°Р·Р°С‚РµР»СЋ
	nodeToIndex := make(map[*stage3_ordering.CoordNode]int)

	// РљР°СЂС‚Р° РґР»СЏ РїРѕРёСЃРєР° РІРµСЂС€РёРЅС‹ РїРѕ ID С‡РµР»РѕРІРµРєР°
	personToNode := make(map[int]*stage3_ordering.CoordNode)

	// РЎРѕР±РёСЂР°РµРј РІСЃРµ РІРµСЂС€РёРЅС‹ РёР· СЃР»РѕС‘РІ (СЌС‚Рѕ С„РёРЅР°Р»СЊРЅС‹Рµ РІРµСЂС€РёРЅС‹ РїРѕСЃР»Рµ split)
	for layerNum := cm.MinLayer; layerNum <= cm.MaxLayer; layerNum++ {
		for _, node := range cm.Layers[layerNum] {
			if node.IsPseudo {
				continue // РџСЃРµРІРґРѕРІРµСЂС€РёРЅС‹ РЅРµ СЂРёСЃСѓРµРј
			}

			idx := len(result.Nodes)
			nodeToIndex[node] = idx

			// Р РµРіРёСЃС‚СЂРёСЂСѓРµРј РІСЃРµС… Р»СЋРґРµР№ СЌС‚РѕР№ РІРµСЂС€РёРЅС‹
			for _, person := range node.People {
				personToNode[person.ID] = node
			}

			info := NodeInfo{
				Left:            node.Left,
				Right:           node.Right,
				Layer:           layerNum,
				People:          node.People,
				MergePartnerIdx: -1, // СѓСЃС‚Р°РЅРѕРІРёРј РїРѕР·Р¶Рµ
				AddedLeft:       node.AddedLeft,
			}
			result.Nodes = append(result.Nodes, info)

			if node.Right > result.MaxRight {
				result.MaxRight = node.Right
			}
		}
	}

	// Р’С‚РѕСЂРѕР№ РїСЂРѕС…РѕРґ: Р·Р°РїРѕР»РЅСЏРµРј MergePartnerIdx
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

	// РЎРѕР±РёСЂР°РµРј СЃРІСЏР·Рё СЂРѕРґРёС‚РµР»СЊ-СЂРµР±С‘РЅРѕРє С‡РµСЂРµР· СЃС‚СЂСѓРєС‚СѓСЂСѓ FamilyTree
	addedEdges := make(map[string]bool)

	// РЎРѕР±РёСЂР°РµРј СЃРІСЏР·Рё СЂРѕРґРёС‚РµР»СЊ-СЂРµР±С‘РЅРѕРє С‡РµСЂРµР· FamilyTree
	// РЅРѕ С‚РѕР»СЊРєРѕ РµСЃР»Рё СЂРѕРґРёС‚РµР»СЊ СЏРІР»СЏРµС‚СЃСЏ ParentNodes СЂРµР±С‘РЅРєР° (РґР°РЅРЅС‹Рµ С‚СЂРµС‚СЊРµРіРѕ СЌС‚Р°РїР°)
	for _, person := range tree.People {
		childNode := personToNode[person.ID]
		if childNode == nil {
			continue
		}

		childIdx, ok := nodeToIndex[childNode]
		if !ok {
			continue
		}

		// РџСЂРѕРІРµСЂСЏРµРј СЃРІСЏР·СЊ СЃ РјР°С‚РµСЂСЊСЋ
		if person.Mother != nil {
			motherNode := personToNode[person.Mother.ID]
			if motherNode != nil {
				motherIdx, ok := nodeToIndex[motherNode]
				if ok {
					// РџСЂРѕРІРµСЂСЏРµРј, СЏРІР»СЏРµС‚СЃСЏ Р»Рё РјР°С‚СЊ ParentNodes СЂРµР±С‘РЅРєР°
					isParent := isInParentNodes(childNode, motherNode)
					if isParent {
						// РљР»СЋС‡ РґРѕР»Р¶РµРЅ СѓС‡РёС‚С‹РІР°С‚СЊ MergePartner С‡С‚РѕР±С‹ РёР·Р±РµР¶Р°С‚СЊ РґСѓР±Р»РёРєР°С‚РѕРІ
						// РСЃРїРѕР»СЊР·СѓРµРј РјРµРЅСЊС€РёР№ РёРЅРґРµРєСЃ РёР· РїР°СЂС‹ РєР°Рє РїРµСЂРІС‹Р№ РєР»СЋС‡
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

		// РџСЂРѕРІРµСЂСЏРµРј СЃРІСЏР·СЊ СЃ РѕС‚С†РѕРј
		if person.Father != nil {
			fatherNode := personToNode[person.Father.ID]
			if fatherNode != nil {
				fatherIdx, ok := nodeToIndex[fatherNode]
				if ok {
					// РџСЂРѕРІРµСЂСЏРµРј, СЏРІР»СЏРµС‚СЃСЏ Р»Рё РѕС‚РµС† ParentNodes СЂРµР±С‘РЅРєР°
					isParent := isInParentNodes(childNode, fatherNode)
					if isParent {
						// РљР»СЋС‡ РґРѕР»Р¶РµРЅ СѓС‡РёС‚С‹РІР°С‚СЊ MergePartner С‡С‚РѕР±С‹ РёР·Р±РµР¶Р°С‚СЊ РґСѓР±Р»РёРєР°С‚РѕРІ
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

	// Р”РѕР±Р°РІР»СЏРµРј СЃРІСЏР·Рё РїР°СЂС‚РЅС‘СЂРѕРІ
	// 1. РњРµР¶РґСѓ СЃРјРµР¶РЅС‹РјРё РІРµСЂС€РёРЅР°РјРё (MergePartner)
	// 2. РњРµР¶РґСѓ РІРµСЂС€РёРЅР°РјРё, РєРѕС‚РѕСЂС‹Рµ СЏРІР»СЏСЋС‚СЃСЏ СЂРѕРґРёС‚РµР»СЏРјРё РѕРґРЅРѕРіРѕ СЂРµР±С‘РЅРєР° (С‡РµСЂРµР· ParentNodes)
	for layerNum := cm.MinLayer; layerNum <= cm.MaxLayer; layerNum++ {
		for _, node := range cm.Layers[layerNum] {
			if node.IsPseudo {
				continue
			}

			nodeIdx, ok := nodeToIndex[node]
			if !ok {
				continue
			}

			// РЎРІСЏР·Рё РїР°СЂС‚РЅС‘СЂРѕРІ (РµСЃР»Рё 2 С‡РµР»РѕРІРµРєР° РІ РІРµСЂС€РёРЅРµ) - СЂРёСЃСѓРµРј Р»РёРЅРёСЋ РІРЅСѓС‚СЂРё
			if len(node.People) == 2 {
				key := fmt.Sprintf("partner-%d", nodeIdx)
				if !addedEdges[key] {
					addedEdges[key] = true
					result.Edges = append(result.Edges, EdgeInfo{
						FromNodeIdx: nodeIdx,
						ToNodeIdx:   nodeIdx, // СЃР°Рј РЅР° СЃРµР±СЏ - Р»РёРЅРёСЏ РІРЅСѓС‚СЂРё РІРµСЂС€РёРЅС‹
						FromX:       node.Left + 1,
						FromY:       node.Layer,
						ToX:         node.Right - 1,
						ToY:         node.Layer,
						EdgeType:    "partner",
					})
				}
			}

			// РЎРІСЏР·Рё РјРµР¶РґСѓ СЃРјРµР¶РЅС‹РјРё РІРµСЂС€РёРЅР°РјРё (MergePartner) - РєСЂР°СЃРЅР°СЏ Р»РёРЅРёСЏ РјРµР¶РґСѓ РЅРёРјРё
			if node.MergePartner != nil {
				partnerIdx, ok := nodeToIndex[node.MergePartner]
				if ok {
					// Р”РѕР±Р°РІР»СЏРµРј С‚РѕР»СЊРєРѕ РѕРґРёРЅ СЂР°Р· (РѕС‚ Р»РµРІРѕР№ Рє РїСЂР°РІРѕР№ РІРµСЂС€РёРЅРµ)
					if node.Right <= node.MergePartner.Left {
						key := fmt.Sprintf("merge-partner-%d-%d", nodeIdx, partnerIdx)
						if !addedEdges[key] {
							addedEdges[key] = true
							// Р›РёРЅРёСЏ РѕС‚ С†РµРЅС‚СЂР° РїРµСЂРІРѕР№ РІРµСЂС€РёРЅС‹ РґРѕ С†РµРЅС‚СЂР° РІС‚РѕСЂРѕР№
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

	// РС‰РµРј РїР°СЂС‚РЅС‘СЂРѕРІ С‡РµСЂРµР· РґРµС‚РµР№ (Сѓ СЂРµР±С‘РЅРєР° РјРѕР¶РµС‚ Р±С‹С‚СЊ 2 СЂР°Р·РЅС‹С… ParentNodes = РїР°СЂС‚РЅС‘СЂС‹)
	for layerNum := cm.MinLayer; layerNum <= cm.MaxLayer; layerNum++ {
		for _, childNode := range cm.Layers[layerNum] {
			if childNode.IsPseudo {
				continue
			}

			// Р•СЃР»Рё Сѓ СЂРµР±С‘РЅРєР° РµСЃС‚СЊ 2 СЂР°Р·РЅС‹С… СЂРѕРґРёС‚РµР»СЏ
			if len(childNode.ParentNodes) >= 2 && childNode.ParentNodes[0] != nil && childNode.ParentNodes[1] != nil {
				parent1 := childNode.ParentNodes[0]
				parent2 := childNode.ParentNodes[1]

				// РџСЂРѕРїСѓСЃРєР°РµРј РµСЃР»Рё СЌС‚Рѕ РѕРґРЅР° Рё С‚Р° Р¶Рµ РІРµСЂС€РёРЅР°
				if parent1 == parent2 {
					continue
				}

				// РџСЂРѕРїСѓСЃРєР°РµРј РµСЃР»Рё СЌС‚Рѕ MergePartner (СѓР¶Рµ РґРѕР±Р°РІРёР»Рё РІС‹С€Рµ)
				if parent1.MergePartner == parent2 {
					continue
				}

				parent1Idx, ok1 := nodeToIndex[parent1]
				parent2Idx, ok2 := nodeToIndex[parent2]
				if !ok1 || !ok2 {
					continue
				}

				// Р”РѕР±Р°РІР»СЏРµРј СЃРІСЏР·СЊ РїР°СЂС‚РЅС‘СЂРѕРІ (РѕС‚ Р»РµРІРѕР№ Рє РїСЂР°РІРѕР№)
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

	// РС‰РµРј РїР°СЂС‚РЅС‘СЂРѕРІ С‡РµСЂРµР· FamilyTree (Partners РїРѕР»Рµ РІ Person)
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

			// РџСЂРѕРїСѓСЃРєР°РµРј РµСЃР»Рё СЌС‚Рѕ С‚Р° Р¶Рµ РІРµСЂС€РёРЅР°
			if partnerNode == personNode {
				continue
			}

			// РџСЂРѕРїСѓСЃРєР°РµРј РµСЃР»Рё СЌС‚Рѕ MergePartner (СѓР¶Рµ РґРѕР±Р°РІРёР»Рё)
			if personNode.MergePartner == partnerNode {
				continue
			}

			partnerIdx, ok := nodeToIndex[partnerNode]
			if !ok {
				continue
			}

			// Р”РѕР±Р°РІР»СЏРµРј СЃРІСЏР·СЊ С‚РѕР»СЊРєРѕ РѕРґРёРЅ СЂР°Р· (РѕС‚ РјРµРЅСЊС€РµРіРѕ ID Рє Р±РѕР»СЊС€РµРјСѓ)
			if person.ID > partner.ID {
				continue
			}

			// РћРїСЂРµРґРµР»СЏРµРј Р»РµРІСѓСЋ Рё РїСЂР°РІСѓСЋ РІРµСЂС€РёРЅСѓ
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

// BuildRenderResult СЃС‚СЂРѕРёС‚ СЂРµР·СѓР»СЊС‚Р°С‚ РІРёР·СѓР°Р»РёР·Р°С†РёРё РёР· PeopleGrid
func BuildRenderResult(grid *PeopleGrid, tree *stage1_input.FamilyTree) *RenderResult {
	result := &RenderResult{
		Coords: make(map[int]Coord),
		Edges:  []Edge{},
	}

	// РЎРѕР±РёСЂР°РµРј РєРѕРѕСЂРґРёРЅР°С‚С‹ РІСЃРµС… Р»СЋРґРµР№ РёР· СЃРµС‚РєРё
	for _, pos := range grid.Positions {
		result.Coords[pos.Person.ID] = Coord{
			X: pos.X,
			Y: pos.Y,
		}
	}

	// Р”РѕР±Р°РІР»СЏРµРј СЃРІСЏР·Рё СЂРѕРґРёС‚РµР»СЊ-СЂРµР±С‘РЅРѕРє
	for _, person := range tree.People {
		if person.Layout == nil {
			continue
		}
		personCoord, ok := result.Coords[person.ID]
		if !ok {
			continue
		}

		// РЎРІСЏР·СЊ СЃ РјР°С‚РµСЂСЊСЋ
		if person.Mother != nil {
			if motherCoord, ok := result.Coords[person.Mother.ID]; ok {
				result.Edges = append(result.Edges, Edge{
					From:     motherCoord,
					To:       personCoord,
					EdgeType: "parent-child",
				})
			}
		}

		// РЎРІСЏР·СЊ СЃ РѕС‚С†РѕРј
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

	// Р”РѕР±Р°РІР»СЏРµРј СЃРІСЏР·Рё РјРµР¶РґСѓ РїР°СЂС‚РЅС‘СЂР°РјРё (РёР·Р±РµРіР°РµРј РґСѓР±Р»РёСЂРѕРІР°РЅРёСЏ)
	// РџР°СЂС‚РЅС‘СЂС‹ РґРѕР»Р¶РЅС‹ Р±С‹С‚СЊ РЅР° РѕРґРЅРѕРј СЃР»РѕРµ
	// РџСЂРѕРїСѓСЃРєР°РµРј СЃРІСЏР·Рё, РµСЃР»Рё РїР°СЂС‚РЅС‘СЂС‹ РІ РѕРґРЅРѕР№ РІРµСЂС€РёРЅРµ (РѕРґРёРЅР°РєРѕРІС‹Рµ РєРѕРѕСЂРґРёРЅР°С‚С‹)
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

			// РџР°СЂС‚РЅС‘СЂС‹ РґРѕР»Р¶РЅС‹ Р±С‹С‚СЊ РЅР° РѕРґРЅРѕРј СЃР»РѕРµ
			if personCoord.Y != partnerCoord.Y {
				continue
			}

			// РџСЂРѕРїСѓСЃРєР°РµРј, РµСЃР»Рё РїР°СЂС‚РЅС‘СЂС‹ РІ РѕРґРЅРѕР№ РІРµСЂС€РёРЅРµ (РѕРґРёРЅР°РєРѕРІС‹Рµ X)
			if personCoord.X == partnerCoord.X {
				continue
			}

			// РЎРѕР·РґР°С‘Рј СѓРЅРёРєР°Р»СЊРЅС‹Р№ РєР»СЋС‡ РґР»СЏ РїР°СЂС‹ (РјРµРЅСЊС€РёР№ ID РїРµСЂРІС‹Р№)
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

// Print РІС‹РІРѕРґРёС‚ СЂРµР·СѓР»СЊС‚Р°С‚ РІРёР·СѓР°Р»РёР·Р°С†РёРё
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
