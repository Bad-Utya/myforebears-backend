package stage3_ordering

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

// ProcessPlacementHistory РѕР±СЂР°Р±Р°С‚С‹РІР°РµС‚ РёСЃС‚РѕСЂРёСЋ СЂР°Р·РјРµС‰РµРЅРёР№ Рё СЃС‚СЂРѕРёС‚ РїРѕСЂСЏРґРѕРє РІ СЃР»РѕСЏС…
func ProcessPlacementHistory(
	history *stage1_input.PlacementHistory,
	startPerson *stage1_input.Person,
	startLayer int,
	layouts map[int]*stage1_input.PersonLayout,
) *OrderManager {
	// РћРїСЂРµРґРµР»СЏРµРј РґРёР°РїР°Р·РѕРЅ СЃР»РѕС‘РІ
	minLayer, maxLayer := startLayer, startLayer
	for _, layout := range layouts {
		if layout.Layer < minLayer {
			minLayer = layout.Layer
		}
		if layout.Layer > maxLayer {
			maxLayer = layout.Layer
		}
	}

	// РЎРѕР·РґР°С‘Рј РјРµРЅРµРґР¶РµСЂ
	om := NewOrderManager(minLayer, maxLayer)

	// Р”РѕР±Р°РІР»СЏРµРј РЅР°С‡Р°Р»СЊРЅСѓСЋ РІРµСЂС€РёРЅСѓ
	om.AddStartPerson(startPerson, startLayer)

	// РћР±СЂР°Р±Р°С‚С‹РІР°РµРј РєР°Р¶РґСѓСЋ Р·Р°РїРёСЃСЊ РІ РёСЃС‚РѕСЂРёРё
	for _, record := range history.Records {
		fromPerson := record.FromPerson
		addedPerson := record.AddedPerson
		addedPerson2 := record.AddedPerson2
		direction := record.Direction
		relationType := record.RelationType

		// РџРѕР»СѓС‡Р°РµРј СЃР»РѕРё
		fromLayout := layouts[fromPerson.ID]
		addedLayout := layouts[addedPerson.ID]
		if fromLayout == nil || addedLayout == nil {
			continue
		}

		fromLayer := fromLayout.Layer
		targetLayer := addedLayout.Layer

		// РџРѕР»СѓС‡Р°РµРј СѓР·РµР» РґРѕР±Р°РІР»СЏСЋС‰РµРіРѕ
		fromNode := om.GetPersonNode(fromPerson.ID)
		if fromNode == nil {
			continue
		}

		// РџСЂРѕРІРµСЂСЏРµРј, СЌС‚Рѕ РґРѕР±Р°РІР»РµРЅРёРµ РїР°СЂС‹ СЂРѕРґРёС‚РµР»РµР№?
		if addedPerson2 != nil && relationType == stage1_input.RelationParent {
			// РС‰РµРј Up РґСЂСѓРіРѕРіРѕ С‡РµР»РѕРІРµРєР° РІ РІРµСЂС€РёРЅРµ (РµСЃР»Рё РµСЃС‚СЊ)
			var siblingUp *LayerNode
			var fromPersonIndex int = -1

			// РќР°С…РѕРґРёРј РїРѕР·РёС†РёСЋ fromPerson РІ РІРµСЂС€РёРЅРµ
			for i, p := range fromNode.People {
				if p.ID == fromPerson.ID {
					fromPersonIndex = i
					break
				}
			}

			// Р•СЃР»Рё РІ РІРµСЂС€РёРЅРµ 2 С‡РµР»РѕРІРµРєР°, СЃРјРѕС‚СЂРёРј РЅР° Up РґСЂСѓРіРѕРіРѕ
			if len(fromNode.People) == 2 && fromPersonIndex >= 0 {
				otherIndex := 1 - fromPersonIndex // 0 -> 1, 1 -> 0
				if otherIndex < len(fromNode.Up) && fromNode.Up[otherIndex] != nil {
					siblingUp = fromNode.Up[otherIndex]
				}
			}

			// Р”РѕР±Р°РІР»СЏРµРј РїР°СЂСѓ СЂРѕРґРёС‚РµР»РµР№ РєР°Рє РѕРґРЅСѓ РІРµСЂС€РёРЅСѓ
			if direction == stage1_input.PlacedLeft {
				om.AddParentPairLeft(fromNode, addedPerson, addedPerson2, fromLayer, targetLayer, siblingUp, fromPersonIndex)
			} else {
				om.AddParentPairRight(fromNode, addedPerson, addedPerson2, fromLayer, targetLayer, siblingUp, fromPersonIndex)
			}
			continue
		}

		// РћР±С‹С‡РЅРѕРµ РґРѕР±Р°РІР»РµРЅРёРµ РѕРґРЅРѕРіРѕ С‡РµР»РѕРІРµРєР°
		// РќР°С…РѕРґРёРј РїРѕР·РёС†РёСЋ fromPerson РІ РІРµСЂС€РёРЅРµ (РµСЃР»Рё РµС‰С‘ РЅРµ РІС‹С‡РёСЃР»РµРЅР°)
		var fromPersonIndex int = -1
		for i, p := range fromNode.People {
			if p.ID == fromPerson.ID {
				fromPersonIndex = i
				break
			}
		}
		if fromPersonIndex < 0 {
			fromPersonIndex = 0 // РџРѕ СѓРјРѕР»С‡Р°РЅРёСЋ
		}

		// Р”РѕР±Р°РІР»СЏРµРј РІ Р·Р°РІРёСЃРёРјРѕСЃС‚Рё РѕС‚ РЅР°РїСЂР°РІР»РµРЅРёСЏ
		if direction == stage1_input.PlacedLeft {
			om.AddPersonLeft(fromNode, addedPerson, fromLayer, targetLayer, fromPersonIndex)
		} else {
			om.AddPersonRight(fromNode, addedPerson, fromLayer, targetLayer, fromPersonIndex)
		}
	}

	return om
}
