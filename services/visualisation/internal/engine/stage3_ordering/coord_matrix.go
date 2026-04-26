package stage3_ordering

import (
	"fmt"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"
)

type CoordMatrix struct {
	Layers map[int][]*CoordNode

	MinLayer int
	MaxLayer int

	PersonToNode map[int]*CoordNode
}

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

func (cm *CoordMatrix) AddNode(node *CoordNode) {
	cm.Layers[node.Layer] = append(cm.Layers[node.Layer], node)
}

func (om *OrderManager) BuildCoordMatrix() *CoordMatrix {

	nodeMap := make(map[*LayerNode][]*CoordNode)

	var orderedOrigNodes []*LayerNode

	for _, layer := range om.GetAllLayers() {
		for _, node := range layer.GetNodes() {
			orderedOrigNodes = append(orderedOrigNodes, node)
			if node.IsPseudo {

				cn := &CoordNode{
					IsPseudo:     true,
					Layer:        node.Layer,
					OriginalNode: node,
					Up:           make([]*CoordNode, 0),
					Down:         make([]*CoordNode, 0),
				}
				nodeMap[node] = []*CoordNode{cn}
			} else if len(node.People) == 2 {

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

	for _, origNode := range orderedOrigNodes {
		coordNodes := nodeMap[origNode]
		if origNode.IsPseudo {

			cn := coordNodes[0]

			if len(origNode.Up) > 0 && origNode.Up[0] != nil {
				upNodes := nodeMap[origNode.Up[0]]
				if len(upNodes) > 0 {
					cn.Up = append(cn.Up, upNodes[0])
				}
			}

			if origNode.LeftDown != nil {
				downNodes := nodeMap[origNode.LeftDown]
				if len(downNodes) > 0 {

					cn.Down = append(cn.Down, downNodes[0])
				}
			}
		} else if len(origNode.People) == 2 {

			cn1, cn2 := coordNodes[0], coordNodes[1]

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

			if origNode.LeftDown != nil {

				collectChildren(origNode, nodeMap, cn1, cn2)
			}
		} else if len(origNode.People) == 1 {

			cn := coordNodes[0]

			if len(origNode.Up) > 0 && origNode.Up[0] != nil {
				upNodes := nodeMap[origNode.Up[0]]
				if len(upNodes) > 0 {
					cn.Up = append(cn.Up, upNodes[0])
				}
			}

			if origNode.LeftDown != nil {
				collectChildrenSingle(origNode, nodeMap, cn)
			}
		}
	}

	setupParentChildLinks(om, nodeMap, orderedOrigNodes)

	cm := assignCoordinates(om, nodeMap)

	mergeBackNodes(cm, nodeMap, orderedOrigNodes)

	buildPersonToNode(cm)

	optimizePositions(cm)

	splitMergedNodes(cm)

	buildPersonToNode(cm)

	normalizeCoordinates(cm)

	return cm
}

func setupParentChildLinks(om *OrderManager, nodeMap map[*LayerNode][]*CoordNode, orderedOrigNodes []*LayerNode) {

	for _, origNode := range orderedOrigNodes {
		coordNodes := nodeMap[origNode]
		for _, cn := range coordNodes {
			cn.AddedLeft = origNode.AddedLeft
		}
	}

	for _, origNode := range orderedOrigNodes {
		coordNodes := nodeMap[origNode]
		if origNode.IsPseudo {
			continue
		}

		for _, cn := range coordNodes {

			cn.ParentNodes = make([]*CoordNode, len(cn.People))
			cn.ParentPersonIndex = make([]int, len(cn.People))
			cn.Children = []*CoordNode{}

			for i := range cn.People {
				if i < len(cn.Up) && cn.Up[i] != nil {
					parentCN := cn.Up[i]

					for parentCN != nil && parentCN.IsPseudo {
						if len(parentCN.Up) > 0 {
							parentCN = parentCN.Up[0]
						} else {
							parentCN = nil
						}
					}
					cn.ParentNodes[i] = parentCN

					if parentCN != nil {
						cn.ParentPersonIndex[i] = findPersonIndexInParent(cn.People[i], parentCN, origNode, nodeMap)
					}
				}
			}
		}
	}

	for _, origNode := range orderedOrigNodes {
		coordNodes := nodeMap[origNode]
		for _, cn := range coordNodes {
			for i, parentCN := range cn.ParentNodes {
				if parentCN != nil {

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
					_ = i
				}
			}
		}
	}
}

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

func findPersonIndexInParent(child *stage1_input.Person, parentCN *CoordNode, childOrigNode *LayerNode, nodeMap map[*LayerNode][]*CoordNode) int {

	if len(parentCN.People) == 1 {
		return 0
	}

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

func collectChildren(parentNode *LayerNode, nodeMap map[*LayerNode][]*CoordNode, cn1, cn2 *CoordNode) {

	current := parentNode.LeftDown
	for current != nil {
		if childNodes, ok := nodeMap[current]; ok {
			for _, childCN := range childNodes {

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

func collectChildrenSingle(parentNode *LayerNode, nodeMap map[*LayerNode][]*CoordNode, cn *CoordNode) {

	current := parentNode.LeftDown
	for current != nil {
		if childNodes, ok := nodeMap[current]; ok {
			for _, childCN := range childNodes {

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

func assignCoordinates(om *OrderManager, nodeMap map[*LayerNode][]*CoordNode) *CoordMatrix {
	cm := NewCoordMatrix(om.minLayer, om.maxLayer)

	globalCoord := 0
	const coordStep = 20

	sortedLayers := om.getSortedLayerNumbers()

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

	layerIndices := make(map[int]int)
	for _, layerNum := range sortedLayers {
		layerIndices[layerNum] = 0
	}

	maxNodesInLayer := 0
	for _, nodes := range layerNodes {
		if len(nodes) > maxNodesInLayer {
			maxNodesInLayer = len(nodes)
		}
	}

	for columnIdx := 0; columnIdx < maxNodesInLayer; columnIdx++ {

		for _, layerNum := range sortedLayers {
			idx := layerIndices[layerNum]
			nodes := layerNodes[layerNum]

			if idx >= len(nodes) {
				continue
			}

			cn := nodes[idx]

			if cn.Left != 0 || cn.Right != 0 {
				layerIndices[layerNum]++
				continue
			}

			if cn.IsPseudo {

				cn.Left = globalCoord
				cn.Right = globalCoord
			} else {

				cn.Left = globalCoord
				cn.Right = globalCoord + 2
			}

			cm.AddNode(cn)
			layerIndices[layerNum]++
		}

		globalCoord += coordStep
	}

	return cm
}

func (om *OrderManager) getSortedLayerNumbers() []int {
	layers := []int{}
	for layerNum := om.minLayer; layerNum <= om.maxLayer; layerNum++ {
		layers = append(layers, layerNum)
	}
	return layers
}

func mergeBackNodes(cm *CoordMatrix, nodeMap map[*LayerNode][]*CoordNode, orderedOrigNodes []*LayerNode) {

	merged := make(map[*CoordNode]bool)

	for _, origNode := range orderedOrigNodes {
		coordNodes := nodeMap[origNode]
		if len(coordNodes) == 2 && coordNodes[0].WasMerged && coordNodes[1].WasMerged {
			cn1, cn2 := coordNodes[0], coordNodes[1]
			if merged[cn1] || merged[cn2] {
				continue
			}

			if cn1.Left == 0 && cn1.Right == 0 {
				fmt.Printf("WARN: cn1 РЅРµ СЂР°Р·РјРµС‰РµРЅР°: %v\n", cn1.People)
				continue
			}
			if cn2.Left == 0 && cn2.Right == 0 {
				fmt.Printf("WARN: cn2 РЅРµ СЂР°Р·РјРµС‰РµРЅР°: %v\n", cn2.People)
				continue
			}

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

			minLeft := cn1.Left
			if cn2.Left < minLeft {
				minLeft = cn2.Left
			}
			maxRight := cn1.Right
			if cn2.Right > maxRight {
				maxRight = cn2.Right
			}

			center := (minLeft + maxRight) / 2
			mergedNode.Left = center - 2
			mergedNode.Right = center + 2

			mergedNode.Up = append(mergedNode.Up, cn1.Up...)
			mergedNode.Up = append(mergedNode.Up, cn2.Up...)
			mergedNode.Down = append(mergedNode.Down, cn1.Down...)
			mergedNode.Down = append(mergedNode.Down, cn2.Down...)

			mergedNode.ParentNodes = append(mergedNode.ParentNodes, cn1.ParentNodes...)
			mergedNode.ParentNodes = append(mergedNode.ParentNodes, cn2.ParentNodes...)
			mergedNode.ParentPersonIndex = append(mergedNode.ParentPersonIndex, cn1.ParentPersonIndex...)
			mergedNode.ParentPersonIndex = append(mergedNode.ParentPersonIndex, cn2.ParentPersonIndex...)

			mergedNode.Children = append(mergedNode.Children, cn1.Children...)
			mergedNode.Children = append(mergedNode.Children, cn2.Children...)

			uniqueDown := []*CoordNode{}
			downSet := make(map[*CoordNode]bool)
			for _, d := range mergedNode.Down {
				if !downSet[d] {
					downSet[d] = true
					uniqueDown = append(uniqueDown, d)
				}
			}
			mergedNode.Down = uniqueDown

			uniqueChildren := []*CoordNode{}
			childSet := make(map[*CoordNode]bool)
			for _, c := range mergedNode.Children {
				if !childSet[c] {
					childSet[c] = true
					uniqueChildren = append(uniqueChildren, c)
				}
			}
			mergedNode.Children = uniqueChildren

			for _, child := range mergedNode.Children {
				if child == nil {
					continue
				}

				for i, parentCN := range child.ParentNodes {
					if parentCN == cn1 || parentCN == cn2 {
						child.ParentNodes[i] = mergedNode
					}
				}
			}

			for _, parentCN := range mergedNode.ParentNodes {
				if parentCN == nil {
					continue
				}

				for i, childCN := range parentCN.Children {
					if childCN == cn1 || childCN == cn2 {
						parentCN.Children[i] = mergedNode
					}
				}

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

			if len(cn1.Up) > 0 && cn1.Up[0] != nil && cn1.Up[0].IsPseudo {
				movePseudoChainToCoord(cn1.Up[0], mergedNode.Left+1)
			}
			if len(cn2.Up) > 0 && cn2.Up[0] != nil && cn2.Up[0].IsPseudo {
				movePseudoChainToCoord(cn2.Up[0], mergedNode.Right-1)
			}

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

func splitMergedNodes(cm *CoordMatrix) {
	for layerNum := cm.MinLayer; layerNum <= cm.MaxLayer; layerNum++ {
		newLayer := []*CoordNode{}

		for _, node := range cm.Layers[layerNum] {

			if len(node.People) == 2 && node.Width() == 4 {

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

				cn1.MergePartner = cn2
				cn2.MergePartner = cn1

				if len(node.Up) > 0 {
					cn1.Up = append(cn1.Up, node.Up[0])
				}
				if len(node.Up) > 1 {
					cn2.Up = append(cn2.Up, node.Up[1])
				}

				cn1.Children = append(cn1.Children, node.Children...)
				cn2.Children = append(cn2.Children, node.Children...)

				if len(node.ParentNodes) > 0 {
					cn1.ParentNodes = append(cn1.ParentNodes, node.ParentNodes[0])
				}
				if len(node.ParentNodes) > 1 {
					cn2.ParentNodes = append(cn2.ParentNodes, node.ParentNodes[1])
				}

				for _, child := range node.Children {
					if child == nil {
						continue
					}
					for i, parentCN := range child.ParentNodes {
						if parentCN == node {

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

func normalizeCoordinates(cm *CoordMatrix) {

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

	if minCoord != 0 {
		for layerNum := cm.MinLayer; layerNum <= cm.MaxLayer; layerNum++ {
			for _, node := range cm.Layers[layerNum] {
				node.Left -= minCoord
				node.Right -= minCoord
			}
		}
	}
}

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
