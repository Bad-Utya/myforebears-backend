package stage3_ordering

import (
	"fmt"

	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"
)

// OrderManager СѓРїСЂР°РІР»СЏРµС‚ СѓРїРѕСЂСЏРґРѕС‡РёРІР°РЅРёРµРј РІРµСЂС€РёРЅ РІ СЃР»РѕСЏС…
type OrderManager struct {
	// РЎР»РѕРё (РЅРѕРјРµСЂ СЃР»РѕСЏ -> Layer)
	Layers map[int]*Layer

	// РЈР·Р»С‹ РґР»СЏ РєР°Р¶РґРѕРіРѕ С‡РµР»РѕРІРµРєР° (ID -> LayerNode)
	PersonNodes map[int]*LayerNode

	// Р”РёР°РїР°Р·РѕРЅ СЃР»РѕС‘РІ
	minLayer int
	maxLayer int
}

// NewOrderManager СЃРѕР·РґР°С‘С‚ РЅРѕРІС‹Р№ OrderManager
func NewOrderManager(minLayer, maxLayer int) *OrderManager {
	om := &OrderManager{
		Layers:      make(map[int]*Layer),
		PersonNodes: make(map[int]*LayerNode),
		minLayer:    minLayer,
		maxLayer:    maxLayer,
	}

	// РРЅРёС†РёР°Р»РёР·РёСЂСѓРµРј СЃР»РѕРё
	for layer := minLayer; layer <= maxLayer; layer++ {
		om.initLayer(layer)
	}

	return om
}

// initLayer РёРЅРёС†РёР°Р»РёР·РёСЂСѓРµС‚ СЃР»РѕР№ СЃ С„РёРєС‚РёРІРЅС‹РјРё РіСЂР°РЅРёС‡РЅС‹РјРё СѓР·Р»Р°РјРё
func (om *OrderManager) initLayer(layerNum int) {
	// РЎРѕР·РґР°С‘Рј С„РёРєС‚РёРІРЅС‹Рµ СѓР·Р»С‹ Head Рё Tail
	head := &LayerNode{Layer: layerNum}
	tail := &LayerNode{Layer: layerNum}

	// РЎРІСЏР·С‹РІР°РµРј РёС…
	head.Next = tail
	tail.Prev = head

	// РЎРѕР·РґР°С‘Рј СЃР»РѕР№
	om.Layers[layerNum] = &Layer{
		Number: layerNum,
		Head:   head,
		Tail:   tail,
	}
}

// ensureLayer РіР°СЂР°РЅС‚РёСЂСѓРµС‚, С‡С‚Рѕ СЃР»РѕР№ СЃСѓС‰РµСЃС‚РІСѓРµС‚
func (om *OrderManager) ensureLayer(layerNum int) {
	if _, exists := om.Layers[layerNum]; !exists {
		om.initLayer(layerNum)
		if layerNum < om.minLayer {
			om.minLayer = layerNum
		}
		if layerNum > om.maxLayer {
			om.maxLayer = layerNum
		}
	}
}

// insertAfter РІСЃС‚Р°РІР»СЏРµС‚ СѓР·РµР» node РїРѕСЃР»Рµ СѓР·Р»Р° after
func (om *OrderManager) insertAfter(after, node *LayerNode) {
	node.Prev = after
	node.Next = after.Next
	if after.Next != nil {
		after.Next.Prev = node
	}
	after.Next = node
}

// insertBefore РІСЃС‚Р°РІР»СЏРµС‚ СѓР·РµР» node РїРµСЂРµРґ СѓР·Р»РѕРј before
func (om *OrderManager) insertBefore(before, node *LayerNode) {
	node.Next = before
	node.Prev = before.Prev
	if before.Prev != nil {
		before.Prev.Next = node
	}
	before.Prev = node
}

// CreatePseudoNode СЃРѕР·РґР°С‘С‚ РїСЃРµРІРґРѕРІРµСЂС€РёРЅСѓ РЅР° СѓРєР°Р·Р°РЅРЅРѕРј СЃР»РѕРµ
func (om *OrderManager) CreatePseudoNode(layer int) *LayerNode {
	om.ensureLayer(layer)
	return &LayerNode{
		IsPseudo: true,
		Layer:    layer,
		People:   []*stage1_input.Person{}, // РїСѓСЃС‚РѕР№ СЃРїРёСЃРѕРє
		Up:       []*LayerNode{},
	}
}

// CreatePersonNode СЃРѕР·РґР°С‘С‚ СѓР·РµР» РґР»СЏ СЂРµР°Р»СЊРЅРѕРіРѕ С‡РµР»РѕРІРµРєР°
func (om *OrderManager) CreatePersonNode(person *stage1_input.Person, layer int) *LayerNode {
	om.ensureLayer(layer)
	node := &LayerNode{
		People: []*stage1_input.Person{person},
		Layer:  layer,
		Up:     []*LayerNode{},
	}
	om.PersonNodes[person.ID] = node
	return node
}

// CreatePairNode СЃРѕР·РґР°С‘С‚ СѓР·РµР» РґР»СЏ РїР°СЂС‹ Р»СЋРґРµР№ (РїР°СЂС‚РЅС‘СЂС‹ РёР»Рё СЂРѕРґРёС‚РµР»Рё)
func (om *OrderManager) CreatePairNode(person1, person2 *stage1_input.Person, layer int) *LayerNode {
	om.ensureLayer(layer)
	node := &LayerNode{
		People: []*stage1_input.Person{person1, person2},
		Layer:  layer,
		Up:     []*LayerNode{},
	}
	om.PersonNodes[person1.ID] = node
	om.PersonNodes[person2.ID] = node
	return node
}

