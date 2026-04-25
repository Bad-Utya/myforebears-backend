package stage4_render

import (
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage3_ordering"
)

// Coord РїСЂРµРґСЃС‚Р°РІР»СЏРµС‚ РєРѕРѕСЂРґРёРЅР°С‚Сѓ РІРµСЂС€РёРЅС‹
type Coord struct {
	X int // РёРЅРґРµРєСЃ РІ СЃР»РѕРµ (0 = СЃР°РјС‹Р№ Р»РµРІС‹Р№)
	Y int // РЅРѕРјРµСЂ СЃР»РѕСЏ
}

// Edge РїСЂРµРґСЃС‚Р°РІР»СЏРµС‚ СЃРІСЏР·СЊ РјРµР¶РґСѓ РґРІСѓРјСЏ РІРµСЂС€РёРЅР°РјРё
type Edge struct {
	From     Coord
	To       Coord
	EdgeType string // "parent-child", "partner"
}

// NodeInfo СЃРѕРґРµСЂР¶РёС‚ РёРЅС„РѕСЂРјР°С†РёСЋ Рѕ РІРµСЂС€РёРЅРµ РґР»СЏ СЂРµРЅРґРµСЂРёРЅРіР°
type NodeInfo struct {
	Left            int                    // Р»РµРІР°СЏ РіСЂР°РЅРёС†Р°
	Right           int                    // РїСЂР°РІР°СЏ РіСЂР°РЅРёС†Р°
	Layer           int                    // СЃР»РѕР№
	People          []*stage1_input.Person // Р»СЋРґРё РІ РІРµСЂС€РёРЅРµ
	MergePartnerIdx int                    // РёРЅРґРµРєСЃ РІРµСЂС€РёРЅС‹-РїР°СЂС‚РЅС‘СЂР° (-1 РµСЃР»Рё РЅРµС‚)
	AddedLeft       bool                   // Р±С‹Р»Р° Р»Рё РІРµСЂС€РёРЅР° РґРѕР±Р°РІР»РµРЅР° СЃР»РµРІР°
}

// EdgeInfo СЃРѕРґРµСЂР¶РёС‚ РёРЅС„РѕСЂРјР°С†РёСЋ Рѕ СЃРІСЏР·Рё РґР»СЏ СЂРµРЅРґРµСЂРёРЅРіР°
type EdgeInfo struct {
	FromNodeIdx int    // РёРЅРґРµРєСЃ СѓР·Р»Р°-РёСЃС‚РѕС‡РЅРёРєР° РІ Nodes
	ToNodeIdx   int    // РёРЅРґРµРєСЃ СѓР·Р»Р°-РїСЂРёС‘РјРЅРёРєР° РІ Nodes
	FromX       int    // X-РєРѕРѕСЂРґРёРЅР°С‚Р° РЅР°С‡Р°Р»Р° (РёР· CoordMatrix)
	FromY       int    // Y-РєРѕРѕСЂРґРёРЅР°С‚Р° (СЃР»РѕР№) РЅР°С‡Р°Р»Р°
	ToX         int    // X-РєРѕРѕСЂРґРёРЅР°С‚Р° РєРѕРЅС†Р°
	ToY         int    // Y-РєРѕРѕСЂРґРёРЅР°С‚Р° (СЃР»РѕР№) РєРѕРЅС†Р°
	EdgeType    string // "parent-child", "partner"

	// Р”Р»СЏ parent-child СЃРІСЏР·РµР№:
	ParentsAdjacent bool // СЂРѕРґРёС‚РµР»Рё вЂ” СЃРјРµР¶РЅС‹Рµ РІРµСЂС€РёРЅС‹ (MergePartner)
	AdjacentCenterX int  // X-РєРѕРѕСЂРґРёРЅР°С‚Р° С†РµРЅС‚СЂР° РјРµР¶РґСѓ СЃРјРµР¶РЅС‹РјРё СЂРѕРґРёС‚РµР»СЏРјРё
	ParentAddedLeft bool // СЂРѕРґРёС‚РµР»СЊ Р±С‹Р» РґРѕР±Р°РІР»РµРЅ СЃР»РµРІР°
}

// RenderResult СЃРѕРґРµСЂР¶РёС‚ СЂРµР·СѓР»СЊС‚Р°С‚ РІРёР·СѓР°Р»РёР·Р°С†РёРё
type RenderResult struct {
	Coords map[int]Coord // ID С‡РµР»РѕРІРµРєР° -> РєРѕРѕСЂРґРёРЅР°С‚Р°
	Edges  []Edge        // СЃРїРёСЃРѕРє СЃРІСЏР·РµР№
}

