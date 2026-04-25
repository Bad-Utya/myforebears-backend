package stage3_ordering

// calculateTotalDistance РІС‹С‡РёСЃР»СЏРµС‚ СЃСѓРјРјСѓ РљР’РђР”Р РђРўРћР’ СЂР°СЃСЃС‚РѕСЏРЅРёР№ РґРѕ СЂРѕРґРёС‚РµР»РµР№, РґРµС‚РµР№ Рё РїР°СЂС‚РЅС‘СЂРѕРІ
// Р¤РѕСЂРјСѓР»Р°: 5*(СЂР°СЃСЃС‚РѕСЏРЅРёСЏ РґРѕ СЂРѕРґРёС‚РµР»РµР№ Рё РґРµС‚РµР№) + (СЂР°СЃСЃС‚РѕСЏРЅРёСЏ РґРѕ РїР°СЂС‚РЅС‘СЂРѕРІ РёР· РґСЂСѓРіРёС… РІРµСЂС€РёРЅ)
func calculateTotalDistance(node *CoordNode, cm *CoordMatrix) int {
	parentChildTotal := 0

	// Р Р°СЃСЃС‚РѕСЏРЅРёСЏ РґРѕ СЂРѕРґРёС‚РµР»РµР№ (СЌС‚Р° РІРµСЂС€РёРЅР° РєР°Рє СЂРµР±С‘РЅРѕРє)
	for i, parentCN := range node.ParentNodes {
		if parentCN == nil {
			continue
		}
		// РћРїСЂРµРґРµР»СЏРµРј РёРЅРґРµРєСЃ С‡РµР»РѕРІРµРєР° РІ СЂРѕРґРёС‚РµР»СЊСЃРєРѕР№ РІРµСЂС€РёРЅРµ
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

		// РЎС‡РёС‚Р°РµРј С€С‚СЂР°С„ Р·Р° РїРµСЂРµСЃРµС‡РµРЅРёСЏ СЃ РґСЂСѓРіРёРјРё Р»РёРЅРёСЏРјРё СЂРѕРґРёС‚РµР»СЊ-СЂРµР±С‘РЅРѕРє
		crossPenalty := calculateCrossingPenalty(parentCN, parentCoord, childCoord, cm)
		adjustedDist := dist + crossPenalty*10

		parentChildTotal += adjustedDist * adjustedDist
	}

	// Р Р°СЃСЃС‚РѕСЏРЅРёСЏ РґРѕ РґРµС‚РµР№ (СЌС‚Р° РІРµСЂС€РёРЅР° РєР°Рє СЂРѕРґРёС‚РµР»СЊ)
	for _, childCN := range node.Children {
		if childCN == nil {
			continue
		}
		// РќР°С…РѕРґРёРј РёРЅРґРµРєСЃ СЌС‚РѕРіРѕ СЂРѕРґРёС‚РµР»СЏ РІ СЂРµР±С‘РЅРєРµ Рё РёРЅРґРµРєСЃ С‡РµР»РѕРІРµРєР° РІ СЂРѕРґРёС‚РµР»Рµ
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

			// РЎС‡РёС‚Р°РµРј С€С‚СЂР°С„ Р·Р° РїРµСЂРµСЃРµС‡РµРЅРёСЏ СЃ РґСЂСѓРіРёРјРё Р»РёРЅРёСЏРјРё СЂРѕРґРёС‚РµР»СЊ-СЂРµР±С‘РЅРѕРє
			crossPenalty := calculateCrossingPenalty(node, parentCoord, childCoord, cm)
			adjustedDist := dist + crossPenalty*10

			parentChildTotal += adjustedDist * adjustedDist
		}
	}

	// Р Р°СЃСЃС‚РѕСЏРЅРёСЏ РґРѕ РїР°СЂС‚РЅС‘СЂРѕРІ РёР· РґСЂСѓРіРёС… РІРµСЂС€РёРЅ (РЅРµ MergePartner)
	partnerTotal := calculatePartnerDistance(node, cm)

	// Р¤РѕСЂРјСѓР»Р°: 5*(СЂРѕРґРёС‚РµР»Рё+РґРµС‚Рё) + РїР°СЂС‚РЅС‘СЂС‹
	return 5*parentChildTotal + partnerTotal
}

