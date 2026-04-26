package stage3_ordering

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

func (om *OrderManager) isLeftOf(node, target *LayerNode) bool {
	if node == target {
		return true
	}
	for curr := node.Next; curr != nil && !curr.IsTail(); curr = curr.Next {
		if curr == target {
			return true
		}
	}
	return false
}

func (om *OrderManager) isRightOf(node, target *LayerNode) bool {
	if node == target {
		return true
	}
	for curr := node.Prev; curr != nil && !curr.IsHead(); curr = curr.Prev {
		if curr == target {
			return true
		}
	}
	return false
}

func (om *OrderManager) followPointerToLayer(node *LayerNode, targetLayer int, goingUp bool) *LayerNode {
	current := node
	for current != nil && current.Layer != targetLayer {
		if goingUp {

			next := current.GetRightUp()
			if next == nil {
				next = current.GetLeftUp()
			}
			current = next
		} else {
			next := current.RightDown
			if next == nil {
				next = current.LeftDown
			}
			current = next
		}
	}
	return current
}

func (om *OrderManager) findRightmostReaching(layer, targetLayer int, targetNode *LayerNode, goingUp bool) *LayerNode {
	layerObj := om.Layers[layer]
	if layerObj == nil {
		return nil
	}

	var result *LayerNode

	for node := layerObj.Head.Next; node != nil && !node.IsTail(); node = node.Next {

		reachable := om.followPointerToLayer(node, targetLayer, goingUp)

		if reachable == nil {
			continue
		}

		if om.isLeftOf(reachable, targetNode) {
			result = node
		}
	}

	return result
}

func (om *OrderManager) findLeftmostReaching(layer, targetLayer int, targetNode *LayerNode, goingUp bool) *LayerNode {
	layerObj := om.Layers[layer]
	if layerObj == nil {
		return nil
	}

	for node := layerObj.Head.Next; node != nil && !node.IsTail(); node = node.Next {

		reachable := om.followPointerToLayer(node, targetLayer, goingUp)

		if reachable == nil {
			continue
		}

		if om.isRightOf(reachable, targetNode) {
			return node
		}
	}

	return nil
}

func (om *OrderManager) AddPersonRight(fromNode *LayerNode, addedPerson *stage1_input.Person, fromLayer, targetLayer int, fromPersonIndex int) *LayerNode {
	om.ensureLayer(targetLayer)

	if targetLayer == fromLayer {

		return om.addPartnerRight(fromNode, addedPerson, fromLayer)
	}

	goingUp := targetLayer > fromLayer

	var prevNode *LayerNode = fromNode
	var step int
	var startLayer, endLayer int

	if goingUp {
		step = 1
		startLayer = fromLayer + 1
		endLayer = targetLayer
	} else {
		step = -1
		startLayer = fromLayer - 1
		endLayer = targetLayer
	}

	for layer := startLayer; ; layer += step {
		om.ensureLayer(layer)

		var newNode *LayerNode
		if layer == endLayer {
			newNode = om.CreatePersonNode(addedPerson, layer)
			newNode.AddedLeft = false
		} else {
			newNode = om.CreatePseudoNode(layer)
		}

		insertAfter := om.findRightmostReaching(layer, layer-step, prevNode, !goingUp)

		if insertAfter == nil {

			layerObj := om.Layers[layer]
			insertAfter = layerObj.Head
		}

		om.insertAfter(insertAfter, newNode)

		if goingUp {

			newNode.LeftDown = prevNode
			newNode.RightDown = prevNode

			if prevNode == fromNode && len(prevNode.People) > 0 {

				for len(prevNode.Up) <= fromPersonIndex {
					prevNode.Up = append(prevNode.Up, nil)
				}
				prevNode.Up[fromPersonIndex] = newNode
			} else {

				prevNode.Up = append(prevNode.Up, newNode)
			}
		} else {

			newNode.Up = []*LayerNode{prevNode}

			if prevNode.RightDown == nil {
				prevNode.RightDown = newNode
			}
			if prevNode.LeftDown == nil {
				prevNode.LeftDown = newNode
			}
		}

		prevNode = newNode

		if layer == endLayer {
			return newNode
		}
	}
}

