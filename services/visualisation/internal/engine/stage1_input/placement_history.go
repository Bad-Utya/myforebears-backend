package stage1_input

import "fmt"

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

type RelationType int

const (
	RelationParent RelationType = iota
	RelationPartner
	RelationChild
	RelationSibling
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

type PlacementRecord struct {
	FromPerson   *Person
	AddedPerson  *Person
	AddedPerson2 *Person
	LayerDiff    int
	Direction    PlacementDirection
	RelationType RelationType
}

type PlacementHistory struct {
	Records []*PlacementRecord
}

func NewPlacementHistory() *PlacementHistory {
	return &PlacementHistory{
		Records: make([]*PlacementRecord, 0),
	}
}

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