// calculatePartnerDistance РІС‹С‡РёСЃР»СЏРµС‚ СЃСѓРјРјСѓ РєРІР°РґСЂР°С‚РѕРІ СЂР°СЃСЃС‚РѕСЏРЅРёР№ РґРѕ РїР°СЂС‚РЅС‘СЂРѕРІ РёР· РґСЂСѓРіРёС… РІРµСЂС€РёРЅ
// РќРµ СЃС‡РёС‚Р°РµРј СЂР°СЃСЃС‚РѕСЏРЅРёРµ РјРµР¶РґСѓ РїР°СЂС‚РЅС‘СЂР°РјРё РёР· РѕРґРЅРѕР№ СЃРјРµР¶РЅРѕР№ РІРµСЂС€РёРЅС‹ (MergePartner)
func calculatePartnerDistance(node *CoordNode, cm *CoordMatrix) int {
	if cm == nil || cm.PersonToNode == nil {
		return 0
	}

	total := 0
	processedPartners := make(map[int]bool) // С‡С‚РѕР±С‹ РЅРµ СЃС‡РёС‚Р°С‚СЊ РѕРґРЅРѕРіРѕ РїР°СЂС‚РЅС‘СЂР° РґРІР°Р¶РґС‹

	// Р”Р»СЏ РєР°Р¶РґРѕРіРѕ С‡РµР»РѕРІРµРєР° РІ РІРµСЂС€РёРЅРµ РёС‰РµРј РµРіРѕ РїР°СЂС‚РЅС‘СЂРѕРІ
	for _, person := range node.People {
		for _, partner := range person.Partners {
			// РџСЂРѕРїСѓСЃРєР°РµРј СѓР¶Рµ РѕР±СЂР°Р±РѕС‚Р°РЅРЅС‹С…
			if processedPartners[partner.ID] {
				continue
			}
			processedPartners[partner.ID] = true

			// РќР°С…РѕРґРёРј РІРµСЂС€РёРЅСѓ РїР°СЂС‚РЅС‘СЂР°
			partnerNode := cm.PersonToNode[partner.ID]
			if partnerNode == nil {
				continue
			}

			// РџСЂРѕРїСѓСЃРєР°РµРј РµСЃР»Рё РїР°СЂС‚РЅС‘СЂ РІ С‚РѕР№ Р¶Рµ РІРµСЂС€РёРЅРµ
			if partnerNode == node {
				continue
			}

			// РџСЂРѕРїСѓСЃРєР°РµРј РµСЃР»Рё РїР°СЂС‚РЅС‘СЂ вЂ” MergePartner (СЃРјРµР¶РЅР°СЏ РІРµСЂС€РёРЅР°)
			if node.MergePartner != nil && partnerNode == node.MergePartner {
				continue
			}

			// Р’С‹С‡РёСЃР»СЏРµРј СЂР°СЃСЃС‚РѕСЏРЅРёРµ РјРµР¶РґСѓ РїР°СЂС‚РЅС‘СЂР°РјРё
			// РћРїСЂРµРґРµР»СЏРµРј РєС‚Рѕ СЃР»РµРІР°, РєС‚Рѕ СЃРїСЂР°РІР°
			var leftNode, rightNode *CoordNode
			if node.Left < partnerNode.Left {
				leftNode = node
				rightNode = partnerNode
			} else {
				leftNode = partnerNode
				rightNode = node
			}

			// Р Р°СЃСЃС‚РѕСЏРЅРёРµ РѕС‚ РїСЂР°РІРѕР№ РіСЂР°РЅРёС†С‹ Р»РµРІРѕРіРѕ РґРѕ Р»РµРІРѕР№ РіСЂР°РЅРёС†С‹ РїСЂР°РІРѕРіРѕ
			dist := rightNode.Left - leftNode.Right
			if dist < 0 {
				dist = 0 // РµСЃР»Рё РІРµСЂС€РёРЅС‹ РїРµСЂРµРєСЂС‹РІР°СЋС‚СЃСЏ, СЂР°СЃСЃС‚РѕСЏРЅРёРµ = 0
			}

			total += dist * dist
		}
	}

	return total
}

