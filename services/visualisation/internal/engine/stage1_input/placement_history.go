package stage1_input

import "fmt"

// PlacementDirection РѕРїСЂРµРґРµР»СЏРµС‚ РЅР°РїСЂР°РІР»РµРЅРёРµ СЂР°Р·РјРµС‰РµРЅРёСЏ
type PlacementDirection int

const (
	PlacedLeft PlacementDirection = iota
	PlacedRight
)

func (d PlacementDirection) String() string {
	if d == PlacedLeft {
		return "СЃР»РµРІР°"
	}
	return "СЃРїСЂР°РІР°"
}

// RelationType РѕРїСЂРµРґРµР»СЏРµС‚ С‚РёРї СЂРѕРґСЃС‚РІРµРЅРЅРѕР№ СЃРІСЏР·Рё РїСЂРё РґРѕР±Р°РІР»РµРЅРёРё
type RelationType int

const (
	RelationParent  RelationType = iota // Р РѕРґРёС‚РµР»СЊ (+1 СЃР»РѕР№)
	RelationPartner                     // РџР°СЂС‚РЅС‘СЂ (С‚РѕС‚ Р¶Рµ СЃР»РѕР№)
	RelationChild                       // Р РµР±С‘РЅРѕРє (-1 СЃР»РѕР№)
	RelationSibling                     // Р‘СЂР°С‚/СЃРµСЃС‚СЂР° (С‚РѕС‚ Р¶Рµ СЃР»РѕР№)
)

func (r RelationType) String() string {
	switch r {
	case RelationParent:
		return "СЂРѕРґРёС‚РµР»СЊ"
	case RelationPartner:
		return "РїР°СЂС‚РЅС‘СЂ"
	case RelationChild:
		return "СЂРµР±С‘РЅРѕРє"
	case RelationSibling:
		return "Р±СЂР°С‚/СЃРµСЃС‚СЂР°"
	default:
		return "РЅРµРёР·РІРµСЃС‚РЅРѕ"
	}
}

// PlacementRecord С…СЂР°РЅРёС‚ РёРЅС„РѕСЂРјР°С†РёСЋ Рѕ РґРѕР±Р°РІР»РµРЅРёРё РІРµСЂС€РёРЅС‹
type PlacementRecord struct {
	FromPerson   *Person            // РљС‚Рѕ РґРѕР±Р°РІР»СЏРµС‚
	AddedPerson  *Person            // РљРѕРіРѕ РґРѕР±Р°РІРёР»Рё (РїРµСЂРІС‹Р№ С‡РµР»РѕРІРµРє)
	AddedPerson2 *Person            // Р’С‚РѕСЂРѕР№ РґРѕР±Р°РІР»РµРЅРЅС‹Р№ (РґР»СЏ РїР°СЂС‹ СЂРѕРґРёС‚РµР»РµР№), РјРѕР¶РµС‚ Р±С‹С‚СЊ nil
	LayerDiff    int                // Р Р°Р·РЅРёС†Р° СЃР»РѕС‘РІ: +1 (СЂРѕРґРёС‚РµР»СЊ), 0 (РїР°СЂС‚РЅС‘СЂ/Р±СЂР°С‚), -1 (СЂРµР±С‘РЅРѕРє)
	Direction    PlacementDirection // РЎР»РµРІР° РёР»Рё СЃРїСЂР°РІР°
	RelationType RelationType       // РўРёРї РѕС‚РЅРѕС€РµРЅРёСЏ
}

// PlacementHistory С…СЂР°РЅРёС‚ РёСЃС‚РѕСЂРёСЋ РґРѕР±Р°РІР»РµРЅРёР№
type PlacementHistory struct {
	Records []*PlacementRecord
}

// NewPlacementHistory СЃРѕР·РґР°С‘С‚ РЅРѕРІСѓСЋ РёСЃС‚РѕСЂРёСЋ
func NewPlacementHistory() *PlacementHistory {
	return &PlacementHistory{
		Records: make([]*PlacementRecord, 0),
	}
}

// Add РґРѕР±Р°РІР»СЏРµС‚ Р·Р°РїРёСЃСЊ РІ РёСЃС‚РѕСЂРёСЋ
func (h *PlacementHistory) Add(from, added *Person, layerDiff int, direction PlacementDirection, relationType RelationType) {
	h.Records = append(h.Records, &PlacementRecord{
		FromPerson:   from,
		AddedPerson:  added,
		AddedPerson2: nil,
		LayerDiff:    layerDiff,
		Direction:    direction,
		RelationType: relationType,
	})
}

// AddPair РґРѕР±Р°РІР»СЏРµС‚ Р·Р°РїРёСЃСЊ Рѕ РґРѕР±Р°РІР»РµРЅРёРё РїР°СЂС‹ (РґР»СЏ СЂРѕРґРёС‚РµР»РµР№)
func (h *PlacementHistory) AddPair(from, added1, added2 *Person, layerDiff int, direction PlacementDirection, relationType RelationType) {
	h.Records = append(h.Records, &PlacementRecord{
		FromPerson:   from,
		AddedPerson:  added1,
		AddedPerson2: added2,
		LayerDiff:    layerDiff,
		Direction:    direction,
		RelationType: relationType,
	})
}

// Print РІС‹РІРѕРґРёС‚ РёСЃС‚РѕСЂРёСЋ РґРѕР±Р°РІР»РµРЅРёР№
func (h *PlacementHistory) Print() {
	fmt.Println("\nРСЃС‚РѕСЂРёСЏ РґРѕР±Р°РІР»РµРЅРёР№:")
	fmt.Println("====================")
	for _, r := range h.Records {
		if r.AddedPerson2 != nil {
			fmt.Printf("%s (ID=%d) -> [%s (ID=%d), %s (ID=%d)]: %s, %s\n",
				r.FromPerson.Name, r.FromPerson.ID,
				r.AddedPerson.Name, r.AddedPerson.ID,
				r.AddedPerson2.Name, r.AddedPerson2.ID,
				r.RelationType, r.Direction)
		} else {
			fmt.Printf("%s (ID=%d) -> %s (ID=%d): %s, %s\n",
				r.FromPerson.Name, r.FromPerson.ID,
				r.AddedPerson.Name, r.AddedPerson.ID,
				r.RelationType, r.Direction)
		}
	}
}
