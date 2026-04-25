package stage1_input

import (
	"encoding/json"
	"fmt"
	"os"
)

// InputData РїСЂРµРґСЃС‚Р°РІР»СЏРµС‚ С„РѕСЂРјР°С‚ РІС…РѕРґРЅС‹С… РґР°РЅРЅС‹С… JSON
type InputData struct {
	People        []PersonData `json:"people"`
	Families      []FamilyData `json:"families"`
	StartPersonID int          `json:"start_person_id"`
}

// PersonData РїСЂРµРґСЃС‚Р°РІР»СЏРµС‚ РґР°РЅРЅС‹Рµ Рѕ С‡РµР»РѕРІРµРєРµ
type PersonData struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Gender string `json:"gender"` // "male" РёР»Рё "female"
}

// FamilyData РїСЂРµРґСЃС‚Р°РІР»СЏРµС‚ РґР°РЅРЅС‹Рµ Рѕ СЃРµРјСЊРµ
type FamilyData struct {
	HusbandID int   `json:"husband_id"`
	WifeID    int   `json:"wife_id"`
	Children  []int `json:"children"`
}

// LoadFromFile Р·Р°РіСЂСѓР¶Р°РµС‚ РґР°РЅРЅС‹Рµ РёР· JSON С„Р°Р№Р»Р°
func LoadFromFile(filename string) (*FamilyTree, int, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read file: %w", err)
	}

	return LoadFromJSON(data)
}

// LoadFromJSON Р·Р°РіСЂСѓР¶Р°РµС‚ РґР°РЅРЅС‹Рµ РёР· JSON Р±Р°Р№С‚РѕРІ
func LoadFromJSON(data []byte) (*FamilyTree, int, error) {
	var input InputData
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, 0, fmt.Errorf("failed to parse JSON: %w", err)
	}

	tree := NewFamilyTree()

	// Р”РѕР±Р°РІР»СЏРµРј Р»СЋРґРµР№
	for _, p := range input.People {
		gender := Male
		if p.Gender == "female" {
			gender = Female
		}
		person := NewPerson(p.ID, p.Name, gender)
		tree.AddPerson(person)
	}

	// РћР±СЂР°Р±Р°С‚С‹РІР°РµРј СЃРµРјСЊРё
	for _, family := range input.Families {
		// РЈСЃС‚Р°РЅР°РІР»РёРІР°РµРј РїР°СЂС‚РЅС‘СЂСЃС‚РІРѕ РјРµР¶РґСѓ РјСѓР¶РµРј Рё Р¶РµРЅРѕР№
		if err := tree.AddPartnership(family.HusbandID, family.WifeID); err != nil {
			return nil, 0, fmt.Errorf("failed to add partnership: %w", err)
		}

		// РЈСЃС‚Р°РЅР°РІР»РёРІР°РµРј СЂРѕРґРёС‚РµР»РµР№ РґР»СЏ РєР°Р¶РґРѕРіРѕ СЂРµР±С‘РЅРєР°
		for _, childID := range family.Children {
			if err := tree.SetParents(childID, family.WifeID, family.HusbandID); err != nil {
				return nil, 0, fmt.Errorf("failed to set parents: %w", err)
			}
		}
	}

	return tree, input.StartPersonID, nil
}

// CreateSampleData СЃРѕР·РґР°С‘С‚ РїСЂРёРјРµСЂ РґР°РЅРЅС‹С… РґР»СЏ С‚РµСЃС‚РёСЂРѕРІР°РЅРёСЏ
func CreateSampleData() (*FamilyTree, int) {
	tree := NewFamilyTree()

	// РЎРѕР·РґР°С‘Рј Р»СЋРґРµР№
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

	// РџР°СЂС‚РЅС‘СЂСЃС‚РІР°: РРІР°РЅ + РњР°СЂРёСЏ (С‚РµРєСѓС‰Р°СЏ), РРІР°РЅ + РђРЅРЅР° (Р±С‹РІС€Р°СЏ)
	tree.AddPartnership(1, 2)
	tree.AddPartnership(1, 4)

	// Р РѕРґРёС‚РµР»Рё: РџС‘С‚СЂ вЂ” СЃС‹РЅ РРІР°РЅР° Рё РњР°СЂРёРё
	tree.SetParents(3, 2, 1)

	// РЎРµСЂРіРµР№ вЂ” СЃС‹РЅ РРІР°РЅР° Рё РђРЅРЅС‹
	tree.SetParents(5, 4, 1)

	// РќР°С‡РёРЅР°РµРј СЃ РџРµС‚СЂР°
	return tree, 3
}