// calculateCrossingPenalty РІС‹С‡РёСЃР»СЏРµС‚ СЃСѓРјРјР°СЂРЅСѓСЋ РґР»РёРЅСѓ РїРµСЂРµСЃРµС‡РµРЅРёР№ РѕС‚СЂРµР·РєР° [parentCoord, childCoord]
// СЃ РѕС‚СЂРµР·РєР°РјРё РґСЂСѓРіРёС… СЃРІСЏР·РµР№ СЂРѕРґРёС‚РµР»СЊ-СЂРµР±С‘РЅРѕРє РЅР° С‚РѕРј Р¶Рµ СЃР»РѕРµ СЂРѕРґРёС‚РµР»СЏ.
// РћС‚СЂРµР·РѕРє С‚РµРєСѓС‰РµР№ РІРµСЂС€РёРЅС‹ СЂР°СЃС€РёСЂСЏРµС‚СЃСЏ РЅР° 1 РІР»РµРІРѕ Рё 1 РІРїСЂР°РІРѕ РїРµСЂРµРґ РїРѕРґСЃС‡С‘С‚РѕРј РїРµСЂРµСЃРµС‡РµРЅРёР№.
func calculateCrossingPenalty(parentNode *CoordNode, parentCoord, childCoord int, cm *CoordMatrix) int {
	if cm == nil {
		return 0
	}

	parentLayer := parentNode.Layer

	// РћС‚СЂРµР·РѕРє С‚РµРєСѓС‰РµР№ СЃРІСЏР·Рё (СЂР°СЃС€РёСЂРµРЅРЅС‹Р№ РЅР° 1 РІ РѕР±Рµ СЃС‚РѕСЂРѕРЅС‹)
	myLeft := parentCoord
	myRight := childCoord
	if myLeft > myRight {
		myLeft, myRight = myRight, myLeft
	}
	myLeft -= 1
	myRight += 1

	totalPenalty := 0

	// РџСЂРѕС…РѕРґРёРј РїРѕ РІСЃРµРј РІРµСЂС€РёРЅР°Рј СЃР»РѕСЏ СЂРѕРґРёС‚РµР»СЏ
	for _, otherParent := range cm.Layers[parentLayer] {
		if otherParent == parentNode {
			continue // РїСЂРѕРїСѓСЃРєР°РµРј СЃР°РјСѓ РІРµСЂС€РёРЅСѓ
		}
		if otherParent.IsPseudo {
			continue // РїСЂРѕРїСѓСЃРєР°РµРј РїСЃРµРІРґРѕРІРµСЂС€РёРЅС‹
		}

		// Р”Р»СЏ РєР°Р¶РґРѕРіРѕ СЂРµР±С‘РЅРєР° РґСЂСѓРіРѕРіРѕ СЂРѕРґРёС‚РµР»СЏ
		for _, otherChild := range otherParent.Children {
			if otherChild == nil {
				continue
			}

			// РќР°С…РѕРґРёРј РєРѕРѕСЂРґРёРЅР°С‚С‹ СЃРІСЏР·Рё
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

			// РћС‚СЂРµР·РѕРє РґСЂСѓРіРѕР№ СЃРІСЏР·Рё
			otherLeft := otherParentCoord
			otherRight := otherChildCoord
			if otherLeft > otherRight {
				otherLeft, otherRight = otherRight, otherLeft
			}

			// Р’С‹С‡РёСЃР»СЏРµРј РґР»РёРЅСѓ РїРµСЂРµСЃРµС‡РµРЅРёСЏ РґРІСѓС… РѕС‚СЂРµР·РєРѕРІ
			overlapLen := segmentOverlap(myLeft, myRight, otherLeft, otherRight)
			totalPenalty += overlapLen
		}
	}

	return totalPenalty
}

