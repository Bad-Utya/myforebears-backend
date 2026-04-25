package stage3_ordering

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

// isLeftOf РїСЂРѕРІРµСЂСЏРµС‚, РЅР°С…РѕРґРёС‚СЃСЏ Р»Рё node Р»РµРІРµРµ target РЅР° СЃР»РѕРµ
// РџСЂРѕС…РѕРґРёРј РѕС‚ node РІРїСЂР°РІРѕ Рё РёС‰РµРј target
func (om *OrderManager) isLeftOf(node, target *LayerNode) bool {
	if node == target {
		return true // СЃР°Рј СѓР·РµР» СЃС‡РёС‚Р°РµС‚СЃСЏ "Р»РµРІРµРµ РёР»Рё СЂР°РІРµРЅ"
	}
	for curr := node.Next; curr != nil && !curr.IsTail(); curr = curr.Next {
		if curr == target {
			return true
		}
	}
	return false
}

// isRightOf РїСЂРѕРІРµСЂСЏРµС‚, РЅР°С…РѕРґРёС‚СЃСЏ Р»Рё node РїСЂР°РІРµРµ target РЅР° СЃР»РѕРµ
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

// followPointerToLayer СЃР»РµРґСѓРµС‚ РїРѕ СѓРєР°Р·Р°С‚РµР»СЏРј РѕС‚ node РґРѕ targetLayer
// Р’РѕР·РІСЂР°С‰Р°РµС‚ СѓР·РµР» РЅР° targetLayer РёР»Рё nil РµСЃР»Рё РїСѓС‚СЊ РЅРµ РЅР°Р№РґРµРЅ
func (om *OrderManager) followPointerToLayer(node *LayerNode, targetLayer int, goingUp bool) *LayerNode {
	current := node
	for current != nil && current.Layer != targetLayer {
		if goingUp {
			// РСЃРїРѕР»СЊР·СѓРµРј GetRightUp (РїРѕСЃР»РµРґРЅРёР№ СЌР»РµРјРµРЅС‚ Up)
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

// findRightmostReaching РЅР°С…РѕРґРёС‚ СЃР°РјСѓСЋ РїСЂР°РІСѓСЋ РІРµСЂС€РёРЅСѓ РЅР° layer,
// РєРѕС‚РѕСЂР°СЏ РїСЂРё РїРµСЂРµС…РѕРґРµ РЅР° targetLayer РѕРєР°Р·С‹РІР°РµС‚СЃСЏ Р»РµРІРµРµ РёР»Рё СЂР°РІРЅР° targetNode
// goingUp = true РѕР·РЅР°С‡Р°РµС‚ С‡С‚Рѕ targetLayer > layer (РёРґС‘Рј РІРІРµСЂС… РїРѕ СЃР»РѕСЏРј)
func (om *OrderManager) findRightmostReaching(layer, targetLayer int, targetNode *LayerNode, goingUp bool) *LayerNode {
	layerObj := om.Layers[layer]
	if layerObj == nil {
		return nil
	}

	var result *LayerNode

	// РџСЂРѕС…РѕРґРёРј РїРѕ РІСЃРµРј СѓР·Р»Р°Рј СЃР»РѕСЏ СЃР»РµРІР° РЅР°РїСЂР°РІРѕ
	for node := layerObj.Head.Next; node != nil && !node.IsTail(); node = node.Next {
		// РЎР»РµРґСѓРµРј РїРѕ СѓРєР°Р·Р°С‚РµР»СЏРј РґРѕ С†РµР»РµРІРѕРіРѕ СЃР»РѕСЏ
		reachable := om.followPointerToLayer(node, targetLayer, goingUp)

		// Р•СЃР»Рё РЅРµ РґРѕСЃС‚РёРіР»Рё С†РµР»РµРІРѕРіРѕ СЃР»РѕСЏ вЂ” РІРµСЂС€РёРЅР° РЅРµ СЃРІСЏР·Р°РЅР°, РїСЂРѕРїСѓСЃРєР°РµРј
		if reachable == nil {
			continue
		}

		// РџСЂРѕРІРµСЂСЏРµРј: reachable Р»РµРІРµРµ РёР»Рё СЂР°РІРµРЅ targetNode?
		if om.isLeftOf(reachable, targetNode) {
			result = node // РѕР±РЅРѕРІР»СЏРµРј СЂРµР·СѓР»СЊС‚Р°С‚ (РёС‰РµРј СЃР°РјС‹Р№ РїСЂР°РІС‹Р№)
		}
	}

	return result
}

// findLeftmostReaching РЅР°С…РѕРґРёС‚ СЃР°РјСѓСЋ Р»РµРІСѓСЋ РІРµСЂС€РёРЅСѓ РЅР° layer,
// РєРѕС‚РѕСЂР°СЏ РїСЂРё РїРµСЂРµС…РѕРґРµ РЅР° targetLayer РѕРєР°Р·С‹РІР°РµС‚СЃСЏ РїСЂР°РІРµРµ РёР»Рё СЂР°РІРЅР° targetNode
func (om *OrderManager) findLeftmostReaching(layer, targetLayer int, targetNode *LayerNode, goingUp bool) *LayerNode {
	layerObj := om.Layers[layer]
	if layerObj == nil {
		return nil
	}

	// РџСЂРѕС…РѕРґРёРј РїРѕ РІСЃРµРј СѓР·Р»Р°Рј СЃР»РѕСЏ СЃР»РµРІР° РЅР°РїСЂР°РІРѕ
	for node := layerObj.Head.Next; node != nil && !node.IsTail(); node = node.Next {
		// РЎР»РµРґСѓРµРј РїРѕ СѓРєР°Р·Р°С‚РµР»СЏРј РґРѕ С†РµР»РµРІРѕРіРѕ СЃР»РѕСЏ
		reachable := om.followPointerToLayer(node, targetLayer, goingUp)

		// Р•СЃР»Рё РЅРµ РґРѕСЃС‚РёРіР»Рё С†РµР»РµРІРѕРіРѕ СЃР»РѕСЏ вЂ” РїСЂРѕРїСѓСЃРєР°РµРј
		if reachable == nil {
			continue
		}

		// РџРµСЂРІР°СЏ РІРµСЂС€РёРЅР°, РєРѕС‚РѕСЂР°СЏ РґРѕСЃС‚РёРіР°РµС‚ targetNode РёР»Рё РїСЂР°РІРµРµ
		if om.isRightOf(reachable, targetNode) {
			return node
		}
	}

	return nil
}

// AddPersonRight РґРѕР±Р°РІР»СЏРµС‚ С‡РµР»РѕРІРµРєР° СЃРїСЂР°РІР° РѕС‚ fromNode
// fromLayer вЂ” СЃР»РѕР№ fromNode, targetLayer вЂ” СЃР»РѕР№, РЅР° РєРѕС‚РѕСЂС‹Р№ РґРѕР±Р°РІР»СЏРµРј
// fromPersonIndex вЂ” РїРѕР·РёС†РёСЏ С‡РµР»РѕРІРµРєР° РІ fromNode.People, РєРѕС‚РѕСЂС‹Р№ РґРѕР±Р°РІР»СЏРµС‚
func (om *OrderManager) AddPersonRight(fromNode *LayerNode, addedPerson *stage1_input.Person, fromLayer, targetLayer int, fromPersonIndex int) *LayerNode {
	om.ensureLayer(targetLayer)

	if targetLayer == fromLayer {
		// РџР°СЂС‚РЅС‘СЂ РЅР° С‚РѕРј Р¶Рµ СЃР»РѕРµ
		return om.addPartnerRight(fromNode, addedPerson, fromLayer)
	}

	// РћРїСЂРµРґРµР»СЏРµРј РЅР°РїСЂР°РІР»РµРЅРёРµ
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

	// РС‚РµСЂР°С‚РёРІРЅРѕ РїСЂРѕС…РѕРґРёРј РїРѕ СЃР»РѕСЏРј
	for layer := startLayer; ; layer += step {
		om.ensureLayer(layer)

		// РЎРѕР·РґР°С‘Рј СѓР·РµР» (РїСЃРµРІРґРѕ РёР»Рё СЂРµР°Р»СЊРЅС‹Р№ РЅР° РїРѕСЃР»РµРґРЅРµРј СЃР»РѕРµ)
		var newNode *LayerNode
		if layer == endLayer {
			newNode = om.CreatePersonNode(addedPerson, layer)
			newNode.AddedLeft = false // РґРѕР±Р°РІР»СЏРµРј СЃРїСЂР°РІР°
		} else {
			newNode = om.CreatePseudoNode(layer)
		}

		// РќР°С…РѕРґРёРј РїРѕР·РёС†РёСЋ РґР»СЏ РІСЃС‚Р°РІРєРё
		insertAfter := om.findRightmostReaching(layer, layer-step, prevNode, !goingUp)

		if insertAfter == nil {
			// РЎР»РѕР№ РїСѓСЃС‚ РёР»Рё РЅРµС‚ РїРѕРґС…РѕРґСЏС‰РёС… РІРµСЂС€РёРЅ вЂ” РІСЃС‚Р°РІР»СЏРµРј РїРѕСЃР»Рµ Head
			layerObj := om.Layers[layer]
			insertAfter = layerObj.Head
		}

		// Р’СЃС‚Р°РІР»СЏРµРј СѓР·РµР»
		om.insertAfter(insertAfter, newNode)

		// РЈСЃС‚Р°РЅР°РІР»РёРІР°РµРј РІРµСЂС‚РёРєР°Р»СЊРЅС‹Рµ СЃРІСЏР·Рё
		if goingUp {
			// РќРѕРІС‹Р№ СѓР·РµР» РІС‹С€Рµ prevNode
			newNode.LeftDown = prevNode
			newNode.RightDown = prevNode
			// Р”РѕР±Р°РІР»СЏРµРј СѓРєР°Р·Р°С‚РµР»СЊ РІРІРµСЂС… РІ prevNode РїРѕ РёРЅРґРµРєСЃСѓ
			if prevNode == fromNode && len(prevNode.People) > 0 {
				// Р”Р»СЏ РІРµСЂС€РёРЅС‹ СЃ Р»СЋРґСЊРјРё РёСЃРїРѕР»СЊР·СѓРµРј РёРЅРґРµРєСЃ С‡РµР»РѕРІРµРєР°
				for len(prevNode.Up) <= fromPersonIndex {
					prevNode.Up = append(prevNode.Up, nil)
				}
				prevNode.Up[fromPersonIndex] = newNode
			} else {
				// Р”Р»СЏ РїСЃРµРІРґРѕРІРµСЂС€РёРЅ РїСЂРѕСЃС‚Рѕ РґРѕР±Р°РІР»СЏРµРј
				prevNode.Up = append(prevNode.Up, newNode)
			}
		} else {
			// РќРѕРІС‹Р№ СѓР·РµР» РЅРёР¶Рµ prevNode (СЂРµР±С‘РЅРѕРє)
			// РћРґРёРЅ СѓРєР°Р·Р°С‚РµР»СЊ РІРІРµСЂС… РЅР° СЂРѕРґРёС‚РµР»СЊСЃРєСѓСЋ РІРµСЂС€РёРЅСѓ
			newNode.Up = []*LayerNode{prevNode}
			// РћР±РЅРѕРІР»СЏРµРј СѓРєР°Р·Р°С‚РµР»Рё prevNode РІРЅРёР·
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

// AddPersonLeft РґРѕР±Р°РІР»СЏРµС‚ С‡РµР»РѕРІРµРєР° СЃР»РµРІР° РѕС‚ fromNode
// fromPersonIndex вЂ” РїРѕР·РёС†РёСЏ С‡РµР»РѕРІРµРєР° РІ fromNode.People, РєРѕС‚РѕСЂС‹Р№ РґРѕР±Р°РІР»СЏРµС‚
func (om *OrderManager) AddPersonLeft(fromNode *LayerNode, addedPerson *stage1_input.Person, fromLayer, targetLayer int, fromPersonIndex int) *LayerNode {
	om.ensureLayer(targetLayer)

	if targetLayer == fromLayer {
		// РџР°СЂС‚РЅС‘СЂ РЅР° С‚РѕРј Р¶Рµ СЃР»РѕРµ
		return om.addPartnerLeft(fromNode, addedPerson, fromLayer)
	}

	// РћРїСЂРµРґРµР»СЏРµРј РЅР°РїСЂР°РІР»РµРЅРёРµ
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

	// РС‚РµСЂР°С‚РёРІРЅРѕ РїСЂРѕС…РѕРґРёРј РїРѕ СЃР»РѕСЏРј
	for layer := startLayer; ; layer += step {
		om.ensureLayer(layer)

		// РЎРѕР·РґР°С‘Рј СѓР·РµР» (РїСЃРµРІРґРѕ РёР»Рё СЂРµР°Р»СЊРЅС‹Р№ РЅР° РїРѕСЃР»РµРґРЅРµРј СЃР»РѕРµ)
		var newNode *LayerNode
		if layer == endLayer {
			newNode = om.CreatePersonNode(addedPerson, layer)
			newNode.AddedLeft = true // РґРѕР±Р°РІР»СЏРµРј СЃР»РµРІР°
		} else {
			newNode = om.CreatePseudoNode(layer)
		}

		// РќР°С…РѕРґРёРј РїРѕР·РёС†РёСЋ РґР»СЏ РІСЃС‚Р°РІРєРё
		insertBefore := om.findLeftmostReaching(layer, layer-step, prevNode, !goingUp)

		if insertBefore == nil {
			// РЎР»РѕР№ РїСѓСЃС‚ РёР»Рё РЅРµС‚ РїРѕРґС…РѕРґСЏС‰РёС… РІРµСЂС€РёРЅ вЂ” РІСЃС‚Р°РІР»СЏРµРј РїРµСЂРµРґ Tail
			layerObj := om.Layers[layer]
			insertBefore = layerObj.Tail
		}

		// Р’СЃС‚Р°РІР»СЏРµРј СѓР·РµР»
		om.insertBefore(insertBefore, newNode)

		// РЈСЃС‚Р°РЅР°РІР»РёРІР°РµРј РІРµСЂС‚РёРєР°Р»СЊРЅС‹Рµ СЃРІСЏР·Рё
		if goingUp {
			newNode.LeftDown = prevNode
			newNode.RightDown = prevNode
			// Р”РѕР±Р°РІР»СЏРµРј СѓРєР°Р·Р°С‚РµР»СЊ РІРІРµСЂС… РІ prevNode РїРѕ РёРЅРґРµРєСЃСѓ
			if prevNode == fromNode && len(prevNode.People) > 0 {
				// Р”Р»СЏ РІРµСЂС€РёРЅС‹ СЃ Р»СЋРґСЊРјРё РёСЃРїРѕР»СЊР·СѓРµРј РёРЅРґРµРєСЃ С‡РµР»РѕРІРµРєР°
				for len(prevNode.Up) <= fromPersonIndex {
					prevNode.Up = append(prevNode.Up, nil)
				}
				prevNode.Up[fromPersonIndex] = newNode
			} else {
				// Р”Р»СЏ РїСЃРµРІРґРѕРІРµСЂС€РёРЅ РїСЂРѕСЃС‚Рѕ РґРѕР±Р°РІР»СЏРµРј
				prevNode.Up = append(prevNode.Up, newNode)
			}
		} else {
			// РќРѕРІС‹Р№ СѓР·РµР» РЅРёР¶Рµ prevNode (СЂРµР±С‘РЅРѕРє)
			// РћРґРёРЅ СѓРєР°Р·Р°С‚РµР»СЊ РІРІРµСЂС… РЅР° СЂРѕРґРёС‚РµР»СЊСЃРєСѓСЋ РІРµСЂС€РёРЅСѓ
			newNode.Up = []*LayerNode{prevNode}
			// РћР±РЅРѕРІР»СЏРµРј СѓРєР°Р·Р°С‚РµР»Рё prevNode РІРЅРёР·
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

// AddParentPairRight РґРѕР±Р°РІР»СЏРµС‚ РїР°СЂСѓ СЂРѕРґРёС‚РµР»РµР№ СЃРїСЂР°РІР° РѕС‚ fromNode
// siblingUp - СѓРєР°Р·Р°С‚РµР»СЊ Up РґСЂСѓРіРѕРіРѕ С‡РµР»РѕРІРµРєР° РІ РІРµСЂС€РёРЅРµ (РјРѕР¶РµС‚ Р±С‹С‚СЊ nil)
// fromPersonIndex - РїРѕР·РёС†РёСЏ С‡РµР»РѕРІРµРєР°, РґРѕР±Р°РІР»СЏСЋС‰РµРіРѕ СЂРѕРґРёС‚РµР»РµР№, РІ СЃРїРёСЃРєРµ People РІРµСЂС€РёРЅС‹
func (om *OrderManager) AddParentPairRight(fromNode *LayerNode, parent1, parent2 *stage1_input.Person, fromLayer, targetLayer int, siblingUp *LayerNode, fromPersonIndex int) *LayerNode {
	om.ensureLayer(targetLayer)

	var prevNode *LayerNode = fromNode
	step := 1
	startLayer := fromLayer + 1
	endLayer := targetLayer
	isFirstStep := true

	// РС‚РµСЂР°С‚РёРІРЅРѕ РїСЂРѕС…РѕРґРёРј РїРѕ СЃР»РѕСЏРј
	for layer := startLayer; ; layer += step {
		om.ensureLayer(layer)

		// РЎРѕР·РґР°С‘Рј СѓР·РµР» (РїСЃРµРІРґРѕ РёР»Рё РїР°СЂСѓ РЅР° РїРѕСЃР»РµРґРЅРµРј СЃР»РѕРµ)
		var newNode *LayerNode
		if layer == endLayer {
			// РЎРѕР·РґР°С‘Рј РІРµСЂС€РёРЅСѓ СЃ РїР°СЂРѕР№ СЂРѕРґРёС‚РµР»РµР№
			newNode = om.CreatePairNode(parent1, parent2, layer)
		} else {
			newNode = om.CreatePseudoNode(layer)
		}

		// РќР°С…РѕРґРёРј РїРѕР·РёС†РёСЋ РґР»СЏ РІСЃС‚Р°РІРєРё
		var insertAfter *LayerNode

		// РќР° РїРµСЂРІРѕР№ РёС‚РµСЂР°С†РёРё, РµСЃР»Рё РµСЃС‚СЊ siblingUp РЅР° СЌС‚РѕРј СЃР»РѕРµ, СЂР°Р·РјРµС‰Р°РµРј СЂСЏРґРѕРј СЃ РЅРёРј
		if isFirstStep && siblingUp != nil && siblingUp.Layer == layer {
			// fromPersonIndex РѕРїСЂРµРґРµР»СЏРµС‚ РїРѕСЂСЏРґРѕРє: РµСЃР»Рё fromPerson СЃРїСЂР°РІР° (index=1), РµРіРѕ СЂРѕРґРёС‚РµР»Рё СЃРїСЂР°РІР° РѕС‚ siblingUp
			if fromPersonIndex == 1 {
				// Р§РµР»РѕРІРµРє СЃРїСЂР°РІР° РІ РІРµСЂС€РёРЅРµ в†’ РµРіРѕ СЂРѕРґРёС‚РµР»Рё СЃРїСЂР°РІР° РѕС‚ СЂРѕРґРёС‚РµР»РµР№ СЃРѕСЃРµРґР°
				insertAfter = siblingUp
			} else {
				// Р§РµР»РѕРІРµРє СЃР»РµРІР° РІ РІРµСЂС€РёРЅРµ в†’ РµРіРѕ СЂРѕРґРёС‚РµР»Рё СЃР»РµРІР° РѕС‚ СЂРѕРґРёС‚РµР»РµР№ СЃРѕСЃРµРґР°
				// РќРѕ СЌС‚Рѕ AddParentPairRight, С‚Р°Рє С‡С‚Рѕ РІСЃС‚Р°РІР»СЏРµРј РїРµСЂРµРґ siblingUp
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

		// Р’СЃС‚Р°РІР»СЏРµРј СѓР·РµР»
		om.insertAfter(insertAfter, newNode)

		// РЈСЃС‚Р°РЅР°РІР»РёРІР°РµРј РІРµСЂС‚РёРєР°Р»СЊРЅС‹Рµ СЃРІСЏР·Рё
		newNode.LeftDown = prevNode
		newNode.RightDown = prevNode
		// Р”РѕР±Р°РІР»СЏРµРј СѓРєР°Р·Р°С‚РµР»СЊ РІРІРµСЂС… РІ prevNode
		if isFirstStep && len(prevNode.People) > 0 {
			// Р”Р»СЏ РІРµСЂС€РёРЅС‹ СЃ Р»СЋРґСЊРјРё РёСЃРїРѕР»СЊР·СѓРµРј РёРЅРґРµРєСЃ С‡РµР»РѕРІРµРєР°
			// РЈР±РµР¶РґР°РµРјСЃСЏ, С‡С‚Рѕ Up РёРјРµРµС‚ РЅСѓР¶РЅСѓСЋ РґР»РёРЅСѓ
			for len(prevNode.Up) <= fromPersonIndex {
				prevNode.Up = append(prevNode.Up, nil)
			}
			prevNode.Up[fromPersonIndex] = newNode
		} else {
			// Р”Р»СЏ РїСЃРµРІРґРѕРІРµСЂС€РёРЅ РїСЂРѕСЃС‚Рѕ РґРѕР±Р°РІР»СЏРµРј
			prevNode.Up = append(prevNode.Up, newNode)
		}

		isFirstStep = false
		prevNode = newNode

		if layer == endLayer {
			return newNode
		}
	}
}

// AddParentPairLeft РґРѕР±Р°РІР»СЏРµС‚ РїР°СЂСѓ СЂРѕРґРёС‚РµР»РµР№ СЃР»РµРІР° РѕС‚ fromNode
// siblingUp - СѓРєР°Р·Р°С‚РµР»СЊ Up РґСЂСѓРіРѕРіРѕ С‡РµР»РѕРІРµРєР° РІ РІРµСЂС€РёРЅРµ (РјРѕР¶РµС‚ Р±С‹С‚СЊ nil)
// fromPersonIndex - РїРѕР·РёС†РёСЏ С‡РµР»РѕРІРµРєР°, РґРѕР±Р°РІР»СЏСЋС‰РµРіРѕ СЂРѕРґРёС‚РµР»РµР№, РІ СЃРїРёСЃРєРµ People РІРµСЂС€РёРЅС‹
func (om *OrderManager) AddParentPairLeft(fromNode *LayerNode, parent1, parent2 *stage1_input.Person, fromLayer, targetLayer int, siblingUp *LayerNode, fromPersonIndex int) *LayerNode {
	om.ensureLayer(targetLayer)

	var prevNode *LayerNode = fromNode
	step := 1
	startLayer := fromLayer + 1
	endLayer := targetLayer
	isFirstStep := true

	// РС‚РµСЂР°С‚РёРІРЅРѕ РїСЂРѕС…РѕРґРёРј РїРѕ СЃР»РѕСЏРј
	for layer := startLayer; ; layer += step {
		om.ensureLayer(layer)

		// РЎРѕР·РґР°С‘Рј СѓР·РµР» (РїСЃРµРІРґРѕ РёР»Рё РїР°СЂСѓ РЅР° РїРѕСЃР»РµРґРЅРµРј СЃР»РѕРµ)
		var newNode *LayerNode
		if layer == endLayer {
			// РЎРѕР·РґР°С‘Рј РІРµСЂС€РёРЅСѓ СЃ РїР°СЂРѕР№ СЂРѕРґРёС‚РµР»РµР№
			newNode = om.CreatePairNode(parent1, parent2, layer)
		} else {
			newNode = om.CreatePseudoNode(layer)
		}

		// РќР°С…РѕРґРёРј РїРѕР·РёС†РёСЋ РґР»СЏ РІСЃС‚Р°РІРєРё
		var insertBefore *LayerNode

		// РќР° РїРµСЂРІРѕР№ РёС‚РµСЂР°С†РёРё, РµСЃР»Рё РµСЃС‚СЊ siblingUp РЅР° СЌС‚РѕРј СЃР»РѕРµ, СЂР°Р·РјРµС‰Р°РµРј СЂСЏРґРѕРј СЃ РЅРёРј
		if isFirstStep && siblingUp != nil && siblingUp.Layer == layer {
			// fromPersonIndex РѕРїСЂРµРґРµР»СЏРµС‚ РїРѕСЂСЏРґРѕРє: РµСЃР»Рё fromPerson СЃР»РµРІР° (index=0), РµРіРѕ СЂРѕРґРёС‚РµР»Рё СЃР»РµРІР° РѕС‚ siblingUp
			if fromPersonIndex == 0 {
				// Р§РµР»РѕРІРµРє СЃР»РµРІР° РІ РІРµСЂС€РёРЅРµ в†’ РµРіРѕ СЂРѕРґРёС‚РµР»Рё СЃР»РµРІР° РѕС‚ СЂРѕРґРёС‚РµР»РµР№ СЃРѕСЃРµРґР°
				insertBefore = siblingUp
			} else {
				// Р§РµР»РѕРІРµРє СЃРїСЂР°РІР° РІ РІРµСЂС€РёРЅРµ в†’ РµРіРѕ СЂРѕРґРёС‚РµР»Рё СЃРїСЂР°РІР° РѕС‚ СЂРѕРґРёС‚РµР»РµР№ СЃРѕСЃРµРґР°
				// РќРѕ СЌС‚Рѕ AddParentPairLeft, С‚Р°Рє С‡С‚Рѕ РІСЃС‚Р°РІР»СЏРµРј РїРѕСЃР»Рµ siblingUp
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

		// Р’СЃС‚Р°РІР»СЏРµРј СѓР·РµР»
		om.insertBefore(insertBefore, newNode)

		// РЈСЃС‚Р°РЅР°РІР»РёРІР°РµРј РІРµСЂС‚РёРєР°Р»СЊРЅС‹Рµ СЃРІСЏР·Рё
		newNode.LeftDown = prevNode
		newNode.RightDown = prevNode
		// Р”РѕР±Р°РІР»СЏРµРј СѓРєР°Р·Р°С‚РµР»СЊ РІРІРµСЂС… РІ prevNode
		if isFirstStep && len(prevNode.People) > 0 {
			// Р”Р»СЏ РІРµСЂС€РёРЅС‹ СЃ Р»СЋРґСЊРјРё РёСЃРїРѕР»СЊР·СѓРµРј РёРЅРґРµРєСЃ С‡РµР»РѕРІРµРєР°
			// РЈР±РµР¶РґР°РµРјСЃСЏ, С‡С‚Рѕ Up РёРјРµРµС‚ РЅСѓР¶РЅСѓСЋ РґР»РёРЅСѓ
			for len(prevNode.Up) <= fromPersonIndex {
				prevNode.Up = append(prevNode.Up, nil)
			}
			prevNode.Up[fromPersonIndex] = newNode
		} else {
			// Р”Р»СЏ РїСЃРµРІРґРѕРІРµСЂС€РёРЅ РїСЂРѕСЃС‚Рѕ РґРѕР±Р°РІР»СЏРµРј
			prevNode.Up = append(prevNode.Up, newNode)
		}

		isFirstStep = false
		prevNode = newNode

		if layer == endLayer {
			return newNode
		}
	}
}

// collectPartnerCluster СЃРѕР±РёСЂР°РµС‚ РІСЃРµ СЃРІСЏР·Р°РЅРЅС‹Рµ РїР°СЂС‚РЅС‘СЂР°РјРё РІРµСЂС€РёРЅС‹
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

		// РџР°СЂС‚РЅС‘СЂС‹ вЂ” Р»СЋРґРё РёР· СЌС‚РѕР№ РІРµСЂС€РёРЅС‹
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

// addPartnerRight РґРѕР±Р°РІР»СЏРµС‚ РїР°СЂС‚РЅС‘СЂР° СЃРїСЂР°РІР°
func (om *OrderManager) addPartnerRight(fromNode *LayerNode, addedPerson *stage1_input.Person, layer int) *LayerNode {
	// РџСЂРѕРІРµСЂСЏРµРј: РµСЃР»Рё РІ РІРµСЂС€РёРЅРµ С‚РѕР»СЊРєРѕ РѕРґРёРЅ С‡РµР»РѕРІРµРє вЂ” РґРѕР±Р°РІР»СЏРµРј РІ С‚Сѓ Р¶Рµ РІРµСЂС€РёРЅСѓ
	if len(fromNode.People) == 1 {
		// РџРµСЂРІС‹Р№ РїР°СЂС‚РЅС‘СЂ вЂ” РґРѕР±Р°РІР»СЏРµРј РІ СЃСѓС‰РµСЃС‚РІСѓСЋС‰СѓСЋ РІРµСЂС€РёРЅСѓ СЃРїСЂР°РІР°
		om.AddPersonToExistingNode(fromNode, addedPerson, "right")
		return fromNode
	}

	// РРЅР°С‡Рµ вЂ” СЌС‚Рѕ РІС‚РѕСЂРѕР№+ РїР°СЂС‚РЅС‘СЂ, СЃРѕР·РґР°С‘Рј РЅРѕРІСѓСЋ РІРµСЂС€РёРЅСѓ
	// РЎРѕР±РёСЂР°РµРј РєР»Р°СЃС‚РµСЂ РїР°СЂС‚РЅС‘СЂРѕРІ
	cluster := om.collectPartnerCluster(fromNode)

	// РќР°С…РѕРґРёРј СЃР°РјРѕРіРѕ РїСЂР°РІРѕРіРѕ РІ РєР»Р°СЃС‚РµСЂРµ
	var rightmost *LayerNode = fromNode
	for _, node := range cluster {
		if om.isRightOf(node, rightmost) && node != rightmost {
			// node РїСЂР°РІРµРµ rightmost
			rightmost = node
		}
	}

	// Р‘РѕР»РµРµ С‚РѕС‡РЅС‹Р№ РїРѕРёСЃРє: РёРґС‘Рј РІРїСЂР°РІРѕ РѕС‚ rightmost Рё РїСЂРѕРІРµСЂСЏРµРј, РІ РєР»Р°СЃС‚РµСЂРµ Р»Рё
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

	// РЎРѕР·РґР°С‘Рј СѓР·РµР» Рё РІСЃС‚Р°РІР»СЏРµРј РїРѕСЃР»Рµ rightmost
	newNode := om.CreatePersonNode(addedPerson, layer)
	newNode.AddedLeft = false // РґРѕР±Р°РІР»СЏРµРј СЃРїСЂР°РІР°
	om.insertAfter(rightmost, newNode)

	return newNode
}

// addPartnerLeft РґРѕР±Р°РІР»СЏРµС‚ РїР°СЂС‚РЅС‘СЂР° СЃР»РµРІР°
func (om *OrderManager) addPartnerLeft(fromNode *LayerNode, addedPerson *stage1_input.Person, layer int) *LayerNode {
	// РџСЂРѕРІРµСЂСЏРµРј: РµСЃР»Рё РІ РІРµСЂС€РёРЅРµ С‚РѕР»СЊРєРѕ РѕРґРёРЅ С‡РµР»РѕРІРµРє вЂ” РґРѕР±Р°РІР»СЏРµРј РІ С‚Сѓ Р¶Рµ РІРµСЂС€РёРЅСѓ
	if len(fromNode.People) == 1 {
		// РџРµСЂРІС‹Р№ РїР°СЂС‚РЅС‘СЂ вЂ” РґРѕР±Р°РІР»СЏРµРј РІ СЃСѓС‰РµСЃС‚РІСѓСЋС‰СѓСЋ РІРµСЂС€РёРЅСѓ СЃР»РµРІР°
		om.AddPersonToExistingNode(fromNode, addedPerson, "left")
		return fromNode
	}

	// РРЅР°С‡Рµ вЂ” СЌС‚Рѕ РІС‚РѕСЂРѕР№+ РїР°СЂС‚РЅС‘СЂ, СЃРѕР·РґР°С‘Рј РЅРѕРІСѓСЋ РІРµСЂС€РёРЅСѓ
	// РЎРѕР±РёСЂР°РµРј РєР»Р°СЃС‚РµСЂ РїР°СЂС‚РЅС‘СЂРѕРІ
	cluster := om.collectPartnerCluster(fromNode)

	// РќР°С…РѕРґРёРј СЃР°РјРѕРіРѕ Р»РµРІРѕРіРѕ РІ РєР»Р°СЃС‚РµСЂРµ
	var leftmost *LayerNode = fromNode
	for _, node := range cluster {
		if om.isLeftOf(node, leftmost) && node != leftmost {
			leftmost = node
		}
	}

	// РРґС‘Рј РІР»РµРІРѕ Рё РїСЂРѕРІРµСЂСЏРµРј, РІ РєР»Р°СЃС‚РµСЂРµ Р»Рё
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

	// РЎРѕР·РґР°С‘Рј СѓР·РµР» Рё РІСЃС‚Р°РІР»СЏРµРј РїРµСЂРµРґ leftmost
	newNode := om.CreatePersonNode(addedPerson, layer)
	newNode.AddedLeft = true // РґРѕР±Р°РІР»СЏРµРј СЃР»РµРІР°
	om.insertBefore(leftmost, newNode)

	return newNode
}
