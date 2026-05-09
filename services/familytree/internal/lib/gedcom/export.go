package gedcom

import (
	"fmt"
	"strings"
	"time"

	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
)

// ExportToGEDCOM converts persons and relationships to GEDCOM format text
func ExportToGEDCOM(persons []models.Person, relationships []models.Relationship) string {
	// Build internal data structures
	data := &GEDCOMData{
		Persons:  make(map[string]*GEDCOMPerson),
		Families: make(map[string]*GEDCOMFamily),
	}

	// Create persons and build ID mapping
	uuidToGEDCOMID := make(map[string]string)
	for i, person := range persons {
		gedcomID := fmt.Sprintf("I%d", i+1)
		uuidToGEDCOMID[person.ID.String()] = gedcomID

		gender := "U" // Unknown
		if person.Gender == models.GenderMale {
			gender = "M"
		} else if person.Gender == models.GenderFemale {
			gender = "F"
		}

		data.Persons[gedcomID] = &GEDCOMPerson{
			ID:         gedcomID,
			FirstName:  person.FirstName,
			LastName:   person.LastName,
			Patronymic: person.Patronymic,
			Gender:     gender,
		}
		data.NextPersonID = i + 2
	}

	// Process relationships
	familyMap := make(map[string]*GEDCOMFamily) // Key: "husband_id:wife_id"
	familyCounter := 0

	for _, rel := range relationships {
		fromID := uuidToGEDCOMID[rel.PersonIDFrom.String()]
		toID := uuidToGEDCOMID[rel.PersonIDTo.String()]

		switch rel.Type {
		case models.RelationshipParentChild:
			// Add child to parent's family
			// We need to find or create the family for this parent
			// First, find all spouses of this parent
			var familyKey string
			parent := data.Persons[fromID]

			if parent.Gender == "M" {
				// Find spouse
				spouseID := ""
				for _, frel := range relationships {
					if frel.PersonIDFrom.String() == rel.PersonIDFrom.String() &&
						frel.Type == models.RelationshipPartner ||
						frel.Type == models.RelationshipPartnerMarried ||
						frel.Type == models.RelationshipPartnerDivorced ||
						frel.Type == models.RelationshipPartnerUnmarried {
						spouseID = uuidToGEDCOMID[frel.PersonIDTo.String()]
						break
					} else if frel.PersonIDTo.String() == rel.PersonIDFrom.String() &&
						(frel.Type == models.RelationshipPartner ||
							frel.Type == models.RelationshipPartnerMarried ||
							frel.Type == models.RelationshipPartnerDivorced ||
							frel.Type == models.RelationshipPartnerUnmarried) {
						spouseID = uuidToGEDCOMID[frel.PersonIDFrom.String()]
						break
					}
				}
				familyKey = fromID + ":" + spouseID
			} else {
				// Find spouse
				spouseID := ""
				for _, frel := range relationships {
					if frel.PersonIDFrom.String() == rel.PersonIDFrom.String() &&
						(frel.Type == models.RelationshipPartner ||
							frel.Type == models.RelationshipPartnerMarried ||
							frel.Type == models.RelationshipPartnerDivorced ||
							frel.Type == models.RelationshipPartnerUnmarried) {
						spouseID = uuidToGEDCOMID[frel.PersonIDTo.String()]
						break
					} else if frel.PersonIDTo.String() == rel.PersonIDFrom.String() &&
						(frel.Type == models.RelationshipPartner ||
							frel.Type == models.RelationshipPartnerMarried ||
							frel.Type == models.RelationshipPartnerDivorced ||
							frel.Type == models.RelationshipPartnerUnmarried) {
						spouseID = uuidToGEDCOMID[frel.PersonIDFrom.String()]
						break
					}
				}
				familyKey = spouseID + ":" + fromID
			}

			// Create or get family
			if _, exists := familyMap[familyKey]; !exists {
				familyCounter++
				familyID := fmt.Sprintf("F%d", familyCounter)
				parts := strings.Split(familyKey, ":")
				family := &GEDCOMFamily{
					ID:       familyID,
					Husband:  "",
					Wife:     "",
					Children: []string{},
				}
				if len(parts) > 0 && parts[0] != "" {
					family.Husband = parts[0]
				}
				if len(parts) > 1 && parts[1] != "" {
					family.Wife = parts[1]
				}
				familyMap[familyKey] = family
			}

			// Add child to family
			family := familyMap[familyKey]
			family.Children = append(family.Children, toID)
			data.Persons[toID].FamilyAsChild = family.ID

		case models.RelationshipPartner, models.RelationshipPartnerMarried,
			models.RelationshipPartnerDivorced, models.RelationshipPartnerUnmarried:
			// Add spouse link
			data.Persons[fromID].FamilySpouses = append(data.Persons[fromID].FamilySpouses, toID)
		}
	}

	// Add families to data
	for _, family := range familyMap {
		data.Families[family.ID] = family
	}
	data.NextFamilyID = familyCounter + 1

	return renderGEDCOM(data)
}