// AddStartPerson РґРѕР±Р°РІР»СЏРµС‚ РЅР°С‡Р°Р»СЊРЅСѓСЋ РІРµСЂС€РёРЅСѓ
func (om *OrderManager) AddStartPerson(person *stage1_input.Person, layer int) *LayerNode {
	om.ensureLayer(layer)

	// РЎРѕР·РґР°С‘Рј СѓР·РµР» РґР»СЏ С‡РµР»РѕРІРµРєР°
	personNode := om.CreatePersonNode(person, layer)

	// Р’СЃС‚Р°РІР»СЏРµРј РјРµР¶РґСѓ Head Рё Tail РЅР° СЃР»РѕРµ
	layerObj := om.Layers[layer]
	om.insertAfter(layerObj.Head, personNode)

	return personNode
}

// GetLayerOrder РІРѕР·РІСЂР°С‰Р°РµС‚ РїРѕСЂСЏРґРѕРє Р»СЋРґРµР№ РЅР° СЃР»РѕРµ
func (om *OrderManager) GetLayerOrder(layer int) []int {
	if l, exists := om.Layers[layer]; exists {
		return l.GetPeopleIDs()
	}
	return nil
}

// GetAllLayers РІРѕР·РІСЂР°С‰Р°РµС‚ РІСЃРµ СЃР»РѕРё РІ РїРѕСЂСЏРґРєРµ СѓР±С‹РІР°РЅРёСЏ РЅРѕРјРµСЂР°
func (om *OrderManager) GetAllLayers() []*Layer {
	var layers []*Layer
	for l := om.maxLayer; l >= om.minLayer; l-- {
		if layer, exists := om.Layers[l]; exists {
			layers = append(layers, layer)
		}
	}
	return layers
}

// GetPersonNode РІРѕР·РІСЂР°С‰Р°РµС‚ СѓР·РµР» РґР»СЏ С‡РµР»РѕРІРµРєР° РїРѕ ID
func (om *OrderManager) GetPersonNode(personID int) *LayerNode {
	return om.PersonNodes[personID]
}

// AddPersonToExistingNode РґРѕР±Р°РІР»СЏРµС‚ С‡РµР»РѕРІРµРєР° РІ СЃСѓС‰РµСЃС‚РІСѓСЋС‰СѓСЋ РІРµСЂС€РёРЅСѓ
func (om *OrderManager) AddPersonToExistingNode(node *LayerNode, person *stage1_input.Person, position string) {
	node.AddPersonToNode(person, position)
	om.PersonNodes[person.ID] = node
}

// PrintDebugInfo РІС‹РІРѕРґРёС‚ РѕС‚Р»Р°РґРѕС‡РЅСѓСЋ РёРЅС„РѕСЂРјР°С†РёСЋ Рѕ РІСЃРµС… РІРµСЂС€РёРЅР°С…
func (om *OrderManager) PrintDebugInfo() {
	fmt.Println("\n=== РћС‚Р»Р°РґРѕС‡РЅР°СЏ РёРЅС„РѕСЂРјР°С†РёСЏ Рѕ РІРµСЂС€РёРЅР°С… ===")
	for _, layer := range om.GetAllLayers() {
		fmt.Printf("\nРЎР»РѕР№ %d:\n", layer.Number)
		nodeIndex := 0
		for node := layer.Head.Next; node != nil && node != layer.Tail; node = node.Next {
			fmt.Printf("  Р’РµСЂС€РёРЅР° %d: ", nodeIndex)

			// Р›СЋРґРё РІ РІРµСЂС€РёРЅРµ
			if node.IsPseudo {
				fmt.Print("[РїСЃРµРІРґРѕ]")
			} else if len(node.People) == 0 {
				fmt.Print("[РїСѓСЃС‚Рѕ]")
			} else {
				names := make([]string, len(node.People))
				for i, p := range node.People {
					names[i] = fmt.Sprintf("%s(%d)", p.Name, p.ID)
				}
				fmt.Printf("%v", names)
			}

			// Up СѓРєР°Р·Р°С‚РµР»Рё
			fmt.Print(" | Up: [")
			for i, upNode := range node.Up {
				if i > 0 {
					fmt.Print(", ")
				}
				if upNode == nil {
					fmt.Print("nil")
				} else if upNode.IsPseudo {
					fmt.Print("РїСЃРµРІРґРѕ")
				} else if len(upNode.People) > 0 {
					upNames := make([]string, len(upNode.People))
					for j, p := range upNode.People {
						upNames[j] = fmt.Sprintf("%s(%d)", p.Name, p.ID)
					}
					fmt.Printf("%v", upNames)
				}
			}
			fmt.Print("]")

			// LeftDown
			fmt.Print(" | LeftDown: ")
			if node.LeftDown == nil {
				fmt.Print("nil")
			} else if node.LeftDown.IsPseudo {
				fmt.Print("РїСЃРµРІРґРѕ")
			} else if len(node.LeftDown.People) > 0 {
				downNames := make([]string, len(node.LeftDown.People))
				for j, p := range node.LeftDown.People {
					downNames[j] = fmt.Sprintf("%s(%d)", p.Name, p.ID)
				}
				fmt.Printf("%v", downNames)
			}

			// RightDown
			fmt.Print(" | RightDown: ")
			if node.RightDown == nil {
				fmt.Print("nil")
			} else if node.RightDown.IsPseudo {
				fmt.Print("РїСЃРµРІРґРѕ")
			} else if len(node.RightDown.People) > 0 {
				downNames := make([]string, len(node.RightDown.People))
				for j, p := range node.RightDown.People {
					downNames[j] = fmt.Sprintf("%s(%d)", p.Name, p.ID)
				}
				fmt.Printf("%v", downNames)
			}

			fmt.Println()
			nodeIndex++
		}
	}
}