// CoordRenderResult СЃРѕРґРµСЂР¶РёС‚ СЂРµР·СѓР»СЊС‚Р°С‚ РІРёР·СѓР°Р»РёР·Р°С†РёРё РёР· CoordMatrix
type CoordRenderResult struct {
	Nodes    []NodeInfo // СЃРїРёСЃРѕРє РІРµСЂС€РёРЅ
	Edges    []EdgeInfo // СЃРїРёСЃРѕРє СЃРІСЏР·РµР№
	MinLayer int        // РјРёРЅРёРјР°Р»СЊРЅС‹Р№ СЃР»РѕР№
	MaxLayer int        // РјР°РєСЃРёРјР°Р»СЊРЅС‹Р№ СЃР»РѕР№
	MaxRight int        // РјР°РєСЃРёРјР°Р»СЊРЅР°СЏ РїСЂР°РІР°СЏ РіСЂР°РЅРёС†Р°
}

// PersonPosition РёСЃРїРѕР»СЊР·СѓРµС‚СЃСЏ РґР»СЏ РїРѕСЃС‚СЂРѕРµРЅРёСЏ RenderResult
type PersonPosition struct {
	Person *stage1_input.Person
	X      int
	Y      int
}

// PeopleGrid РїСЂРµРґСЃС‚Р°РІР»СЏРµС‚ СЃРµС‚РєСѓ Р»СЋРґРµР№ РїРѕ СЃР»РѕСЏРј
type PeopleGrid struct {
	Layers    map[int][]PersonPosition
	Positions []PersonPosition
}

// isInParentNodes РїСЂРѕРІРµСЂСЏРµС‚, СЏРІР»СЏРµС‚СЃСЏ Р»Рё parentNode РѕРґРЅРёРј РёР· ParentNodes СЂРµР±С‘РЅРєР°
// РёР»Рё РµРіРѕ MergePartner. РЎСЂР°РІРЅРµРЅРёРµ РїСЂРѕРёСЃС…РѕРґРёС‚ РїРѕ ID Р»СЋРґРµР№, Р° РЅРµ РїРѕ СѓРєР°Р·Р°С‚РµР»СЏРј,
// РїРѕС‚РѕРјСѓ С‡С‚Рѕ РїРѕСЃР»Рµ split СѓРєР°Р·Р°С‚РµР»Рё РјРѕРіСѓС‚ РёР·РјРµРЅРёС‚СЊСЃСЏ
func isInParentNodes(childNode, parentNode *stage3_ordering.CoordNode) bool {
	// РЎРѕР±РёСЂР°РµРј РІСЃРµ ID Р»СЋРґРµР№ РІ parentNode
	parentIDs := make(map[int]bool)
	for _, person := range parentNode.People {
		parentIDs[person.ID] = true
	}
	// РўР°РєР¶Рµ РІРєР»СЋС‡Р°РµРј ID Р»СЋРґРµР№ РёР· MergePartner
	if parentNode.MergePartner != nil {
		for _, person := range parentNode.MergePartner.People {
			parentIDs[person.ID] = true
		}
	}

	// РџСЂРѕРІРµСЂСЏРµРј, СЃРѕРґРµСЂР¶РёС‚ Р»Рё РєР°РєРѕР№-Р»РёР±Рѕ ParentNodes СЂРµР±С‘РЅРєР° СЌС‚РёС… Р»СЋРґРµР№
	for _, pn := range childNode.ParentNodes {
		if pn == nil {
			continue
		}
		for _, person := range pn.People {
			if parentIDs[person.ID] {
				return true
			}
		}
		// РўР°РєР¶Рµ РїСЂРѕРІРµСЂСЏРµРј MergePartner
		if pn.MergePartner != nil {
			for _, person := range pn.MergePartner.People {
				if parentIDs[person.ID] {
					return true
				}
			}
		}
	}
	return false
}