func (om *OrderManager) AddPersonLeft(fromNode *LayerNode, addedPerson *stage1_input.Person, fromLayer, targetLayer int, fromPersonIndex int) *LayerNode {
	om.ensureLayer(targetLayer)

	if targetLayer == fromLayer {

		return om.addPartnerLeft(fromNode, addedPerson, fromLayer)
	}

	goingUp := targetLayer > fromLayer

	var prevNode *LayerNode = fromNode
	var step int
	var startLayer, endLayer int

	if goingUp {
		step = 1
		startLayer = fromLayer + 1
		endLayer = targetLayer
	} else {
		step = -1
		startLayer = fromLayer - 1
		endLayer = targetLayer
	}

	for layer := startLayer; ; layer += step {
		om.ensureLayer(layer)

		var newNode *LayerNode
		if layer == endLayer {
			newNode = om.CreatePersonNode(addedPerson, layer)
			newNode.AddedLeft = true
		} else {
			newNode = om.CreatePseudoNode(layer)
		}

		insertBefore := om.findLeftmostReaching(layer, layer-step, prevNode, !goingUp)

		if insertBefore == nil {

			layerObj := om.Layers[layer]
			insertBefore = layerObj.Tail
		}

		om.insertBefore(insertBefore, newNode)

		if goingUp {
			newNode.LeftDown = prevNode
			newNode.RightDown = prevNode

			if prevNode == fromNode && len(prevNode.People) > 0 {

				for len(prevNode.Up) <= fromPersonIndex {
					prevNode.Up = append(prevNode.Up, nil)
				}
				prevNode.Up[fromPersonIndex] = newNode
			} else {

				prevNode.Up = append(prevNode.Up, newNode)
			}
		} else {

			newNode.Up = []*LayerNode{prevNode}

			if prevNode.LeftDown == nil {
				prevNode.LeftDown = newNode
			}
			if prevNode.RightDown == nil {
				prevNode.RightDown = newNode
			}
		}

		prevNode = newNode

		if layer == endLayer {
			return newNode
		}
	}
}

func (om *OrderManager) AddParentPairRight(fromNode *LayerNode, parent1, parent2 *stage1_input.Person, fromLayer, targetLayer int, siblingUp *LayerNode, fromPersonIndex int) *LayerNode {
	om.ensureLayer(targetLayer)

	var prevNode *LayerNode = fromNode
	step := 1
	startLayer := fromLayer + 1
	endLayer := targetLayer
	isFirstStep := true

	for layer := startLayer; ; layer += step {
		om.ensureLayer(layer)

		var newNode *LayerNode
		if layer == endLayer {

			newNode = om.CreatePairNode(parent1, parent2, layer)
		} else {
			newNode = om.CreatePseudoNode(layer)
		}

		var insertAfter *LayerNode

		if isFirstStep && siblingUp != nil && siblingUp.Layer == layer {

			if fromPersonIndex == 1 {

				insertAfter = siblingUp
			} else {

				if siblingUp.Prev != nil && !siblingUp.Prev.IsHead() {
					insertAfter = siblingUp.Prev
				} else {
					insertAfter = om.Layers[layer].Head
				}
			}
		} else {
			insertAfter = om.findRightmostReaching(layer, layer-step, prevNode, false)
			if insertAfter == nil {
				layerObj := om.Layers[layer]
				insertAfter = layerObj.Head
			}
		}

		om.insertAfter(insertAfter, newNode)

		newNode.LeftDown = prevNode
		newNode.RightDown = prevNode

		if isFirstStep && len(prevNode.People) > 0 {

			for len(prevNode.Up) <= fromPersonIndex {
				prevNode.Up = append(prevNode.Up, nil)
			}
			prevNode.Up[fromPersonIndex] = newNode
		} else {

			prevNode.Up = append(prevNode.Up, newNode)
		}

		isFirstStep = false
		prevNode = newNode

		if layer == endLayer {
			return newNode
		}
	}
}

