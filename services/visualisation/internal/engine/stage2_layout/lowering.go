package stage2_layout

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

// LowerSubtree РѕРїСѓСЃРєР°РµС‚ РїРѕРґРґРµСЂРµРІРѕ РЅР° РѕРґРёРЅ СЃР»РѕР№ РІРЅРёР·
// isInitial = true РґР»СЏ РЅР°С‡Р°Р»СЊРЅРѕР№ РІРµСЂС€РёРЅС‹ (РЅРµ РѕРїСѓСЃРєР°РµРј СЂРѕРґРёС‚РµР»РµР№)
// isInitial = false РґР»СЏ РІСЃРµС… РѕСЃС‚Р°Р»СЊРЅС‹С… РІРµСЂС€РёРЅ (РѕРїСѓСЃРєР°РµРј РІСЃРµС… СЂРѕРґСЃС‚РІРµРЅРЅРёРєРѕРІ)
func LowerSubtree(person *stage1_input.Person, isInitial bool, visited VisitedSet) {
	// Р•СЃР»Рё РІРµСЂС€РёРЅР° СѓР¶Рµ РїРѕСЃРµС‰РµРЅР° вЂ” РІС‹С…РѕРґ
	if visited.Contains(person.ID) {
		return
	}

	// Р•СЃР»Рё Сѓ С‡РµР»РѕРІРµРєР° РЅРµС‚ layout вЂ” РїСЂРѕРїСѓСЃРєР°РµРј
	if person.Layout == nil {
		return
	}

	// РџРѕРјРµС‡Р°РµРј РєР°Рє РїРѕСЃРµС‰С‘РЅРЅСѓСЋ
	visited.Add(person.ID)

	// РЈРјРµРЅСЊС€Р°РµРј СЃР»РѕР№ РЅР° 1
	person.Layout.Layer--

	// Р РµРєСѓСЂСЃРёРІРЅРѕ РѕР±СЂР°Р±Р°С‚С‹РІР°РµРј СЂРѕРґСЃС‚РІРµРЅРЅРёРєРѕРІ
	// РџР°СЂС‚РЅС‘СЂС‹ вЂ” РІСЃРµРіРґР°
	for _, partner := range person.Partners {
		if partner.Layout != nil {
			LowerSubtree(partner, false, visited)
		}
	}

	// Р”РµС‚Рё вЂ” РІСЃРµРіРґР°
	for _, child := range person.Children {
		if child.Layout != nil {
			LowerSubtree(child, false, visited)
		}
	}

	// Р РѕРґРёС‚РµР»Рё вЂ” С‚РѕР»СЊРєРѕ РµСЃР»Рё РЅРµ РЅР°С‡Р°Р»СЊРЅР°СЏ РІРµСЂС€РёРЅР°
	if !isInitial {
		if person.Mother != nil && person.Mother.Layout != nil {
			LowerSubtree(person.Mother, false, visited)
		}
		if person.Father != nil && person.Father.Layout != nil {
			LowerSubtree(person.Father, false, visited)
		}
	}
}
