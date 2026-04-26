package stage1_input

import (
	"encoding/json"
	"fmt"
	"os"
)

type InputData struct {
	People        []PersonData `json:"people"`
	Families      []FamilyData `json:"families"`
	StartPersonID int          `json:"start_person_id"`
}

type PersonData struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Gender string `json:"gender"`
}

type FamilyData struct {
	HusbandID int   `json:"husband_id"`
	WifeID    int   `json:"wife_id"`
	Children  []int `json:"children"`
}

func LoadFromFile(filename string) (*FamilyTree, int, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read file: %w", err)
	}

	return LoadFromJSON(data)
}

func LoadFromJSON(data []byte) (*FamilyTree, int, error) {
	var input InputData
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, 0, fmt.Errorf("failed to parse JSON: %w", err)
	}

	tree := NewFamilyTree()

	for _, p := range input.People {
		gender := Male
		if p.Gender == "female" {
			gender = Female
		}
		person := NewPerson(p.ID, p.Name, gender)
		tree.AddPerson(person)
	}

	for _, family := range input.Families {

		if err := tree.AddPartnership(family.HusbandID, family.WifeID); err != nil {
			return nil, 0, fmt.Errorf("failed to add partnership: %w", err)
		}

		for _, childID := range family.Children {
			if err := tree.SetParents(childID, family.WifeID, family.HusbandID); err != nil {
				return nil, 0, fmt.Errorf("failed to set parents: %w", err)
			}
		}
	}

	return tree, input.StartPersonID, nil
}

func CreateSampleData() (*FamilyTree, int) {
	tree := NewFamilyTree()

	ivan := NewPerson(1, "РРІР°РЅ", Male)
	maria := NewPerson(2, "РњР°СЂРёСЏ", Female)
	petr := NewPerson(3, "РџС‘С‚СЂ", Male)
	anna := NewPerson(4, "РђРЅРЅР°", Female)
	sergey := NewPerson(5, "РЎРµСЂРіРµР№", Male)

	tree.AddPerson(ivan)
	tree.AddPerson(maria)
	tree.AddPerson(petr)
	tree.AddPerson(anna)
	tree.AddPerson(sergey)

	tree.AddPartnership(1, 2)
	tree.AddPartnership(1, 4)

	tree.SetParents(3, 2, 1)

	tree.SetParents(5, 4, 1)

	return tree, 3
}
