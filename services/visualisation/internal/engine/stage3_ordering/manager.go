package stage3_ordering

import (
	"fmt"

	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine/stage1_input"
)

type OrderManager struct {
	Layers map[int]*Layer

	PersonNodes map[int]*LayerNode

	minLayer int
	maxLayer int
}

func NewOrderManager(minLayer, maxLayer int) *OrderManager {
	om := &OrderManager{
		Layers:      make(map[int]*Layer),
		PersonNodes: make(map[int]*LayerNode),
		minLayer:    minLayer,
		maxLayer:    maxLayer,
	}

	for layer := minLayer; layer <= maxLayer; layer++ {
		om.initLayer(layer)
	}

	return om
}

func (om *OrderManager) initLayer(layerNum int) {

	head := &LayerNode{Layer: layerNum}
	tail := &LayerNode{Layer: layerNum}

	head.Next = tail
	tail.Prev = head

	om.Layers[layerNum] = &Layer{
		Number: layerNum,
		Head:   head,
		Tail:   tail,
	}
}

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

func (om *OrderManager) insertAfter(after, node *LayerNode) {
	node.Prev = after
	node.Next = after.Next
	if after.Next != nil {
		after.Next.Prev = node
	}
	after.Next = node
}

func (om *OrderManager) insertBefore(before, node *LayerNode) {
	node.Next = before
	node.Prev = before.Prev
	if before.Prev != nil {
		before.Prev.Next = node
	}
	before.Prev = node
}

func (om *OrderManager) CreatePseudoNode(layer int) *LayerNode {
	om.ensureLayer(layer)
	return &LayerNode{
		IsPseudo: true,
		Layer:    layer,
		People:   []*stage1_input.Person{},
		Up:       []*LayerNode{},
	}
}

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

func (om *OrderManager) AddStartPerson(person *stage1_input.Person, layer int) *LayerNode {
	om.ensureLayer(layer)

	personNode := om.CreatePersonNode(person, layer)

	layerObj := om.Layers[layer]
	om.insertAfter(layerObj.Head, personNode)

	return personNode
}

func (om *OrderManager) GetLayerOrder(layer int) []int {
	if l, exists := om.Layers[layer]; exists {
		return l.GetPeopleIDs()
	}
	return nil
}

func (om *OrderManager) GetAllLayers() []*Layer {
	var layers []*Layer
	for l := om.maxLayer; l >= om.minLayer; l-- {
		if layer, exists := om.Layers[l]; exists {
			layers = append(layers, layer)
		}
	}
	return layers
}

func (om *OrderManager) GetPersonNode(personID int) *LayerNode {
	return om.PersonNodes[personID]
}

func (om *OrderManager) AddPersonToExistingNode(node *LayerNode, person *stage1_input.Person, position string) {
	node.AddPersonToNode(person, position)
	om.PersonNodes[person.ID] = node
}

func (om *OrderManager) PrintDebugInfo() {
	fmt.Println("\n=== РћС‚Р»Р°РґРѕС‡РЅР°СЏ РёРЅС„РѕСЂРјР°С†РёСЏ Рѕ РІРµСЂС€РёРЅР°С… ===")
	for _, layer := range om.GetAllLayers() {
		fmt.Printf("\nРЎР»РѕР№ %d:\n", layer.Number)
		nodeIndex := 0
		for node := layer.Head.Next; node != nil && node != layer.Tail; node = node.Next {
			fmt.Printf("  Р’РµСЂС€РёРЅР° %d: ", nodeIndex)

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