// addParentChildEdge РґРѕР±Р°РІР»СЏРµС‚ СЃРІСЏР·СЊ СЂРѕРґРёС‚РµР»СЊ-СЂРµР±С‘РЅРѕРє РІ СЂРµР·СѓР»СЊС‚Р°С‚
func addParentChildEdge(result *CoordRenderResult, parentNode, childNode *stage3_ordering.CoordNode, parentIdx, childIdx int, addedEdges *map[string]bool) {
	// РљРѕРѕСЂРґРёРЅР°С‚Р° СЂРѕРґРёС‚РµР»СЏ (С†РµРЅС‚СЂ РІРµСЂС€РёРЅС‹)
	fromX := (parentNode.Left + parentNode.Right) / 2

	// РљРѕРѕСЂРґРёРЅР°С‚Р° СЂРµР±С‘РЅРєР° (С†РµРЅС‚СЂ РІРµСЂС€РёРЅС‹)
	toX := (childNode.Left + childNode.Right) / 2

	// РџСЂРѕРІРµСЂСЏРµРј, РµСЃС‚СЊ Р»Рё СЃРјРµР¶РЅС‹Р№ РїР°СЂС‚РЅС‘СЂ (MergePartner)
	parentsAdjacent := false
	adjacentCenterX := 0
	if parentNode.MergePartner != nil {
		parentsAdjacent = true
		// Р¦РµРЅС‚СЂ РјРµР¶РґСѓ СЃРјРµР¶РЅС‹РјРё РІРµСЂС€РёРЅР°РјРё
		left := parentNode.Left
		right := parentNode.Right
		if parentNode.MergePartner.Left < left {
			left = parentNode.MergePartner.Left
		}
		if parentNode.MergePartner.Right > right {
			right = parentNode.MergePartner.Right
		}
		adjacentCenterX = (left + right) / 2
	}

	result.Edges = append(result.Edges, EdgeInfo{
		FromNodeIdx:     parentIdx,
		ToNodeIdx:       childIdx,
		FromX:           fromX,
		FromY:           parentNode.Layer,
		ToX:             toX,
		ToY:             childNode.Layer,
		EdgeType:        "parent-child",
		ParentsAdjacent: parentsAdjacent,
		AdjacentCenterX: adjacentCenterX,
		ParentAddedLeft: parentNode.AddedLeft,
	})
}

// getChildXCoord РІРѕР·РІСЂР°С‰Р°РµС‚ X РєРѕРѕСЂРґРёРЅР°С‚Сѓ РєРѕРЅРєСЂРµС‚РЅРѕРіРѕ СЂРµР±С‘РЅРєР° РІ РІРµСЂС€РёРЅРµ РїРѕ РµРіРѕ ID
func getChildXCoord(node *stage3_ordering.CoordNode, personID int) int {
	if len(node.People) == 1 {
		return (node.Left + node.Right) / 2
	}
	// РЎРєР»РµРµРЅРЅР°СЏ РІРµСЂС€РёРЅР° - РёС‰РµРј РїРѕР·РёС†РёСЋ С‡РµР»РѕРІРµРєР°
	for i, p := range node.People {
		if p.ID == personID {
			if i == 0 {
				return node.Left + 1
			}
			return node.Right - 1
		}
	}
	// РќРµ РЅР°Р№РґРµРЅ - РІРѕР·РІСЂР°С‰Р°РµРј С†РµРЅС‚СЂ
	return (node.Left + node.Right) / 2
}

// getPersonXCoord РІРѕР·РІСЂР°С‰Р°РµС‚ X РєРѕРѕСЂРґРёРЅР°С‚Сѓ РєРѕРЅРєСЂРµС‚РЅРѕРіРѕ С‡РµР»РѕРІРµРєР° РІ РІРµСЂС€РёРЅРµ
func getPersonXCoord(node *stage3_ordering.CoordNode, personIndex int) int {
	if len(node.People) == 1 {
		return (node.Left + node.Right) / 2
	}
	// РЎРєР»РµРµРЅРЅР°СЏ РІРµСЂС€РёРЅР°
	if personIndex == 0 {
		return node.Left + 1
	}
	return node.Right - 1
}

// getPersonXCoordByIndex РІРѕР·РІСЂР°С‰Р°РµС‚ X РєРѕРѕСЂРґРёРЅР°С‚Сѓ С‡РµР»РѕРІРµРєР° РїРѕ РёРЅРґРµРєСЃСѓ РІ РІРµСЂС€РёРЅРµ
func getPersonXCoordByIndex(node *stage3_ordering.CoordNode, index int) int {
	if len(node.People) == 1 {
		return (node.Left + node.Right) / 2
	}
	// РЎРєР»РµРµРЅРЅР°СЏ РІРµСЂС€РёРЅР°
	if index == 0 {
		return node.Left + 1
	}
	return node.Right - 1
}

// findPersonIndexInNode РЅР°С…РѕРґРёС‚ РёРЅРґРµРєСЃ С‡РµР»РѕРІРµРєР° РїРѕ ID РІ РІРµСЂС€РёРЅРµ
func findPersonIndexInNode(node *stage3_ordering.CoordNode, personID int) int {
	for i, p := range node.People {
		if p.ID == personID {
			return i
		}
	}
	return 0
}

// containsSamePerson РїСЂРѕРІРµСЂСЏРµС‚, РµСЃС‚СЊ Р»Рё С…РѕС‚СЏ Р±С‹ РѕРґРёРЅ РѕР±С‰РёР№ С‡РµР»РѕРІРµРє РІ РґРІСѓС… СЃРїРёСЃРєР°С…
func containsSamePerson(people1, people2 []*stage1_input.Person) bool {
	for _, p1 := range people1 {
		for _, p2 := range people2 {
			if p1.ID == p2.ID {
				return true
			}
		}
	}
	return false
}
