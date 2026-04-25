package stage3_ordering

import (
	"fmt"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"
)

// CoordMatrix вЂ” РјР°С‚СЂРёС†Р° РІРµСЂС€РёРЅ СЃ РєРѕРѕСЂРґРёРЅР°С‚Р°РјРё
type CoordMatrix struct {
	// РЎР»РѕРё (РЅРѕРјРµСЂ СЃР»РѕСЏ -> СЃРїРёСЃРѕРє РІРµСЂС€РёРЅ)
	Layers map[int][]*CoordNode

	// РњРёРЅРёРјР°Р»СЊРЅС‹Р№ Рё РјР°РєСЃРёРјР°Р»СЊРЅС‹Р№ СЃР»РѕРё
	MinLayer int
	MaxLayer int

	// PersonToNode вЂ” РєР°СЂС‚Р° РґР»СЏ Р±С‹СЃС‚СЂРѕРіРѕ РїРѕРёСЃРєР° РІРµСЂС€РёРЅС‹ РїРѕ ID С‡РµР»РѕРІРµРєР°
	PersonToNode map[int]*CoordNode
}

// NewCoordMatrix СЃРѕР·РґР°С‘С‚ РЅРѕРІСѓСЋ РјР°С‚СЂРёС†Сѓ РєРѕРѕСЂРґРёРЅР°С‚
func NewCoordMatrix(minLayer, maxLayer int) *CoordMatrix {
	cm := &CoordMatrix{
		Layers:       make(map[int][]*CoordNode),
		MinLayer:     minLayer,
		MaxLayer:     maxLayer,
		PersonToNode: make(map[int]*CoordNode),
	}
	for layer := minLayer; layer <= maxLayer; layer++ {
		cm.Layers[layer] = []*CoordNode{}
	}
	return cm
}

// AddNode РґРѕР±Р°РІР»СЏРµС‚ РІРµСЂС€РёРЅСѓ РІ СЃР»РѕР№
func (cm *CoordMatrix) AddNode(node *CoordNode) {
	cm.Layers[node.Layer] = append(cm.Layers[node.Layer], node)
}