func (om *OrderManager) AddParentPairLeft(fromNode *LayerNode, parent1, parent2 *stage1_input.Person, fromLayer, targetLayer int, siblingUp *LayerNode, fromPersonIndex int) *LayerNode {
	om.ensureLayer(targetLayer)

	var prevNode *LayerNode = fromNode
	step := 1
	startLayer := fromLayer + 1
	endLayer := targetLayer
	isFirstStep := true

	for layer := startLayer; ; layer += step {
		om.ensureLayer(layer)

		var newNode *LayerNode
		if layer == endLayer {

			newNode = om.CreatePairNode(parent1, parent2, layer)
		} else {
			newNode = om.CreatePseudoNode(layer)
		}

		var insertBefore *LayerNode

		if isFirstStep && siblingUp != nil && siblingUp.Layer == layer {

			if fromPersonIndex == 0 {

				insertBefore = siblingUp
			} else {

				if siblingUp.Next != nil && !siblingUp.Next.IsTail() {
					insertBefore = siblingUp.Next
				} else {
					insertBefore = om.Layers[layer].Tail
				}
			}
		} else {
			insertBefore = om.findLeftmostReaching(layer, layer-step, prevNode, false)
			if insertBefore == nil {
				layerObj := om.Layers[layer]
				insertBefore = layerObj.Tail
			}
		}

		om.insertBefore(insertBefore, newNode)

		newNode.LeftDown = prevNode
		newNode.RightDown = prevNode

		if isFirstStep && len(prevNode.People) > 0 {

			for len(prevNode.Up) <= fromPersonIndex {
				prevNode.Up = append(prevNode.Up, nil)
			}
			prevNode.Up[fromPersonIndex] = newNode
		} else {

			prevNode.Up = append(prevNode.Up, newNode)
		}

		isFirstStep = false
		prevNode = newNode

		if layer == endLayer {
			return newNode
		}
	}
}

func (om *OrderManager) collectPartnerCluster(startNode *LayerNode) []*LayerNode {
	visited := make(map[*LayerNode]bool)
	var cluster []*LayerNode

	var dfs func(node *LayerNode)
	dfs = func(node *LayerNode) {
		if node == nil || visited[node] {
			return
		}
		visited[node] = true
		cluster = append(cluster, node)

		for _, person := range node.People {
			for _, partner := range person.Partners {
				partnerNode := om.PersonNodes[partner.ID]
				if partnerNode != nil && partnerNode.Layer == node.Layer {
					dfs(partnerNode)
				}
			}
		}
	}

	dfs(startNode)
	return cluster
}

func (om *OrderManager) addPartnerRight(fromNode *LayerNode, addedPerson *stage1_input.Person, layer int) *LayerNode {

	if len(fromNode.People) == 1 {

		om.AddPersonToExistingNode(fromNode, addedPerson, "right")
		return fromNode
	}

	cluster := om.collectPartnerCluster(fromNode)

	var rightmost *LayerNode = fromNode
	for _, node := range cluster {
		if om.isRightOf(node, rightmost) && node != rightmost {

			rightmost = node
		}
	}

	for curr := rightmost.Next; curr != nil && !curr.IsTail(); curr = curr.Next {
		inCluster := false
		for _, node := range cluster {
			if node == curr {
				inCluster = true
				break
			}
		}
		if inCluster {
			rightmost = curr
		} else {
			break
		}
	}

	newNode := om.CreatePersonNode(addedPerson, layer)
	newNode.AddedLeft = false
	om.insertAfter(rightmost, newNode)

	return newNode
}

func (om *OrderManager) addPartnerLeft(fromNode *LayerNode, addedPerson *stage1_input.Person, layer int) *LayerNode {

	if len(fromNode.People) == 1 {

		om.AddPersonToExistingNode(fromNode, addedPerson, "left")
		return fromNode
	}

	cluster := om.collectPartnerCluster(fromNode)

	var leftmost *LayerNode = fromNode
	for _, node := range cluster {
		if om.isLeftOf(node, leftmost) && node != leftmost {
			leftmost = node
		}
	}

	for curr := leftmost.Prev; curr != nil && !curr.IsHead(); curr = curr.Prev {
		inCluster := false
		for _, node := range cluster {
			if node == curr {
				inCluster = true
				break
			}
		}
		if inCluster {
			leftmost = curr
		} else {
			break
		}
	}

	newNode := om.CreatePersonNode(addedPerson, layer)
	newNode.AddedLeft = true
	om.insertBefore(leftmost, newNode)

	return newNode
}
