package stage1_input

import "fmt"

// FamilyTree РїСЂРµРґСЃС‚Р°РІР»СЏРµС‚ РІСЃС‘ СЂРѕРґРѕСЃР»РѕРІРЅРѕРµ РґРµСЂРµРІРѕ
type FamilyTree struct {
	People map[int]*Person // Р’СЃРµ Р»СЋРґРё РїРѕ ID
}

// NewFamilyTree СЃРѕР·РґР°С‘С‚ РЅРѕРІРѕРµ РїСѓСЃС‚РѕРµ РґРµСЂРµРІРѕ
func NewFamilyTree() *FamilyTree {
	return &FamilyTree{
		People: make(map[int]*Person),
	}
}

// AddPerson РґРѕР±Р°РІР»СЏРµС‚ С‡РµР»РѕРІРµРєР° РІ РґРµСЂРµРІРѕ
func (ft *FamilyTree) AddPerson(person *Person) {
	ft.People[person.ID] = person
}

// GetPerson РІРѕР·РІСЂР°С‰Р°РµС‚ С‡РµР»РѕРІРµРєР° РїРѕ ID
func (ft *FamilyTree) GetPerson(id int) *Person {
	return ft.People[id]
}

// SetParents СѓСЃС‚Р°РЅР°РІР»РёРІР°РµС‚ СЂРѕРґРёС‚РµР»РµР№ РґР»СЏ СЂРµР±С‘РЅРєР°
func (ft *FamilyTree) SetParents(childID, motherID, fatherID int) error {
	child := ft.People[childID]
	mother := ft.People[motherID]
	father := ft.People[fatherID]

	if child == nil {
		return fmt.Errorf("child with ID %d not found", childID)
	}
	if mother == nil {
		return fmt.Errorf("mother with ID %d not found", motherID)
	}
	if father == nil {
		return fmt.Errorf("father with ID %d not found", fatherID)
	}

	child.Mother = mother
	child.Father = father

	// Р”РѕР±Р°РІР»СЏРµРј СЂРµР±С‘РЅРєР° РІ СЃРїРёСЃРєРё РґРµС‚РµР№ СЂРѕРґРёС‚РµР»РµР№, РµСЃР»Рё РµС‰С‘ РЅРµС‚
	if !containsPerson(mother.Children, child) {
		mother.Children = append(mother.Children, child)
	}
	if !containsPerson(father.Children, child) {
		father.Children = append(father.Children, child)
	}

	return nil
}

// AddPartnership РґРѕР±Р°РІР»СЏРµС‚ РїР°СЂС‚РЅС‘СЂСЃС‚РІРѕ РјРµР¶РґСѓ РґРІСѓРјСЏ Р»СЋРґСЊРјРё
func (ft *FamilyTree) AddPartnership(person1ID, person2ID int) error {
	person1 := ft.People[person1ID]
	person2 := ft.People[person2ID]

	if person1 == nil {
		return fmt.Errorf("person with ID %d not found", person1ID)
	}
	if person2 == nil {
		return fmt.Errorf("person with ID %d not found", person2ID)
	}

	// Р”РѕР±Р°РІР»СЏРµРј РґСЂСѓРі РґСЂСѓРіР° РІ СЃРїРёСЃРєРё РїР°СЂС‚РЅС‘СЂРѕРІ, РµСЃР»Рё РµС‰С‘ РЅРµС‚
	if !containsPerson(person1.Partners, person2) {
		person1.Partners = append(person1.Partners, person2)
	}
	if !containsPerson(person2.Partners, person1) {
		person2.Partners = append(person2.Partners, person1)
	}

	return nil
}

// PrintResults РІС‹РІРѕРґРёС‚ СЂРµР·СѓР»СЊС‚Р°С‚С‹ СЂР°Р·РјРµС‰РµРЅРёСЏ
func (ft *FamilyTree) PrintResults() {
	for _, person := range ft.People {
		if person.Layout != nil {
			fmt.Printf("%s (ID=%d): Layer %d\n", person.Name, person.ID, person.Layout.Layer)
		} else {
			fmt.Printf("%s (ID=%d): РЅРµ СЂР°Р·РјРµС‰С‘РЅ\n", person.Name, person.ID)
		}
	}
}

// containsPerson РїСЂРѕРІРµСЂСЏРµС‚, СЃРѕРґРµСЂР¶РёС‚СЃСЏ Р»Рё С‡РµР»РѕРІРµРє РІ СЃСЂРµР·Рµ
func containsPerson(slice []*Person, person *Person) bool {
	for _, p := range slice {
		if p.ID == person.ID {
			return true
		}
	}
	return false
}