// BuildCoordMatrix СЃРѕР·РґР°С‘С‚ РјР°С‚СЂРёС†Сѓ РєРѕРѕСЂРґРёРЅР°С‚ РёР· OrderManager
func (om *OrderManager) BuildCoordMatrix() *CoordMatrix {
	// РЁР°Рі 1: РЎРѕР·РґР°С‘Рј CoordNode РґР»СЏ РєР°Р¶РґРѕР№ РІРµСЂС€РёРЅС‹, СЂР°Р·СЉРµРґРёРЅСЏСЏ СЃРєР»РµРµРЅРЅС‹Рµ
	nodeMap := make(map[*LayerNode][]*CoordNode) // РѕСЂРёРіРёРЅР°Р» -> СЃРѕР·РґР°РЅРЅС‹Рµ CoordNode
	// РЎРѕС…СЂР°РЅСЏРµРј РїРѕСЂСЏРґРѕРє РѕР±С…РѕРґР° РґР»СЏ РґРµС‚РµСЂРјРёРЅРёСЂРѕРІР°РЅРЅРѕСЃС‚Рё
	var orderedOrigNodes []*LayerNode

	// РЎРѕР·РґР°С‘Рј РІРµСЂС€РёРЅС‹
	for _, layer := range om.GetAllLayers() {
		for _, node := range layer.GetNodes() {
			orderedOrigNodes = append(orderedOrigNodes, node) // СЃРѕС…СЂР°РЅСЏРµРј РїРѕСЂСЏРґРѕРє
			if node.IsPseudo {
				// РџСЃРµРІРґРѕРІРµСЂС€РёРЅР° вЂ” РѕРґРёРЅ CoordNode
				cn := &CoordNode{
					IsPseudo:     true,
					Layer:        node.Layer,
					OriginalNode: node,
					Up:           make([]*CoordNode, 0),
					Down:         make([]*CoordNode, 0),
				}
				nodeMap[node] = []*CoordNode{cn}
			} else if len(node.People) == 2 {
				// РЎРєР»РµРµРЅРЅР°СЏ РІРµСЂС€РёРЅР° вЂ” СЂР°Р·СЉРµРґРёРЅСЏРµРј РЅР° РґРІРµ
				cn1 := &CoordNode{
					People:       []*stage1_input.Person{node.People[0]},
					Layer:        node.Layer,
					OriginalNode: node,
					WasMerged:    true,
					Up:           make([]*CoordNode, 0),
					Down:         make([]*CoordNode, 0),
				}
				cn2 := &CoordNode{
					People:       []*stage1_input.Person{node.People[1]},
					Layer:        node.Layer,
					OriginalNode: node,
					WasMerged:    true,
					Up:           make([]*CoordNode, 0),
					Down:         make([]*CoordNode, 0),
				}
				cn1.MergePartner = cn2
				cn2.MergePartner = cn1
				nodeMap[node] = []*CoordNode{cn1, cn2}
			} else if len(node.People) == 1 {
				// РћР±С‹С‡РЅР°СЏ РІРµСЂС€РёРЅР° СЃ РѕРґРЅРёРј С‡РµР»РѕРІРµРєРѕРј
				cn := &CoordNode{
					People:       node.People,
					Layer:        node.Layer,
					OriginalNode: node,
					Up:           make([]*CoordNode, 0),
					Down:         make([]*CoordNode, 0),
				}
				nodeMap[node] = []*CoordNode{cn}
			}
		}
	}

	// РЁР°Рі 2: РЈСЃС‚Р°РЅР°РІР»РёРІР°РµРј СЃРІСЏР·Рё Up Рё Down
	for _, origNode := range orderedOrigNodes {
		coordNodes := nodeMap[origNode]
		if origNode.IsPseudo {
			// РџСЃРµРІРґРѕРІРµСЂС€РёРЅР°: РѕРґРёРЅ Up, РѕРґРёРЅ Down
			cn := coordNodes[0]

			// Up
			if len(origNode.Up) > 0 && origNode.Up[0] != nil {
				upNodes := nodeMap[origNode.Up[0]]
				if len(upNodes) > 0 {
					cn.Up = append(cn.Up, upNodes[0])
				}
			}

			// Down (LeftDown = RightDown РґР»СЏ РїСЃРµРІРґРѕ)
			if origNode.LeftDown != nil {
				downNodes := nodeMap[origNode.LeftDown]
				if len(downNodes) > 0 {
					// Р‘РµСЂС‘Рј РїРµСЂРІСѓСЋ РІРµСЂС€РёРЅСѓ (РµСЃР»Рё Р±С‹Р»Р° СЃРєР»РµРµРЅР° вЂ” Р»РµРІСѓСЋ)
					cn.Down = append(cn.Down, downNodes[0])
				}
			}
		} else if len(origNode.People) == 2 {
			// РЎРєР»РµРµРЅРЅР°СЏ РІРµСЂС€РёРЅР° вЂ” СЂР°Р·СЉРµРґРёРЅС‘РЅРЅР°СЏ РЅР° РґРІРµ
			cn1, cn2 := coordNodes[0], coordNodes[1]

			// Up[0] -> cn1, Up[1] -> cn2
			if len(origNode.Up) > 0 && origNode.Up[0] != nil {
				upNodes := nodeMap[origNode.Up[0]]
				if len(upNodes) > 0 {
					cn1.Up = append(cn1.Up, upNodes[0])
				}
			}
			if len(origNode.Up) > 1 && origNode.Up[1] != nil {
				upNodes := nodeMap[origNode.Up[1]]
				if len(upNodes) > 0 {
					cn2.Up = append(cn2.Up, upNodes[0])
				}
			}

			// Down вЂ” РѕР±С‰РёРµ РґРµС‚Рё, РґРѕР±Р°РІР»СЏРµРј Рє РѕР±РѕРёРј
			if origNode.LeftDown != nil {
				// РЎРѕР±РёСЂР°РµРј РІСЃРµС… РґРµС‚РµР№
				collectChildren(origNode, nodeMap, cn1, cn2)
			}
		} else if len(origNode.People) == 1 {
			// РћР±С‹С‡РЅР°СЏ РІРµСЂС€РёРЅР°
			cn := coordNodes[0]

			// Up
			if len(origNode.Up) > 0 && origNode.Up[0] != nil {
				upNodes := nodeMap[origNode.Up[0]]
				if len(upNodes) > 0 {
					cn.Up = append(cn.Up, upNodes[0])
				}
			}

			// Down вЂ” СЃРѕР±РёСЂР°РµРј РІСЃРµС… РґРµС‚РµР№
			if origNode.LeftDown != nil {
				collectChildrenSingle(origNode, nodeMap, cn)
			}
		}
	}

	// РЁР°Рі 3: РЈСЃС‚Р°РЅР°РІР»РёРІР°РµРј СЃРІСЏР·Рё СЂРѕРґРёС‚РµР»СЊ-СЂРµР±С‘РЅРѕРє Рё AddedLeft
	setupParentChildLinks(om, nodeMap, orderedOrigNodes)

	// РЁР°Рі 4: Р Р°СЃСЃС‚Р°РІР»СЏРµРј РєРѕРѕСЂРґРёРЅР°С‚С‹
	cm := assignCoordinates(om, nodeMap)

	// РЁР°Рі 5: РЎРєР»РµРёРІР°РµРј РѕР±СЂР°С‚РЅРѕ РІРµСЂС€РёРЅС‹, РєРѕС‚РѕСЂС‹Рµ Р±С‹Р»Рё СЃРєР»РµРµРЅС‹
	mergeBackNodes(cm, nodeMap, orderedOrigNodes)

	// РЁР°Рі 5.5: Р—Р°РїРѕР»РЅСЏРµРј PersonToNode РґР»СЏ Р±С‹СЃС‚СЂРѕРіРѕ РїРѕРёСЃРєР° РїР°СЂС‚РЅС‘СЂРѕРІ
	buildPersonToNode(cm)

	// РЁР°Рі 6: РћРїС‚РёРјРёР·РёСЂСѓРµРј РїРѕР·РёС†РёРё РІРµСЂС€РёРЅ
	optimizePositions(cm)

	// РЁР°Рі 7: Р Р°Р·РґРµР»СЏРµРј СЃРєР»РµРµРЅРЅС‹Рµ РІРµСЂС€РёРЅС‹ С€РёСЂРёРЅС‹ 4 РЅР° РґРІРµ РІРµСЂС€РёРЅС‹ С€РёСЂРёРЅС‹ 2
	splitMergedNodes(cm)

	// РЁР°Рі 7.5: РћР±РЅРѕРІР»СЏРµРј PersonToNode РїРѕСЃР»Рµ split
	buildPersonToNode(cm)

	// РЁР°Рі 8: РќРѕСЂРјР°Р»РёР·СѓРµРј РєРѕРѕСЂРґРёРЅР°С‚С‹ (СЃРґРІРёРіР°РµРј С‡С‚РѕР±С‹ РјРёРЅРёРјСѓРј Р±С‹Р» 0)
	normalizeCoordinates(cm)

	return cm
}

