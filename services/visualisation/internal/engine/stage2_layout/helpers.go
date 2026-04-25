package stage2_layout

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

// VisitedSet РїСЂРµРґСЃС‚Р°РІР»СЏРµС‚ РјРЅРѕР¶РµСЃС‚РІРѕ РїРѕСЃРµС‰С‘РЅРЅС‹С… РІРµСЂС€РёРЅ РґР»СЏ DFS
type VisitedSet map[int]bool

// NewVisitedSet СЃРѕР·РґР°С‘С‚ РЅРѕРІРѕРµ РїСѓСЃС‚РѕРµ РјРЅРѕР¶РµСЃС‚РІРѕ
func NewVisitedSet() VisitedSet {
	return make(VisitedSet)
}

// Add РґРѕР±Р°РІР»СЏРµС‚ ID РІ РјРЅРѕР¶РµСЃС‚РІРѕ
func (v VisitedSet) Add(id int) {
	v[id] = true
}

// Contains РїСЂРѕРІРµСЂСЏРµС‚, СЃРѕРґРµСЂР¶РёС‚СЃСЏ Р»Рё ID РІ РјРЅРѕР¶РµСЃС‚РІРµ
func (v VisitedSet) Contains(id int) bool {
	return v[id]
}

// GetCommonChildren РІРѕР·РІСЂР°С‰Р°РµС‚ РѕР±С‰РёС… РґРµС‚РµР№ РґРІСѓС… СЂРѕРґРёС‚РµР»РµР№
// РџРѕСЂСЏРґРѕРє РґРµС‚РµР№ Р±РµСЂС‘С‚СЃСЏ РёР· СЃРїРёСЃРєР° Children РїРµСЂРІРѕРіРѕ СЂРѕРґРёС‚РµР»СЏ
func GetCommonChildren(parent1, parent2 *stage1_input.Person) []*stage1_input.Person {
	result := make([]*stage1_input.Person, 0)

	// РЎРѕР·РґР°С‘Рј РјРЅРѕР¶РµСЃС‚РІРѕ ID РґРµС‚РµР№ РІС‚РѕСЂРѕРіРѕ СЂРѕРґРёС‚РµР»СЏ РґР»СЏ Р±С‹СЃС‚СЂРѕРіРѕ РїРѕРёСЃРєР°
	parent2ChildrenIDs := make(map[int]bool)
	for _, child := range parent2.Children {
		parent2ChildrenIDs[child.ID] = true
	}

	// РџСЂРѕС…РѕРґРёРј РїРѕ РґРµС‚СЏРј РїРµСЂРІРѕРіРѕ СЂРѕРґРёС‚РµР»СЏ Рё РїСЂРѕРІРµСЂСЏРµРј РЅР°Р»РёС‡РёРµ РІРѕ РІС‚РѕСЂРѕРј
	for _, child := range parent1.Children {
		if parent2ChildrenIDs[child.ID] {
			result = append(result, child)
		}
	}

	return result
}

// IsCurrentPartner РїСЂРѕРІРµСЂСЏРµС‚, СЏРІР»СЏРµС‚СЃСЏ Р»Рё РїР°СЂС‚РЅС‘СЂ С‚РµРєСѓС‰РёРј (РїРµСЂРІС‹Рј РІ СЃРїРёСЃРєРµ)
func IsCurrentPartner(person, partner *stage1_input.Person) bool {
	if len(person.Partners) == 0 {
		return false
	}
	return person.Partners[0].ID == partner.ID
}

// ShouldAddPartnerLeft РѕРїСЂРµРґРµР»СЏРµС‚, РЅСѓР¶РЅРѕ Р»Рё РґРѕР±Р°РІРёС‚СЊ РїР°СЂС‚РЅС‘СЂР° СЃР»РµРІР°
// Р’РѕР·РІСЂР°С‰Р°РµС‚ true РµСЃР»Рё СЃР»РµРІР°, false РµСЃР»Рё СЃРїСЂР°РІР°
func ShouldAddPartnerLeft(person, partner *stage1_input.Person, dirConstraint stage1_input.DirectionConstraint) bool {
	// РџСЂРёРѕСЂРёС‚РµС‚ 1: DirectionConstraint
	if dirConstraint == stage1_input.OnlyLeft {
		return true
	}
	if dirConstraint == stage1_input.OnlyRight {
		return false
	}

	// РџСЂРёРѕСЂРёС‚РµС‚ 2: РЎС‚Р°РЅРґР°СЂС‚РЅС‹Рµ РїСЂР°РІРёР»Р°
	isCurrent := IsCurrentPartner(person, partner)

	if person.Gender == stage1_input.Male {
		// РњСѓР¶С‡РёРЅР°: С‚РµРєСѓС‰Р°СЏ Р¶РµРЅР° СЃРїСЂР°РІР°, Р±С‹РІС€РёРµ Р¶С‘РЅС‹ СЃР»РµРІР°
		if isCurrent {
			return false // СЃРїСЂР°РІР°
		}
		return true // СЃР»РµРІР°
	} else {
		// Р–РµРЅС‰РёРЅР°: С‚РµРєСѓС‰РёР№ РјСѓР¶ СЃР»РµРІР°, Р±С‹РІС€РёРµ РјСѓР¶СЊСЏ СЃРїСЂР°РІР°
		if isCurrent {
			return true // СЃР»РµРІР°
		}
		return false // СЃРїСЂР°РІР°
	}
}
