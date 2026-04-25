package stage2_layout

import (
	"fmt"

	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"
)

// LayoutFromPerson Р·Р°РїСѓСЃРєР°РµС‚ Р°Р»РіРѕСЂРёС‚Рј СЂР°Р·РјРµС‰РµРЅРёСЏ РЅР°С‡РёРЅР°СЏ СЃ СѓРєР°Р·Р°РЅРЅРѕРіРѕ С‡РµР»РѕРІРµРєР°
// Р’РѕР·РІСЂР°С‰Р°РµС‚ РёСЃС‚РѕСЂРёСЋ СЂР°Р·РјРµС‰РµРЅРёР№
func LayoutFromPerson(tree *stage1_input.FamilyTree, startPersonID int) (*stage1_input.PlacementHistory, error) {
	startPerson := tree.GetPerson(startPersonID)
	if startPerson == nil {
		return nil, fmt.Errorf("person with ID %d not found", startPersonID)
	}

	// РЎРѕР·РґР°С‘Рј РёСЃС‚РѕСЂРёСЋ СЂР°Р·РјРµС‰РµРЅРёР№
	history := stage1_input.NewPlacementHistory()

	// РРЅРёС†РёР°Р»РёР·РёСЂСѓРµРј РЅР°С‡Р°Р»СЊРЅРѕРіРѕ С‡РµР»РѕРІРµРєР°
	startPerson.Layout = stage1_input.NewPersonLayout(0)
	startPerson.Layout.IsStartPerson = true

	// РЎРѕР·РґР°С‘Рј РѕС‡РµСЂРµРґСЊ Рё РґРѕР±Р°РІР»СЏРµРј РЅР°С‡Р°Р»СЊРЅРѕРіРѕ С‡РµР»РѕРІРµРєР°
	queue := NewQueue()
	queue.Enqueue(startPerson)

	// BFS РѕР±С…РѕРґ
	for !queue.IsEmpty() {
		person := queue.Dequeue()

		// Р•СЃР»Рё СѓР¶Рµ РѕР±СЂР°Р±РѕС‚Р°РЅ вЂ” РїСЂРѕРїСѓСЃРєР°РµРј
		if person.IsProcessed() {
			continue
		}

		// РћР±СЂР°Р±Р°С‚С‹РІР°РµРј С‡РµР»РѕРІРµРєР°
		processPersonLayout(person, queue, history)

		// РџРѕРјРµС‡Р°РµРј РєР°Рє РѕР±СЂР°Р±РѕС‚Р°РЅРЅРѕРіРѕ
		person.Layout.Processed = true
	}

	return history, nil
}

// processPersonLayout РѕР±СЂР°Р±Р°С‚С‹РІР°РµС‚ РѕРґРЅРѕРіРѕ С‡РµР»РѕРІРµРєР°: РґРѕР±Р°РІР»СЏРµС‚ СЂРѕРґРёС‚РµР»РµР№, РїР°СЂС‚РЅС‘СЂРѕРІ, РґРµС‚РµР№
func processPersonLayout(person *stage1_input.Person, queue *Queue, history *stage1_input.PlacementHistory) {
	// 1. Р”РѕР±Р°РІР»СЏРµРј СЂРѕРґРёС‚РµР»РµР№
	// Р•СЃР»Рё РЅРµ СѓРґР°Р»РѕСЃСЊ (РёР·-Р·Р° РѕРіСЂР°РЅРёС‡РµРЅРёР№) вЂ” РѕРїСѓСЃРєР°РµРј Рё РїСЂРѕР±СѓРµРј СЃРЅРѕРІР°
	for !AddParents(person, queue, history) {
		// AddParents СѓР¶Рµ РІС‹Р·РІР°Р» LowerSubtree, РїСЂРѕР±СѓРµРј СЃРЅРѕРІР°
	}

	// 2. Р”РѕР±Р°РІР»СЏРµРј РїР°СЂС‚РЅС‘СЂРѕРІ Рё РёС… РѕР±С‰РёС… РґРµС‚РµР№
	AddPartners(person, queue, history)
}