// setupParentChildLinks СѓСЃС‚Р°РЅР°РІР»РёРІР°РµС‚ СЃРІСЏР·Рё СЂРѕРґРёС‚РµР»СЊ-СЂРµР±С‘РЅРѕРє Рё С„Р»Р°Рі AddedLeft
func setupParentChildLinks(om *OrderManager, nodeMap map[*LayerNode][]*CoordNode, orderedOrigNodes []*LayerNode) {
	// РЈСЃС‚Р°РЅР°РІР»РёРІР°РµРј AddedLeft РёР· OriginalNode
	for _, origNode := range orderedOrigNodes {
		coordNodes := nodeMap[origNode]
		for _, cn := range coordNodes {
			cn.AddedLeft = origNode.AddedLeft
		}
	}

	// РЈСЃС‚Р°РЅР°РІР»РёРІР°РµРј СЃРІСЏР·Рё ParentNodes Рё Children
	for _, origNode := range orderedOrigNodes {
		coordNodes := nodeMap[origNode]
		if origNode.IsPseudo {
			continue
		}

		for _, cn := range coordNodes {
			// РРЅРёС†РёР°Р»РёР·РёСЂСѓРµРј РјР°СЃСЃРёРІС‹
			cn.ParentNodes = make([]*CoordNode, len(cn.People))
			cn.ParentPersonIndex = make([]int, len(cn.People))
			cn.Children = []*CoordNode{}

			// РќР°С…РѕРґРёРј СЂРѕРґРёС‚РµР»РµР№ С‡РµСЂРµР· Up
			for i := range cn.People {
				if i < len(cn.Up) && cn.Up[i] != nil {
					parentCN := cn.Up[i]
					// Р•СЃР»Рё СЂРѕРґРёС‚РµР»СЊ вЂ” РїСЃРµРІРґРѕРІРµСЂС€РёРЅР°, РёРґС‘Рј РІРІРµСЂС… РґРѕ СЂРµР°Р»СЊРЅРѕР№ РІРµСЂС€РёРЅС‹
					for parentCN != nil && parentCN.IsPseudo {
						if len(parentCN.Up) > 0 {
							parentCN = parentCN.Up[0]
						} else {
							parentCN = nil
						}
					}
					cn.ParentNodes[i] = parentCN
					// РћРїСЂРµРґРµР»СЏРµРј РёРЅРґРµРєСЃ С‡РµР»РѕРІРµРєР° РІ СЂРѕРґРёС‚РµР»СЊСЃРєРѕР№ РІРµСЂС€РёРЅРµ
					if parentCN != nil {
						cn.ParentPersonIndex[i] = findPersonIndexInParent(cn.People[i], parentCN, origNode, nodeMap)
					}
				}
			}
		}
	}

	// РЈСЃС‚Р°РЅР°РІР»РёРІР°РµРј Children РґР»СЏ РєР°Р¶РґРѕР№ РІРµСЂС€РёРЅС‹
	for _, origNode := range orderedOrigNodes {
		coordNodes := nodeMap[origNode]
		for _, cn := range coordNodes {
			for i, parentCN := range cn.ParentNodes {
				if parentCN != nil {
					// РџСЂРѕРІРµСЂСЏРµРј, РЅРµ РґРѕР±Р°РІР»РµРЅ Р»Рё СѓР¶Рµ СЌС‚РѕС‚ СЂРµР±С‘РЅРѕРє
					found := false
					for _, child := range parentCN.Children {
						if child == cn {
							found = true
							break
						}
					}
					if !found {
						parentCN.Children = append(parentCN.Children, cn)
					}
					_ = i // РёСЃРїРѕР»СЊР·СѓРµРј РґР»СЏ РѕРїСЂРµРґРµР»РµРЅРёСЏ РєР°РєРѕР№ РёРјРµРЅРЅРѕ С‡РµР»РѕРІРµРє вЂ” РїРѕРєР° РЅРµ РЅСѓР¶РЅРѕ
				}
			}
		}
	}
}

