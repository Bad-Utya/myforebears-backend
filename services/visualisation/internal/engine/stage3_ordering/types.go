package stage3_ordering

import "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"

// LayerNode вЂ” СѓР·РµР» РІ РґРІСѓСЃРІСЏР·РЅРѕРј СЃРїРёСЃРєРµ СЃР»РѕСЏ
// РњРѕР¶РµС‚ СЃРѕРґРµСЂР¶Р°С‚СЊ 0, 1 РёР»Рё 2 Р»СЋРґРµР№:
// - 0 Р»СЋРґРµР№ = РїСЃРµРІРґРѕРІРµСЂС€РёРЅР°
// - 1 С‡РµР»РѕРІРµРє = РѕР±С‹С‡РЅР°СЏ РІРµСЂС€РёРЅР°
// - 2 С‡РµР»РѕРІРµРєР° = РїР°СЂС‚РЅС‘СЂС‹ РёР»Рё СЂРѕРґРёС‚РµР»Рё СЂРµР±С‘РЅРєР°
type LayerNode struct {
	// РЎРїРёСЃРѕРє Р»СЋРґРµР№ РІ СЌС‚РѕР№ РІРµСЂС€РёРЅРµ (РјРѕР¶РµС‚ Р±С‹С‚СЊ РїСѓСЃС‚С‹Рј РґР»СЏ РїСЃРµРІРґРѕРІРµСЂС€РёРЅ)
	People []*stage1_input.Person

	// IsPseudo вЂ” true РµСЃР»Рё СЌС‚Рѕ РїСЃРµРІРґРѕРІРµСЂС€РёРЅР° (People РїСѓСЃС‚РѕР№)
	IsPseudo bool

	// Р“РѕСЂРёР·РѕРЅС‚Р°Р»СЊРЅС‹Рµ СЃРІСЏР·Рё (РґРІСѓСЃРІСЏР·РЅС‹Р№ СЃРїРёСЃРѕРє РЅР° СЃР»РѕРµ)
	Prev *LayerNode
	Next *LayerNode

	// Р’РµСЂС‚РёРєР°Р»СЊРЅС‹Рµ СЃРІСЏР·Рё РІРІРµСЂС… вЂ” РґР»СЏ РєР°Р¶РґРѕРіРѕ С‡РµР»РѕРІРµРєР° СЃРІРѕР№ СѓРєР°Р·Р°С‚РµР»СЊ
	// Up[0] = Р»РµРІС‹Р№ СЂРѕРґРёС‚РµР»СЊ (Р±С‹РІС€РёР№ LeftUp)
	// Up[len-1] = РїСЂР°РІС‹Р№ СЂРѕРґРёС‚РµР»СЊ (Р±С‹РІС€РёР№ RightUp)
	// Р”Р»СЏ РїСЃРµРІРґРѕРІРµСЂС€РёРЅ вЂ” РѕРґРёРЅ СѓРєР°Р·Р°С‚РµР»СЊ
	Up []*LayerNode

	// Р’РµСЂС‚РёРєР°Р»СЊРЅС‹Рµ СЃРІСЏР·Рё РІРЅРёР· (РЅР° РѕР±С‰РёС… РґРµС‚РµР№ РїР°СЂС‹)
	LeftDown  *LayerNode // СЃР°РјС‹Р№ Р»РµРІС‹Р№ СЂРµР±С‘РЅРѕРє
	RightDown *LayerNode // СЃР°РјС‹Р№ РїСЂР°РІС‹Р№ СЂРµР±С‘РЅРѕРє

	// РЎР»РѕР№, РЅР° РєРѕС‚РѕСЂРѕРј РЅР°С…РѕРґРёС‚СЃСЏ СѓР·РµР»
	Layer int

	// AddedLeft вЂ” true РµСЃР»Рё РІРµСЂС€РёРЅР° Р±С‹Р»Р° РґРѕР±Р°РІР»РµРЅР° СЃР»РµРІР°
	AddedLeft bool
}

// IsPerson РїСЂРѕРІРµСЂСЏРµС‚, СЏРІР»СЏРµС‚СЃСЏ Р»Рё СѓР·РµР» СЂРµР°Р»СЊРЅС‹Рј С‡РµР»РѕРІРµРєРѕРј (РЅРµ РїСЃРµРІРґРѕ)
func (n *LayerNode) IsPerson() bool {
	return len(n.People) > 0 && !n.IsPseudo
}

// HasPerson РїСЂРѕРІРµСЂСЏРµС‚, СЃРѕРґРµСЂР¶РёС‚ Р»Рё СѓР·РµР» РєРѕРЅРєСЂРµС‚РЅРѕРіРѕ С‡РµР»РѕРІРµРєР°
func (n *LayerNode) HasPerson(personID int) bool {
	for _, p := range n.People {
		if p.ID == personID {
			return true
		}
	}
	return false
}

// GetLeftUp РІРѕР·РІСЂР°С‰Р°РµС‚ Р»РµРІС‹Р№ СѓРєР°Р·Р°С‚РµР»СЊ РІРІРµСЂС… (РїРµСЂРІС‹Р№ СЌР»РµРјРµРЅС‚ Up)
// Р’РѕР·РІСЂР°С‰Р°РµС‚ РїРµСЂРІС‹Р№ РЅРµ-nil СЌР»РµРјРµРЅС‚, РЅР°С‡РёРЅР°СЏ СЃ Р»РµРІРѕРіРѕ РєСЂР°СЏ
func (n *LayerNode) GetLeftUp() *LayerNode {
	for i := 0; i < len(n.Up); i++ {
		if n.Up[i] != nil {
			return n.Up[i]
		}
	}
	return nil
}

