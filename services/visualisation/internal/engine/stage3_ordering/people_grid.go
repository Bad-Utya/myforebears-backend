package stage3_ordering

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

// PersonPosition РїСЂРµРґСЃС‚Р°РІР»СЏРµС‚ РїРѕР·РёС†РёСЋ С‡РµР»РѕРІРµРєР° РІ СЃРµС‚РєРµ
type PersonPosition struct {
	Person *stage1_input.Person
	X      int // РёРЅРґРµРєСЃ РІ СЃР»РѕРµ
	Y      int // РЅРѕРјРµСЂ СЃР»РѕСЏ
}

// PeopleGrid РїСЂРµРґСЃС‚Р°РІР»СЏРµС‚ СЃРµС‚РєСѓ Р»СЋРґРµР№ РїРѕ СЃР»РѕСЏРј
type PeopleGrid struct {
	// РЎР»РѕРё (РЅРѕРјРµСЂ СЃР»РѕСЏ -> СЃРїРёСЃРѕРє РїРѕР·РёС†РёР№ Р»СЋРґРµР№)
	Layers map[int][]PersonPosition

	// Р’СЃРµ РїРѕР·РёС†РёРё РІ РІРёРґРµ РїР»РѕСЃРєРѕРіРѕ СЃРїРёСЃРєР°
	Positions []PersonPosition
}

// BuildPeopleGrid СЃРѕР·РґР°С‘С‚ СЃРµС‚РєСѓ Р»СЋРґРµР№ РёР· OrderManager
// РџСЂРѕС…РѕРґРёС‚ РїРѕ РІСЃРµРј СЃР»РѕСЏРј Рё СЃРѕР±РёСЂР°РµС‚ Р»СЋРґРµР№ РёР· РІРµСЂС€РёРЅ
// РџСЃРµРІРґРѕРІРµСЂС€РёРЅС‹ РёРіРЅРѕСЂРёСЂСѓСЋС‚СЃСЏ (С‚.Рє. People РїСѓСЃС‚РѕР№)
// РљР°Р¶РґС‹Р№ С‡РµР»РѕРІРµРє РїРѕР»СѓС‡Р°РµС‚ СЃРІРѕСЋ СѓРЅРёРєР°Р»СЊРЅСѓСЋ РїРѕР·РёС†РёСЋ X
func (om *OrderManager) BuildPeopleGrid() *PeopleGrid {
	grid := &PeopleGrid{
		Layers:    make(map[int][]PersonPosition),
		Positions: []PersonPosition{},
	}

	// РџСЂРѕС…РѕРґРёРј РїРѕ РІСЃРµРј СЃР»РѕСЏРј
	for _, layer := range om.GetAllLayers() {
		positions := []PersonPosition{}
		nodes := layer.GetNodes()
		xIndex := 0

		for _, node := range nodes {
			// РџСЂРѕРїСѓСЃРєР°РµРј РїСЃРµРІРґРѕРІРµСЂС€РёРЅС‹
			if node.IsPseudo {
				continue
			}

			// Р”РѕР±Р°РІР»СЏРµРј РєР°Р¶РґРѕРіРѕ С‡РµР»РѕРІРµРєР° СЃ СѓРЅРёРєР°Р»СЊРЅС‹Рј X
			for _, person := range node.People {
				pos := PersonPosition{
					Person: person,
					X:      xIndex,
					Y:      layer.Number,
				}
				positions = append(positions, pos)
				grid.Positions = append(grid.Positions, pos)
				xIndex++ // РЈРІРµР»РёС‡РёРІР°РµРј X РґР»СЏ РєР°Р¶РґРѕРіРѕ С‡РµР»РѕРІРµРєР°
			}
		}

		grid.Layers[layer.Number] = positions
	}

	return grid
}