// buildPersonToNode Р·Р°РїРѕР»РЅСЏРµС‚ РєР°СЂС‚Сѓ PersonToNode РІ CoordMatrix
func buildPersonToNode(cm *CoordMatrix) {
	cm.PersonToNode = make(map[int]*CoordNode)
	for layerNum := cm.MinLayer; layerNum <= cm.MaxLayer; layerNum++ {
		for _, node := range cm.Layers[layerNum] {
			if node.IsPseudo {
				continue
			}
			for _, person := range node.People {
				cm.PersonToNode[person.ID] = node
			}
		}
	}
}

// findPersonIndexInParent РѕРїСЂРµРґРµР»СЏРµС‚ РёРЅРґРµРєСЃ С‡РµР»РѕРІРµРєР° РІ СЂРѕРґРёС‚РµР»СЊСЃРєРѕР№ РІРµСЂС€РёРЅРµ
func findPersonIndexInParent(child *stage1_input.Person, parentCN *CoordNode, childOrigNode *LayerNode, nodeMap map[*LayerNode][]*CoordNode) int {
	// Р•СЃР»Рё СЂРѕРґРёС‚РµР»СЊ вЂ” РѕРґРёРЅРѕС‡РЅР°СЏ РІРµСЂС€РёРЅР°, РёРЅРґРµРєСЃ 0
	if len(parentCN.People) == 1 {
		return 0
	}
	// Р•СЃР»Рё СЂРѕРґРёС‚РµР»СЊ вЂ” РїР°СЂР°, РЅСѓР¶РЅРѕ РѕРїСЂРµРґРµР»РёС‚СЊ С‡РµСЂРµР· РєРѕРіРѕ СЃРІСЏР·СЊ
	// РЎРјРѕС‚СЂРёРј Up РІ РѕСЂРёРіРёРЅР°Р»СЊРЅРѕР№ РІРµСЂС€РёРЅРµ
	for i, upNode := range childOrigNode.Up {
		if upNode == nil {
			continue
		}
		upCoordNodes := nodeMap[upNode]
		for _, upCN := range upCoordNodes {
			if upCN == parentCN || (upCN.MergePartner != nil && upCN.MergePartner == parentCN) {
				return i
			}
		}
	}
	return 0
}

// collectChildren СЃРѕР±РёСЂР°РµС‚ РґРµС‚РµР№ РґР»СЏ РїР°СЂС‹ СЂРѕРґРёС‚РµР»РµР№
func collectChildren(parentNode *LayerNode, nodeMap map[*LayerNode][]*CoordNode, cn1, cn2 *CoordNode) {
	// РџСЂРѕС…РѕРґРёРј РїРѕ РІСЃРµРј РґРµС‚СЏРј С‡РµСЂРµР· LeftDown -> Next
	current := parentNode.LeftDown
	for current != nil {
		if childNodes, ok := nodeMap[current]; ok {
			for _, childCN := range childNodes {
				// РџСЂРѕРІРµСЂСЏРµРј, РЅРµ РґРѕР±Р°РІР»РµРЅ Р»Рё СѓР¶Рµ
				found := false
				for _, existing := range cn1.Down {
					if existing == childCN {
						found = true
						break
					}
				}
				if !found {
					cn1.Down = append(cn1.Down, childCN)
					cn2.Down = append(cn2.Down, childCN)
				}
			}
		}
		if current == parentNode.RightDown {
			break
		}
		current = current.Next
	}
}

// collectChildrenSingle СЃРѕР±РёСЂР°РµС‚ РґРµС‚РµР№ РґР»СЏ РѕРґРёРЅРѕС‡РЅРѕРіРѕ СЂРѕРґРёС‚РµР»СЏ
func collectChildrenSingle(parentNode *LayerNode, nodeMap map[*LayerNode][]*CoordNode, cn *CoordNode) {
	// РџСЂРѕС…РѕРґРёРј РїРѕ РІСЃРµРј РґРµС‚СЏРј С‡РµСЂРµР· LeftDown -> Next РґРѕ RightDown
	current := parentNode.LeftDown
	for current != nil {
		if childNodes, ok := nodeMap[current]; ok {
			for _, childCN := range childNodes {
				// РџСЂРѕРІРµСЂСЏРµРј, РЅРµ РґРѕР±Р°РІР»РµРЅ Р»Рё СѓР¶Рµ
				found := false
				for _, existing := range cn.Down {
					if existing == childCN {
						found = true
						break
					}
				}
				if !found {
					cn.Down = append(cn.Down, childCN)
				}
			}
		}
		if current == parentNode.RightDown {
			break
		}
		current = current.Next
	}
}