// GetRightUp РІРѕР·РІСЂР°С‰Р°РµС‚ РїСЂР°РІС‹Р№ СѓРєР°Р·Р°С‚РµР»СЊ РІРІРµСЂС… (РїРѕСЃР»РµРґРЅРёР№ СЌР»РµРјРµРЅС‚ Up)
// Р’РѕР·РІСЂР°С‰Р°РµС‚ РїРµСЂРІС‹Р№ РЅРµ-nil СЌР»РµРјРµРЅС‚, РЅР°С‡РёРЅР°СЏ СЃ РїСЂР°РІРѕРіРѕ РєСЂР°СЏ
func (n *LayerNode) GetRightUp() *LayerNode {
	for i := len(n.Up) - 1; i >= 0; i-- {
		if n.Up[i] != nil {
			return n.Up[i]
		}
	}
	return nil
}

// AddPersonToNode РґРѕР±Р°РІР»СЏРµС‚ С‡РµР»РѕРІРµРєР° РІ РІРµСЂС€РёРЅСѓ
// position: "left" вЂ” РІ РЅР°С‡Р°Р»Рѕ, "right" вЂ” РІ РєРѕРЅРµС†
func (n *LayerNode) AddPersonToNode(person *stage1_input.Person, position string) {
	if position == "left" {
		n.People = append([]*stage1_input.Person{person}, n.People...)
		// РџСЂРё РґРѕР±Р°РІР»РµРЅРёРё РІ РЅР°С‡Р°Р»Рѕ РЅСѓР¶РЅРѕ СЃРґРІРёРЅСѓС‚СЊ РёРЅРґРµРєСЃС‹ РІ Up
		// Р’СЃС‚Р°РІР»СЏРµРј nil РІ РЅР°С‡Р°Р»Рѕ Up, С‡С‚РѕР±С‹ СЃРѕС…СЂР°РЅРёС‚СЊ СЃРѕРѕС‚РІРµС‚СЃС‚РІРёРµ Up[i] -> People[i]
		n.Up = append([]*LayerNode{nil}, n.Up...)
	} else {
		n.People = append(n.People, person)
	}
}

// IsHead РїСЂРѕРІРµСЂСЏРµС‚, СЏРІР»СЏРµС‚СЃСЏ Р»Рё СѓР·РµР» Р»РµРІС‹Рј РєСЂР°РµРј СЃР»РѕСЏ
func (n *LayerNode) IsHead() bool {
	return n.Prev == nil
}

// IsTail РїСЂРѕРІРµСЂСЏРµС‚, СЏРІР»СЏРµС‚СЃСЏ Р»Рё СѓР·РµР» РїСЂР°РІС‹Рј РєСЂР°РµРј СЃР»РѕСЏ
func (n *LayerNode) IsTail() bool {
	return n.Next == nil
}

// Layer РїСЂРµРґСЃС‚Р°РІР»СЏРµС‚ РѕРґРёРЅ СЃР»РѕР№ (РґРІСѓСЃРІСЏР·РЅС‹Р№ СЃРїРёСЃРѕРє)
type Layer struct {
	Number int
	Head   *LayerNode // Р›РµРІС‹Р№ РєСЂР°Р№ (С„РёРєС‚РёРІРЅС‹Р№ СѓР·РµР»)
	Tail   *LayerNode // РџСЂР°РІС‹Р№ РєСЂР°Р№ (С„РёРєС‚РёРІРЅС‹Р№ СѓР·РµР»)
}

// GetNodes РІРѕР·РІСЂР°С‰Р°РµС‚ РІСЃРµ СѓР·Р»С‹ СЃР»РѕСЏ (РІРєР»СЋС‡Р°СЏ РїСЃРµРІРґРѕРІРµСЂС€РёРЅС‹, РёСЃРєР»СЋС‡Р°СЏ Head/Tail)
func (l *Layer) GetNodes() []*LayerNode {
	var nodes []*LayerNode
	for node := l.Head.Next; node != nil && node != l.Tail; node = node.Next {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetPeople РІРѕР·РІСЂР°С‰Р°РµС‚ СЃРїРёСЃРѕРє Р»СЋРґРµР№ РЅР° СЃР»РѕРµ РІ РїРѕСЂСЏРґРєРµ СЃР»РµРІР° РЅР°РїСЂР°РІРѕ
func (l *Layer) GetPeople() []*stage1_input.Person {
	var people []*stage1_input.Person
	for node := l.Head.Next; node != nil && node != l.Tail; node = node.Next {
		for _, p := range node.People {
			people = append(people, p)
		}
	}
	return people
}

// GetPeopleIDs РІРѕР·РІСЂР°С‰Р°РµС‚ СЃРїРёСЃРѕРє ID Р»СЋРґРµР№ РЅР° СЃР»РѕРµ РІ РїРѕСЂСЏРґРєРµ СЃР»РµРІР° РЅР°РїСЂР°РІРѕ
func (l *Layer) GetPeopleIDs() []int {
	var ids []int
	for node := l.Head.Next; node != nil && node != l.Tail; node = node.Next {
		for _, p := range node.People {
			ids = append(ids, p.ID)
		}
	}
	return ids
}

// IsEmpty РїСЂРѕРІРµСЂСЏРµС‚, РїСѓСЃС‚ Р»Рё СЃР»РѕР№ (С‚РѕР»СЊРєРѕ Head Рё Tail)
func (l *Layer) IsEmpty() bool {
	return l.Head.Next == l.Tail
}