// renderGEDCOM renders GEDCOMData as GEDCOM text format
func renderGEDCOM(data *GEDCOMData) string {
	var sb strings.Builder

	// Write header
	sb.WriteString("0 HEAD\n")
	sb.WriteString("1 SOUR MyForebears\n")
	sb.WriteString("1 VERS 5.5.5\n")
	sb.WriteString("1 GEDC\n")
	sb.WriteString("2 VERS 5.5.5\n")
	sb.WriteString("2 FORM LINEAGE-LINKED\n")
	sb.WriteString("1 CHAR UTF-8\n")
	sb.WriteString(fmt.Sprintf("1 DATE %s\n", time.Now().Format("2 JAN 2006")))
	sb.WriteString("1 FILE myforebears_export.ged\n")
	sb.WriteString("1 LANG Russian\n")

	// Write individuals
	for _, id := range getSortedKeys(data.Persons) {
		person := data.Persons[id]
		sb.WriteString(fmt.Sprintf("0 @%s@ INDI\n", person.ID))

		// Name: format as "FirstName LastName /Patronymic/"
		name := fmt.Sprintf("%s %s", person.FirstName, person.LastName)
		sb.WriteString(fmt.Sprintf("1 NAME %s /%s/\n", name, person.Patronymic))

		// Given name (GIVN)
		sb.WriteString(fmt.Sprintf("2 GIVN %s\n", person.FirstName))

		// Surname (SURN)
		sb.WriteString(fmt.Sprintf("2 SURN %s\n", person.LastName))

		// Patronymic as custom tag (PATR)
		if person.Patronymic != "" {
			sb.WriteString(fmt.Sprintf("2 PATR %s\n", person.Patronymic))
		}

		// Gender (SEX)
		sb.WriteString(fmt.Sprintf("1 SEX %s\n", person.Gender))

		// Family as child (FAMC)
		if person.FamilyAsChild != "" {
			sb.WriteString(fmt.Sprintf("1 FAMC @%s@\n", person.FamilyAsChild))
		}

		// Family as spouse (FAMS)
		for _, familyID := range person.FamilySpouses {
			sb.WriteString(fmt.Sprintf("1 FAMS @%s@\n", familyID))
		}
	}

	// Write families
	for _, id := range getSortedKeys(data.Families) {
		family := data.Families[id]
		sb.WriteString(fmt.Sprintf("0 @%s@ FAM\n", family.ID))

		// Husband (HUSB)
		if family.Husband != "" {
			sb.WriteString(fmt.Sprintf("1 HUSB @%s@\n", family.Husband))
		}

		// Wife (WIFE)
		if family.Wife != "" {
			sb.WriteString(fmt.Sprintf("1 WIFE @%s@\n", family.Wife))
		}

		// Children (CHIL)
		for _, childID := range family.Children {
			sb.WriteString(fmt.Sprintf("1 CHIL @%s@\n", childID))
		}
	}

	// Write trailer
	sb.WriteString("0 TRLR\n")

	return sb.String()
}

// getSortedKeys returns sorted keys for consistent output
func getSortedKeys(m interface{}) []string {
	var keys []string
	switch v := m.(type) {
	case map[string]*GEDCOMPerson:
		for k := range v {
			keys = append(keys, k)
		}
	case map[string]*GEDCOMFamily:
		for k := range v {
			keys = append(keys, k)
		}
	}

	// Simple string sort for order like I1, I2, ... F1, F2, ...
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}