// assignCoordinates СЂР°СЃСЃС‚Р°РІР»СЏРµС‚ РєРѕРѕСЂРґРёРЅР°С‚С‹ РІРµСЂС€РёРЅР°Рј
// РћР±С…РѕРґРёРј РєРѕР»РѕРЅРєР° Р·Р° РєРѕР»РѕРЅРєРѕР№ (СЃРЅРёР·Сѓ РІРІРµСЂС… РІ РєР°Р¶РґРѕР№ РєРѕР»РѕРЅРєРµ)
func assignCoordinates(om *OrderManager, nodeMap map[*LayerNode][]*CoordNode) *CoordMatrix {
	cm := NewCoordMatrix(om.minLayer, om.maxLayer)

	// Р“Р»РѕР±Р°Р»СЊРЅС‹Р№ СЃС‡С‘С‚С‡РёРє РєРѕРѕСЂРґРёРЅР°С‚
	globalCoord := 0
	const coordStep = 20

	// РџРѕР»СѓС‡Р°РµРј РѕС‚СЃРѕСЂС‚РёСЂРѕРІР°РЅРЅС‹Рµ РЅРѕРјРµСЂР° СЃР»РѕС‘РІ (СЃРЅРёР·Сѓ РІРІРµСЂС…)
	sortedLayers := om.getSortedLayerNumbers()

	// РЎРѕР±РёСЂР°РµРј СЃРїРёСЃРєРё CoordNode РґР»СЏ РєР°Р¶РґРѕРіРѕ СЃР»РѕСЏ
	layerNodes := make(map[int][]*CoordNode)
	for _, layerNum := range sortedLayers {
		layer := om.Layers[layerNum]
		if layer == nil {
			continue
		}
		coordNodes := []*CoordNode{}
		for _, node := range layer.GetNodes() {
			if cns, ok := nodeMap[node]; ok {
				coordNodes = append(coordNodes, cns...)
			}
		}
		layerNodes[layerNum] = coordNodes
	}

	// РРЅРґРµРєСЃ С‚РµРєСѓС‰РµР№ РїРѕР·РёС†РёРё РІ РєР°Р¶РґРѕРј СЃР»РѕРµ
	layerIndices := make(map[int]int)
	for _, layerNum := range sortedLayers {
		layerIndices[layerNum] = 0
	}

	// РћРїСЂРµРґРµР»СЏРµРј РјР°РєСЃРёРјР°Р»СЊРЅРѕРµ РєРѕР»РёС‡РµСЃС‚РІРѕ РІРµСЂС€РёРЅ РІ СЃР»РѕРµ
	maxNodesInLayer := 0
	for _, nodes := range layerNodes {
		if len(nodes) > maxNodesInLayer {
			maxNodesInLayer = len(nodes)
		}
	}

	// РћР±С…РѕРґРёРј РєРѕР»РѕРЅРєР° Р·Р° РєРѕР»РѕРЅРєРѕР№
	for columnIdx := 0; columnIdx < maxNodesInLayer; columnIdx++ {
		// РћР±СЂР°Р±Р°С‚С‹РІР°РµРј РєР°Р¶РґС‹Р№ СЃР»РѕР№ СЃРЅРёР·Сѓ РІРІРµСЂС…
		for _, layerNum := range sortedLayers {
			idx := layerIndices[layerNum]
			nodes := layerNodes[layerNum]

			if idx >= len(nodes) {
				continue
			}

			cn := nodes[idx]

			// РџСЂРѕРІРµСЂСЏРµРј, СѓР¶Рµ СЂР°Р·РјРµС‰РµРЅР° Р»Рё РІРµСЂС€РёРЅР°
			if cn.Left != 0 || cn.Right != 0 {
				layerIndices[layerNum]++
				continue
			}

			// Р Р°Р·РјРµС‰Р°РµРј РІРµСЂС€РёРЅСѓ
			if cn.IsPseudo {
				// РџСЃРµРІРґРѕРІРµСЂС€РёРЅР° вЂ” С€РёСЂРёРЅР° 0
				cn.Left = globalCoord
				cn.Right = globalCoord
			} else {
				// РћР±С‹С‡РЅР°СЏ РІРµСЂС€РёРЅР° вЂ” С€РёСЂРёРЅР° 2
				cn.Left = globalCoord
				cn.Right = globalCoord + 2
			}

			cm.AddNode(cn)
			layerIndices[layerNum]++
		}

		// РЈРІРµР»РёС‡РёРІР°РµРј РєРѕРѕСЂРґРёРЅР°С‚Сѓ РѕРґРёРЅ СЂР°Р· РЅР° РІСЃСЋ РєРѕР»РѕРЅРєСѓ
		globalCoord += coordStep
	}

	return cm
}

// getSortedLayerNumbers РІРѕР·РІСЂР°С‰Р°РµС‚ РЅРѕРјРµСЂР° СЃР»РѕС‘РІ РѕС‚СЃРѕСЂС‚РёСЂРѕРІР°РЅРЅС‹Рµ СЃРЅРёР·Сѓ РІРІРµСЂС…
func (om *OrderManager) getSortedLayerNumbers() []int {
	layers := []int{}
	for layerNum := om.minLayer; layerNum <= om.maxLayer; layerNum++ {
		layers = append(layers, layerNum)
	}
	return layers
}

