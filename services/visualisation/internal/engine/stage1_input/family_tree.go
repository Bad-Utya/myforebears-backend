package stage1_input

import "fmt"

type FamilyTree struct {
	People map[int]*Person
}

func NewFamilyTree() *FamilyTree {
	return &FamilyTree{
		People: make(map[int]*Person),
	}
}

func (ft *FamilyTree) AddPerson(person *Person) {
	ft.People[person.ID] = person
}

func (ft *FamilyTree) GetPerson(id int) *Person {
	return ft.People[id]
}

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

	if !containsPerson(mother.Children, child) {
		mother.Children = append(mother.Children, child)
	}
	if !containsPerson(father.Children, child) {
		father.Children = append(father.Children, child)
	}

	return nil
}

func (ft *FamilyTree) AddPartnership(person1ID, person2ID int) error {
	person1 := ft.People[person1ID]
	person2 := ft.People[person2ID]

	if person1 == nil {
		return fmt.Errorf("person with ID %d not found", person1ID)
	}
	if person2 == nil {
		return fmt.Errorf("person with ID %d not found", person2ID)
	}

	if !containsPerson(person1.Partners, person2) {
		person1.Partners = append(person1.Partners, person2)
	}
	if !containsPerson(person2.Partners, person1) {
		person2.Partners = append(person2.Partners, person1)
	}

	return nil
}

func (ft *FamilyTree) PrintResults() {
	for _, person := range ft.People {
		if person.Layout != nil {
			fmt.Printf("%s (ID=%d): Layer %d\n", person.Name, person.ID, person.Layout.Layer)
		} else {
			fmt.Printf("%s (ID=%d): РЅРµ СЂР°Р·РјРµС‰С‘РЅ\n", person.Name, person.ID)
		}
	}
}

func containsPerson(slice []*Person, person *Person) bool {
	for _, p := range slice {
		if p.ID == person.ID {
			return true
		}
	}
	return false
}