// segmentOverlap РІС‹С‡РёСЃР»СЏРµС‚ РґР»РёРЅСѓ РїРµСЂРµСЃРµС‡РµРЅРёСЏ РґРІСѓС… РѕС‚СЂРµР·РєРѕРІ [a1, a2] Рё [b1, b2]
func segmentOverlap(a1, a2, b1, b2 int) int {
	// РЈР±РµРґРёРјСЃСЏ, С‡С‚Рѕ a1 <= a2 Рё b1 <= b2
	if a1 > a2 {
		a1, a2 = a2, a1
	}
	if b1 > b2 {
		b1, b2 = b2, b1
	}

	// РќР°С…РѕРґРёРј РїРµСЂРµСЃРµС‡РµРЅРёРµ
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

// getParentCoordinateForPerson РІРѕР·РІСЂР°С‰Р°РµС‚ РєРѕРѕСЂРґРёРЅР°С‚Сѓ РєРѕРЅРєСЂРµС‚РЅРѕРіРѕ С‡РµР»РѕРІРµРєР° РІ СЂРѕРґРёС‚РµР»СЊСЃРєРѕР№ РІРµСЂС€РёРЅРµ
func getParentCoordinateForPerson(parent *CoordNode, personIndex int) int {
	// Р”Р»СЏ СЂРѕРґРёС‚РµР»СЏ РІСЃРµРіРґР° РёСЃРїРѕР»СЊР·СѓРµРј С†РµРЅС‚СЂ РІРµСЂС€РёРЅС‹
	return (parent.Left + parent.Right) / 2
}

// getParentCoordinate РІРѕР·РІСЂР°С‰Р°РµС‚ РєРѕРѕСЂРґРёРЅР°С‚Сѓ СЂРѕРґРёС‚РµР»СЏ (legacy - С†РµРЅС‚СЂ РІРµСЂС€РёРЅС‹)
func getParentCoordinate(parent *CoordNode) int {
	// Р•СЃР»Рё СЂРѕРґРёС‚РµР»СЊ вЂ” СЃРєР»РµРµРЅРЅР°СЏ РІРµСЂС€РёРЅР° (2 С‡РµР»РѕРІРµРєР°), С‚Рѕ СЃСЂРµРґРЅРµРµ
	if len(parent.People) == 2 {
		return (parent.Left + parent.Right) / 2
	}
	// Р•СЃР»Рё РѕРґРёРЅРѕС‡РЅР°СЏ РІРµСЂС€РёРЅР°
	if parent.AddedLeft {
		return parent.Right // РґРѕР±Р°РІР»СЏР»Рё СЃР»РµРІР° вЂ” РєРѕРѕСЂРґРёРЅР°С‚Р° СЌС‚Рѕ РїСЂР°РІР°СЏ РіСЂР°РЅРёС†Р°
	}
	return parent.Left // РґРѕР±Р°РІР»СЏР»Рё СЃРїСЂР°РІР° вЂ” РєРѕРѕСЂРґРёРЅР°С‚Р° СЌС‚Рѕ Р»РµРІР°СЏ РіСЂР°РЅРёС†Р°
}

// getChildCoordinate РІРѕР·РІСЂР°С‰Р°РµС‚ РєРѕРѕСЂРґРёРЅР°С‚Сѓ СЂРµР±С‘РЅРєР°
func getChildCoordinate(child *CoordNode, personIndex int) int {
	// Р•СЃР»Рё РѕРґРёРЅРѕС‡РЅР°СЏ РІРµСЂС€РёРЅР° вЂ” СЃСЂРµРґРЅРµРµ
	if len(child.People) == 1 {
		return (child.Left + child.Right) / 2
	}
	// Р•СЃР»Рё СЃРєР»РµРµРЅРЅР°СЏ РІРµСЂС€РёРЅР°
	if personIndex == 0 {
		return child.Left + 1
	}
	return child.Right - 1
}

// collectPseudoChain СЃРѕР±РёСЂР°РµС‚ С†РµРїРѕС‡РєСѓ РїСЃРµРІРґРѕРІРµСЂС€РёРЅ + РЅРёР¶РЅСЋСЋ РІРµСЂС€РёРЅСѓ
func collectPseudoChain(start *CoordNode) []*CoordNode {
	chain := []*CoordNode{}
	visited := make(map[*CoordNode]bool)

	// РРґС‘Рј РІРЅРёР·
	current := start
	for current != nil && !visited[current] {
		visited[current] = true
		chain = append([]*CoordNode{current}, chain...) // РґРѕР±Р°РІР»СЏРµРј РІ РЅР°С‡Р°Р»Рѕ
		if len(current.Down) > 0 && current.Down[0] != nil {
			next := current.Down[0]
			if next.IsPseudo || current.IsPseudo {
				current = next
			} else {
				// Р”РѕС€Р»Рё РґРѕ РЅРµ-РїСЃРµРІРґРѕ РІРµСЂС€РёРЅС‹ СЃРЅРёР·Сѓ вЂ” РґРѕР±Р°РІР»СЏРµРј РµС‘
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

	// РРґС‘Рј РІРІРµСЂС… РѕС‚ start
	current = start
	for len(current.Up) > 0 && current.Up[0] != nil && !visited[current.Up[0]] {
		current = current.Up[0]
		visited[current] = true
		chain = append(chain, current)
	}

	return chain
}

// placeChainAtCoord СЂР°Р·РјРµС‰Р°РµС‚ РІСЃСЋ С†РµРїРѕС‡РєСѓ РЅР° Р·Р°РґР°РЅРЅРѕР№ РєРѕРѕСЂРґРёРЅР°С‚Рµ
func placeChainAtCoord(cn *CoordNode, coord int, cm *CoordMatrix, chainMap map[*CoordNode]*int, layerPaused map[int]bool) {
	chain := collectPseudoChain(cn)

	for _, node := range chain {
		placeNodeAtCoord(node, coord, cm)
		layerPaused[node.Layer] = false
		delete(chainMap, node)
	}
}

// placeNodeAtCoord СЂР°Р·РјРµС‰Р°РµС‚ РѕРґРЅСѓ РІРµСЂС€РёРЅСѓ РЅР° Р·Р°РґР°РЅРЅРѕР№ РєРѕРѕСЂРґРёРЅР°С‚Рµ
func placeNodeAtCoord(cn *CoordNode, coord int, cm *CoordMatrix) {
	if cn.Left != 0 || cn.Right != 0 {
		// РЈР¶Рµ СЂР°Р·РјРµС‰РµРЅР°
		return
	}

	if cn.IsPseudo {
		// РџСЃРµРІРґРѕРІРµСЂС€РёРЅР° вЂ” РєРѕРѕСЂРґРёРЅР°С‚С‹ СЃРѕРІРїР°РґР°СЋС‚ (С€РёСЂРёРЅР° 0)
		cn.Left = coord
		cn.Right = coord
	} else {
		// РћР±С‹С‡РЅР°СЏ РІРµСЂС€РёРЅР° вЂ” С€РёСЂРёРЅР° 2
		cn.Left = coord
		cn.Right = coord + 2
	}

	cm.AddNode(cn)
}

// placeChain СЂР°Р·РјРµС‰Р°РµС‚ РІСЃСЋ С†РµРїРѕС‡РєСѓ РЅР° Р·Р°РґР°РЅРЅРѕР№ РєРѕРѕСЂРґРёРЅР°С‚Рµ (legacy)
func placeChain(cn *CoordNode, coord int, cm *CoordMatrix, chainMap map[*CoordNode]*int, layerPaused map[int]bool) {
	placeChainAtCoord(cn, coord, cm, chainMap, layerPaused)
}

// placeNode СЂР°Р·РјРµС‰Р°РµС‚ РѕРґРЅСѓ РІРµСЂС€РёРЅСѓ РЅР° Р·Р°РґР°РЅРЅРѕР№ РєРѕРѕСЂРґРёРЅР°С‚Рµ (legacy)
func placeNode(cn *CoordNode, coord int, cm *CoordMatrix) {
	placeNodeAtCoord(cn, coord, cm)
}