// mergeBackNodes СЃРєР»РµРёРІР°РµС‚ РѕР±СЂР°С‚РЅРѕ СЂР°Р·СЉРµРґРёРЅС‘РЅРЅС‹Рµ РІРµСЂС€РёРЅС‹
func mergeBackNodes(cm *CoordMatrix, nodeMap map[*LayerNode][]*CoordNode, orderedOrigNodes []*LayerNode) {
	// РќР°С…РѕРґРёРј РїР°СЂС‹ РґР»СЏ СЃРєР»РµР№РєРё
	merged := make(map[*CoordNode]bool)

	for _, origNode := range orderedOrigNodes {
		coordNodes := nodeMap[origNode]
		if len(coordNodes) == 2 && coordNodes[0].WasMerged && coordNodes[1].WasMerged {
			cn1, cn2 := coordNodes[0], coordNodes[1]
			if merged[cn1] || merged[cn2] {
				continue
			}

			// РџСЂРѕРІРµСЂСЏРµРј, С‡С‚Рѕ РѕР±Рµ РІРµСЂС€РёРЅС‹ СЂР°Р·РјРµС‰РµРЅС‹
			if cn1.Left == 0 && cn1.Right == 0 {
				fmt.Printf("WARN: cn1 РЅРµ СЂР°Р·РјРµС‰РµРЅР°: %v\n", cn1.People)
				continue
			}
			if cn2.Left == 0 && cn2.Right == 0 {
				fmt.Printf("WARN: cn2 РЅРµ СЂР°Р·РјРµС‰РµРЅР°: %v\n", cn2.People)
				continue
			}

			// РЎРѕР·РґР°С‘Рј СЃРєР»РµРµРЅРЅСѓСЋ РІРµСЂС€РёРЅСѓ
			mergedNode := &CoordNode{
				People:            append(cn1.People, cn2.People...),
				Layer:             cn1.Layer,
				IsPseudo:          false,
				Up:                make([]*CoordNode, 0),
				Down:              make([]*CoordNode, 0),
				Children:          make([]*CoordNode, 0),
				ParentNodes:       make([]*CoordNode, 0),
				ParentPersonIndex: make([]int, 0),
			}

			// РљРѕРѕСЂРґРёРЅР°С‚С‹: С†РµРЅС‚СЂ РјРµР¶РґСѓ РґРІСѓРјСЏ РІРµСЂС€РёРЅР°РјРё
			minLeft := cn1.Left
			if cn2.Left < minLeft {
				minLeft = cn2.Left
			}
			maxRight := cn1.Right
			if cn2.Right > maxRight {
				maxRight = cn2.Right
			}

			// РЎРєР»РµРµРЅРЅР°СЏ РІРµСЂС€РёРЅР° С€РёСЂРёРЅС‹ 4
			center := (minLeft + maxRight) / 2
			mergedNode.Left = center - 2
			mergedNode.Right = center + 2

			// РљРѕРїРёСЂСѓРµРј СЃРІСЏР·Рё
			mergedNode.Up = append(mergedNode.Up, cn1.Up...)
			mergedNode.Up = append(mergedNode.Up, cn2.Up...)
			mergedNode.Down = append(mergedNode.Down, cn1.Down...)
			mergedNode.Down = append(mergedNode.Down, cn2.Down...)

			// РљРѕРїРёСЂСѓРµРј ParentNodes Рё ParentPersonIndex (РґР»СЏ СЂР°СЃС‡С‘С‚Р° СЂР°СЃСЃС‚РѕСЏРЅРёР№)
			// cn1 СЃРѕРѕС‚РІРµС‚СЃС‚РІСѓРµС‚ People[0], cn2 СЃРѕРѕС‚РІРµС‚СЃС‚РІСѓРµС‚ People[1]
			mergedNode.ParentNodes = append(mergedNode.ParentNodes, cn1.ParentNodes...)
			mergedNode.ParentNodes = append(mergedNode.ParentNodes, cn2.ParentNodes...)
			mergedNode.ParentPersonIndex = append(mergedNode.ParentPersonIndex, cn1.ParentPersonIndex...)
			mergedNode.ParentPersonIndex = append(mergedNode.ParentPersonIndex, cn2.ParentPersonIndex...)

			// РљРѕРїРёСЂСѓРµРј Children
			mergedNode.Children = append(mergedNode.Children, cn1.Children...)
			mergedNode.Children = append(mergedNode.Children, cn2.Children...)

			// РЈРґР°Р»СЏРµРј РґСѓР±Р»РёРєР°С‚С‹ РІ Down
			uniqueDown := []*CoordNode{}
			downSet := make(map[*CoordNode]bool)
			for _, d := range mergedNode.Down {
				if !downSet[d] {
					downSet[d] = true
					uniqueDown = append(uniqueDown, d)
				}
			}
			mergedNode.Down = uniqueDown

			// РЈРґР°Р»СЏРµРј РґСѓР±Р»РёРєР°С‚С‹ РІ Children
			uniqueChildren := []*CoordNode{}
			childSet := make(map[*CoordNode]bool)
			for _, c := range mergedNode.Children {
				if !childSet[c] {
					childSet[c] = true
					uniqueChildren = append(uniqueChildren, c)
				}
			}
			mergedNode.Children = uniqueChildren

			// РћР±РЅРѕРІР»СЏРµРј СЃСЃС‹Р»РєРё ParentNodes РІ РґРµС‚СЏС… РЅР° mergedNode
			for _, child := range mergedNode.Children {
				if child == nil {
					continue
				}
				// Р—Р°РјРµРЅСЏРµРј cn1 Рё cn2 РЅР° mergedNode РІ ParentNodes СЂРµР±С‘РЅРєР°
				for i, parentCN := range child.ParentNodes {
					if parentCN == cn1 || parentCN == cn2 {
						child.ParentNodes[i] = mergedNode
					}
				}
			}

			// РћР±РЅРѕРІР»СЏРµРј СЃСЃС‹Р»РєРё Children РІ СЂРѕРґРёС‚РµР»СЏС… РЅР° mergedNode
			for _, parentCN := range mergedNode.ParentNodes {
				if parentCN == nil {
					continue
				}
				// Р—Р°РјРµРЅСЏРµРј cn1 Рё cn2 РЅР° mergedNode РІ Children СЂРѕРґРёС‚РµР»СЏ
				for i, childCN := range parentCN.Children {
					if childCN == cn1 || childCN == cn2 {
						parentCN.Children[i] = mergedNode
					}
				}
				// РЈРґР°Р»СЏРµРј РґСѓР±Р»РёРєР°С‚С‹ РІ Children СЂРѕРґРёС‚РµР»СЏ
				uniqueParentChildren := []*CoordNode{}
				parentChildSet := make(map[*CoordNode]bool)
				for _, c := range parentCN.Children {
					if !parentChildSet[c] {
						parentChildSet[c] = true
						uniqueParentChildren = append(uniqueParentChildren, c)
					}
				}
				parentCN.Children = uniqueParentChildren
			}

			// РџРµСЂРµРјРµС‰Р°РµРј РїСЃРµРІРґРѕРІРµСЂС€РёРЅС‹ РЅР°Рґ cn1 Рё cn2
			if len(cn1.Up) > 0 && cn1.Up[0] != nil && cn1.Up[0].IsPseudo {
				movePseudoChainToCoord(cn1.Up[0], mergedNode.Left+1)
			}
			if len(cn2.Up) > 0 && cn2.Up[0] != nil && cn2.Up[0].IsPseudo {
				movePseudoChainToCoord(cn2.Up[0], mergedNode.Right-1)
			}

			// РЈРґР°Р»СЏРµРј СЃС‚Р°СЂС‹Рµ РІРµСЂС€РёРЅС‹ РёР· СЃР»РѕСЏ Рё РґРѕР±Р°РІР»СЏРµРј РЅРѕРІСѓСЋ
			layer := cm.Layers[cn1.Layer]
			newLayer := []*CoordNode{}
			for _, node := range layer {
				if node != cn1 && node != cn2 {
					newLayer = append(newLayer, node)
				}
			}
			newLayer = append(newLayer, mergedNode)
			cm.Layers[cn1.Layer] = newLayer

			merged[cn1] = true
			merged[cn2] = true
		}
	}
}

