package stage3_ordering

func calculateTotalDistance(node *CoordNode, cm *CoordMatrix) int {
	parentChildTotal := 0

	for i, parentCN := range node.ParentNodes {
		if parentCN == nil {
			continue
		}

		parentPersonIdx := 0
		if i < len(node.ParentPersonIndex) {
			parentPersonIdx = node.ParentPersonIndex[i]
		}
		parentCoord := getParentCoordinateForPerson(parentCN, parentPersonIdx)
		childCoord := getChildCoordinate(node, i)
		dist := parentCoord - childCoord
		if dist < 0 {
			dist = -dist
		}

		crossPenalty := calculateCrossingPenalty(parentCN, parentCoord, childCoord, cm)
		adjustedDist := dist + crossPenalty*10

		parentChildTotal += adjustedDist * adjustedDist
	}

	for _, childCN := range node.Children {
		if childCN == nil {
			continue
		}

		childIdx := -1
		parentPersonIdx := 0
		for i, pn := range childCN.ParentNodes {
			if pn == node {
				childIdx = i
				if i < len(childCN.ParentPersonIndex) {
					parentPersonIdx = childCN.ParentPersonIndex[i]
				}
				break
			}
		}
		if childIdx >= 0 {
			parentCoord := getParentCoordinateForPerson(node, parentPersonIdx)
			childCoord := getChildCoordinate(childCN, childIdx)
			dist := parentCoord - childCoord
			if dist < 0 {
				dist = -dist
			}

			crossPenalty := calculateCrossingPenalty(node, parentCoord, childCoord, cm)
			adjustedDist := dist + crossPenalty*10

			parentChildTotal += adjustedDist * adjustedDist
		}
	}

	partnerTotal := calculatePartnerDistance(node, cm)

	return 5*parentChildTotal + partnerTotal
}

func calculatePartnerDistance(node *CoordNode, cm *CoordMatrix) int {
	if cm == nil || cm.PersonToNode == nil {
		return 0
	}

	total := 0
	processedPartners := make(map[int]bool)

	for _, person := range node.People {
		for _, partner := range person.Partners {

			if processedPartners[partner.ID] {
				continue
			}
			processedPartners[partner.ID] = true

			partnerNode := cm.PersonToNode[partner.ID]
			if partnerNode == nil {
				continue
			}

			if partnerNode == node {
				continue
			}

			if node.MergePartner != nil && partnerNode == node.MergePartner {
				continue
			}

			var leftNode, rightNode *CoordNode
			if node.Left < partnerNode.Left {
				leftNode = node
				rightNode = partnerNode
			} else {
				leftNode = partnerNode
				rightNode = node
			}

			dist := rightNode.Left - leftNode.Right
			if dist < 0 {
				dist = 0
			}

			total += dist * dist
		}
	}

	return total
}

func calculateCrossingPenalty(parentNode *CoordNode, parentCoord, childCoord int, cm *CoordMatrix) int {
	if cm == nil {
		return 0
	}

	parentLayer := parentNode.Layer

	myLeft := parentCoord
	myRight := childCoord
	if myLeft > myRight {
		myLeft, myRight = myRight, myLeft
	}
	myLeft -= 1
	myRight += 1

	totalPenalty := 0

	for _, otherParent := range cm.Layers[parentLayer] {
		if otherParent == parentNode {
			continue
		}
		if otherParent.IsPseudo {
			continue
		}

		for _, otherChild := range otherParent.Children {
			if otherChild == nil {
				continue
			}

			childIdx := -1
			parentPersonIdx := 0
			for i, pn := range otherChild.ParentNodes {
				if pn == otherParent {
					childIdx = i
					if i < len(otherChild.ParentPersonIndex) {
						parentPersonIdx = otherChild.ParentPersonIndex[i]
					}
					break
				}
			}
			if childIdx < 0 {
				continue
			}

			otherParentCoord := getParentCoordinateForPerson(otherParent, parentPersonIdx)
			otherChildCoord := getChildCoordinate(otherChild, childIdx)

			otherLeft := otherParentCoord
			otherRight := otherChildCoord
			if otherLeft > otherRight {
				otherLeft, otherRight = otherRight, otherLeft
			}

			overlapLen := segmentOverlap(myLeft, myRight, otherLeft, otherRight)
			totalPenalty += overlapLen
		}
	}

	return totalPenalty
}

func segmentOverlap(a1, a2, b1, b2 int) int {

	if a1 > a2 {
		a1, a2 = a2, a1
	}
	if b1 > b2 {
		b1, b2 = b2, b1
	}

	start := a1
	if b1 > start {
		start = b1
	}
	end := a2
	if b2 < end {
		end = b2
	}

	if start >= end {
		return 0
	}
	return end - start
}

func getParentCoordinateForPerson(parent *CoordNode, personIndex int) int {
	if parent == nil {
		return 0
	}

	if parent.MergePartner != nil {
		left := parent.Left
		right := parent.Right
		if parent.MergePartner.Left < left {
			left = parent.MergePartner.Left
		}
		if parent.MergePartner.Right > right {
			right = parent.MergePartner.Right
		}
		return (left + right) / 2
	}

	return (parent.Left + parent.Right) / 2
}

func getParentCoordinate(parent *CoordNode) int {

	if len(parent.People) == 2 {
		return (parent.Left + parent.Right) / 2
	}

	if parent.AddedLeft {
		return parent.Right
	}
	return parent.Left
}

func getChildCoordinate(child *CoordNode, personIndex int) int {

	if len(child.People) == 1 {
		return (child.Left + child.Right) / 2
	}

	if personIndex == 0 {
		return child.Left + 1
	}
	return child.Right - 1
}

func collectPseudoChain(start *CoordNode) []*CoordNode {
	chain := []*CoordNode{}
	visited := make(map[*CoordNode]bool)

	current := start
	for current != nil && !visited[current] {
		visited[current] = true
		chain = append([]*CoordNode{current}, chain...)
		if len(current.Down) > 0 && current.Down[0] != nil {
			next := current.Down[0]
			if next.IsPseudo || current.IsPseudo {
				current = next
			} else {

				if !visited[next] {
					chain = append([]*CoordNode{next}, chain...)
					visited[next] = true
				}
				break
			}
		} else {
			break
		}
	}

	current = start
	for len(current.Up) > 0 && current.Up[0] != nil && !visited[current.Up[0]] {
		current = current.Up[0]
		visited[current] = true
		chain = append(chain, current)
	}

	return chain
}

func placeChainAtCoord(cn *CoordNode, coord int, cm *CoordMatrix, chainMap map[*CoordNode]*int, layerPaused map[int]bool) {
	chain := collectPseudoChain(cn)

	for _, node := range chain {
		placeNodeAtCoord(node, coord, cm)
		layerPaused[node.Layer] = false
		delete(chainMap, node)
	}
}

func placeNodeAtCoord(cn *CoordNode, coord int, cm *CoordMatrix) {
	if cn.Left != 0 || cn.Right != 0 {

		return
	}

	if cn.IsPseudo {

		cn.Left = coord
		cn.Right = coord
	} else {

		cn.Left = coord
		cn.Right = coord + 2
	}

	cm.AddNode(cn)
}

func placeChain(cn *CoordNode, coord int, cm *CoordMatrix, chainMap map[*CoordNode]*int, layerPaused map[int]bool) {
	placeChainAtCoord(cn, coord, cm, chainMap, layerPaused)
}

func placeNode(cn *CoordNode, coord int, cm *CoordMatrix) {
	placeNodeAtCoord(cn, coord, cm)
}
