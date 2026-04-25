package stage3_ordering

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

// CoordNode вЂ” РІРµСЂС€РёРЅР° СЃ РєРѕРѕСЂРґРёРЅР°С‚Р°РјРё РґР»СЏ РѕС‚СЂРёСЃРѕРІРєРё
type CoordNode struct {
	// РљРѕРѕСЂРґРёРЅР°С‚С‹ РіСЂР°РЅРёС†
	Left  int
	Right int

	// РЎРїРёСЃРѕРє Р»СЋРґРµР№ (РїСѓСЃС‚РѕР№ РґР»СЏ РїСЃРµРІРґРѕРІРµСЂС€РёРЅ)
	People []*stage1_input.Person

	// IsPseudo вЂ” РїСЃРµРІРґРѕРІРµСЂС€РёРЅР°
	IsPseudo bool

	// Р’РµСЂС‚РёРєР°Р»СЊРЅС‹Рµ СЃРІСЏР·Рё РІРІРµСЂС… вЂ” РґР»СЏ РєР°Р¶РґРѕРіРѕ С‡РµР»РѕРІРµРєР° (РёР»Рё РѕРґРёРЅ РґР»СЏ РїСЃРµРІРґРѕ)
	Up []*CoordNode

	// Р’РµСЂС‚РёРєР°Р»СЊРЅС‹Рµ СЃРІСЏР·Рё РІРЅРёР· вЂ” РґР»СЏ РєР°Р¶РґРѕРіРѕ СЂРµР±С‘РЅРєР°
	Down []*CoordNode

	// РќРѕРјРµСЂ СЃР»РѕСЏ
	Layer int

	// РЎСЃС‹Р»РєР° РЅР° РѕСЂРёРіРёРЅР°Р»СЊРЅСѓСЋ РІРµСЂС€РёРЅСѓ (РґР»СЏ РѕС‚Р»Р°РґРєРё)
	OriginalNode *LayerNode

	// Р¤Р»Р°Рі: Р±С‹Р»Р° Р»Рё СЌС‚Р° РІРµСЂС€РёРЅР° С‡Р°СЃС‚СЊСЋ СЃРєР»РµРµРЅРЅРѕР№ РїР°СЂС‹
	WasMerged bool
	// РЎСЃС‹Р»РєР° РЅР° РїР°СЂС‚РЅС‘СЂР° (РµСЃР»Рё Р±С‹Р»Р° СЃРєР»РµРµРЅР°)
	MergePartner *CoordNode

	// ParentNodes вЂ” СѓРєР°Р·Р°С‚РµР»Рё РЅР° РІРµСЂС€РёРЅС‹ СЂРѕРґРёС‚РµР»РµР№ РґР»СЏ РєР°Р¶РґРѕРіРѕ С‡РµР»РѕРІРµРєР°
	// ParentNodes[i] СЃРѕРѕС‚РІРµС‚СЃС‚РІСѓРµС‚ People[i]
	ParentNodes []*CoordNode

	// ParentPersonIndex вЂ” РёРЅРґРµРєСЃ С‡РµР»РѕРІРµРєР° РІ СЂРѕРґРёС‚РµР»СЊСЃРєРѕР№ РІРµСЂС€РёРЅРµ
	// ParentPersonIndex[i] СЃРѕРѕС‚РІРµС‚СЃС‚РІСѓРµС‚ People[i]
	ParentPersonIndex []int

	// Children вЂ” СЃРїРёСЃРѕРє РґРѕС‡РµСЂРЅРёС… РІРµСЂС€РёРЅ (РґР»СЏ РІРµСЂС€РёРЅС‹ РєР°Рє СЂРѕРґРёС‚РµР»СЏ)
	Children []*CoordNode

	// AddedLeft вЂ” С„Р»Р°Рі, Р±С‹Р»Р° Р»Рё РІРµСЂС€РёРЅР° РґРѕР±Р°РІР»РµРЅР° СЃР»РµРІР°
	AddedLeft bool
}

// Width РІРѕР·РІСЂР°С‰Р°РµС‚ С€РёСЂРёРЅСѓ РІРµСЂС€РёРЅС‹
func (cn *CoordNode) Width() int {
	return cn.Right - cn.Left
}

// Center РІРѕР·РІСЂР°С‰Р°РµС‚ С†РµРЅС‚СЂ РІРµСЂС€РёРЅС‹
func (cn *CoordNode) Center() int {
	return (cn.Left + cn.Right) / 2
}