// movePseudoChainToCoord РїРµСЂРµРјРµС‰Р°РµС‚ С†РµРїРѕС‡РєСѓ РїСЃРµРІРґРѕРІРµСЂС€РёРЅ РЅР° Р·Р°РґР°РЅРЅСѓСЋ РєРѕРѕСЂРґРёРЅР°С‚Сѓ
func movePseudoChainToCoord(start *CoordNode, coord int) {
	current := start
	for current != nil && current.IsPseudo {
		current.Left = coord
		current.Right = coord
		if len(current.Up) > 0 {
			current = current.Up[0]
		} else {
			break
		}
	}
}

// splitMergedNodes СЂР°Р·РґРµР»СЏРµС‚ СЃРєР»РµРµРЅРЅС‹Рµ РІРµСЂС€РёРЅС‹ С€РёСЂРёРЅС‹ 4 РЅР° РґРІРµ РІРµСЂС€РёРЅС‹ С€РёСЂРёРЅС‹ 2
func splitMergedNodes(cm *CoordMatrix) {
	for layerNum := cm.MinLayer; layerNum <= cm.MaxLayer; layerNum++ {
		newLayer := []*CoordNode{}

		for _, node := range cm.Layers[layerNum] {
			// Р•СЃР»Рё СЌС‚Рѕ СЃРєР»РµРµРЅРЅР°СЏ РІРµСЂС€РёРЅР° СЃ 2 Р»СЋРґСЊРјРё Рё С€РёСЂРёРЅРѕР№ 4
			if len(node.People) == 2 && node.Width() == 4 {
				// РЎРѕР·РґР°С‘Рј РґРІРµ РѕС‚РґРµР»СЊРЅС‹Рµ РІРµСЂС€РёРЅС‹ С€РёСЂРёРЅС‹ 2
				cn1 := &CoordNode{
					People:            []*stage1_input.Person{node.People[0]},
					Layer:             node.Layer,
					IsPseudo:          false,
					Left:              node.Left,
					Right:             node.Left + 2,
					Up:                make([]*CoordNode, 0),
					Down:              make([]*CoordNode, 0),
					Children:          make([]*CoordNode, 0),
					ParentNodes:       make([]*CoordNode, 0),
					ParentPersonIndex: make([]int, 0),
					WasMerged:         true,
				}
				cn2 := &CoordNode{
					People:            []*stage1_input.Person{node.People[1]},
					Layer:             node.Layer,
					IsPseudo:          false,
					Left:              node.Left + 2,
					Right:             node.Right,
					Up:                make([]*CoordNode, 0),
					Down:              make([]*CoordNode, 0),
					Children:          make([]*CoordNode, 0),
					ParentNodes:       make([]*CoordNode, 0),
					ParentPersonIndex: make([]int, 0),
					WasMerged:         true,
				}

				// РЈСЃС‚Р°РЅР°РІР»РёРІР°РµРј СЃРІСЏР·СЊ РїР°СЂС‚РЅС‘СЂРѕРІ
				cn1.MergePartner = cn2
				cn2.MergePartner = cn1

				// РљРѕРїРёСЂСѓРµРј СЃРІСЏР·Рё Up
				if len(node.Up) > 0 {
					cn1.Up = append(cn1.Up, node.Up[0])
				}
				if len(node.Up) > 1 {
					cn2.Up = append(cn2.Up, node.Up[1])
				}

				// РљРѕРїРёСЂСѓРµРј Children Рє РѕР±РµРёРј РІРµСЂС€РёРЅР°Рј
				cn1.Children = append(cn1.Children, node.Children...)
				cn2.Children = append(cn2.Children, node.Children...)

				// РљРѕРїРёСЂСѓРµРј ParentNodes
				if len(node.ParentNodes) > 0 {
					cn1.ParentNodes = append(cn1.ParentNodes, node.ParentNodes[0])
				}
				if len(node.ParentNodes) > 1 {
					cn2.ParentNodes = append(cn2.ParentNodes, node.ParentNodes[1])
				}

				// РћР±РЅРѕРІР»СЏРµРј СЃСЃС‹Р»РєРё РІ РґРµС‚СЏС…
				for _, child := range node.Children {
					if child == nil {
						continue
					}
					for i, parentCN := range child.ParentNodes {
						if parentCN == node {
							// РћРїСЂРµРґРµР»СЏРµРј, РєР°РєР°СЏ РёР· РЅРѕРІС‹С… РІРµСЂС€РёРЅ РґРѕР»Р¶РЅР° Р±С‹С‚СЊ СЂРѕРґРёС‚РµР»РµРј
							if i < len(node.People) && i == 0 {
								child.ParentNodes[i] = cn1
							} else {
								child.ParentNodes[i] = cn2
							}
						}
					}
				}

				newLayer = append(newLayer, cn1, cn2)
			} else {
				newLayer = append(newLayer, node)
			}
		}

		cm.Layers[layerNum] = newLayer
	}
}

// normalizeCoordinates СЃРґРІРёРіР°РµС‚ РІСЃРµ РєРѕРѕСЂРґРёРЅР°С‚С‹ С‡С‚РѕР±С‹ РјРёРЅРёРјСѓРј Р±С‹Р» 0
func normalizeCoordinates(cm *CoordMatrix) {
	// РќР°С…РѕРґРёРј РјРёРЅРёРјР°Р»СЊРЅСѓСЋ РєРѕРѕСЂРґРёРЅР°С‚Сѓ
	minCoord := 0
	first := true
	for layerNum := cm.MinLayer; layerNum <= cm.MaxLayer; layerNum++ {
		for _, node := range cm.Layers[layerNum] {
			if first || node.Left < minCoord {
				minCoord = node.Left
				first = false
			}
		}
	}

	// РЎРґРІРёРіР°РµРј РІСЃРµ РєРѕРѕСЂРґРёРЅР°С‚С‹
	if minCoord != 0 {
		for layerNum := cm.MinLayer; layerNum <= cm.MaxLayer; layerNum++ {
			for _, node := range cm.Layers[layerNum] {
				node.Left -= minCoord
				node.Right -= minCoord
			}
		}
	}
}

// PrintCoordinates РІС‹РІРѕРґРёС‚ РєРѕРѕСЂРґРёРЅР°С‚С‹ РІСЃРµС… РІРµСЂС€РёРЅ
func (cm *CoordMatrix) PrintCoordinates() {
	fmt.Println("\n=== РљРѕРѕСЂРґРёРЅР°С‚С‹ РІРµСЂС€РёРЅ (РЅРѕРІР°СЏ СЃРёСЃС‚РµРјР°) ===")

	for layerNum := cm.MaxLayer; layerNum >= cm.MinLayer; layerNum-- {
		nodes := cm.Layers[layerNum]
		fmt.Printf("\nРЎР»РѕР№ %d:\n", layerNum)

		for _, node := range nodes {
			if node.IsPseudo {
				fmt.Printf("  [РїСЃРµРІРґРѕ] Left=%d, Right=%d\n", node.Left, node.Right)
			} else {
				names := ""
				for i, p := range node.People {
					if i > 0 {
						names += ", "
					}
					names += fmt.Sprintf("%s(%d)", p.Name, p.ID)
				}
				fmt.Printf("  [%s] Left=%d, Right=%d, Width=%d\n", names, node.Left, node.Right, node.Width())
			}
		}
	}
}
