package gedcom

import (
	"fmt"
	"strings"

	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/google/uuid"
)

// ParsedTreeData holds parsed GEDCOM data ready for import
type ParsedTreeData struct {
	Persons       []models.Person
	Relationships []models.Relationship
	Errors        []string
}

// ImportFromGEDCOM parses GEDCOM content and returns parsed tree data
func ImportFromGEDCOM(gedcomContent string) *ParsedTreeData {
	result := &ParsedTreeData{
		Persons:       []models.Person{},
		Relationships: []models.Relationship{},
		Errors:        []string{},
	}

	lines := strings.Split(gedcomContent, "\n")

	gedcomPersons := make(map[string]*GEDCOMPerson)  // ID -> Person
	gedcomFamilies := make(map[string]*GEDCOMFamily) // ID -> Family
	currentRecord := ""
	var currentPerson *GEDCOMPerson
	var currentFamily *GEDCOMFamily

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse line: LEVEL TAG VALUE or LEVEL @ID@ TAG VALUE
		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 2 {
			continue
		}

		level, tag, value := parseGEDCOMLine(parts)

		switch level {
		case 0:
			// Save previous record
			if currentRecord == "INDI" && currentPerson != nil {
				gedcomPersons[currentPerson.ID] = currentPerson
			} else if currentRecord == "FAM" && currentFamily != nil {
				gedcomFamilies[currentFamily.ID] = currentFamily
			}

			// Start new record
			if strings.HasPrefix(tag, "@") && strings.HasSuffix(tag, "@") {
				// Extract ID and type
				parts := strings.SplitN(value, " ", 2)
				if len(parts) == 2 {
					id := strings.Trim(tag, "@")
					recordType := parts[0]

					if recordType == "INDI" {
						currentPerson = &GEDCOMPerson{
							ID:            id,
							FamilySpouses: []string{},
						}
						currentFamily = nil
						currentRecord = "INDI"
					} else if recordType == "FAM" {
						currentFamily = &GEDCOMFamily{
							ID:       id,
							Children: []string{},
						}
						currentPerson = nil
						currentRecord = "FAM"
					}
				}
			} else if tag == "TRLR" {
				// Trailer
				currentRecord = ""
			}

		case 1:
			if currentPerson != nil {
				switch tag {
				case "NAME":
					parseName(value, currentPerson)
				case "SEX":
					currentPerson.Gender = strings.TrimSpace(value)
				case "FAMC":
					currentPerson.FamilyAsChild = extractID(value)
				case "FAMS":
					spouseID := extractID(value)
					if spouseID != "" {
						currentPerson.FamilySpouses = append(currentPerson.FamilySpouses, spouseID)
					}
				}
			} else if currentFamily != nil {
				switch tag {
				case "HUSB":
					currentFamily.Husband = extractID(value)
				case "WIFE":
					currentFamily.Wife = extractID(value)
				case "CHIL":
					childID := extractID(value)
					if childID != "" {
						currentFamily.Children = append(currentFamily.Children, childID)
					}
				}
			}

		case 2:
			if currentPerson != nil {
				switch tag {
				case "GIVN":
					currentPerson.FirstName = strings.TrimSpace(value)
				case "SURN":
					currentPerson.LastName = strings.TrimSpace(value)
				case "PATR":
					currentPerson.Patronymic = strings.TrimSpace(value)
				}
			}
		}
	}

	// Save last record
	if currentRecord == "INDI" && currentPerson != nil {
		gedcomPersons[currentPerson.ID] = currentPerson
	} else if currentRecord == "FAM" && currentFamily != nil {
		gedcomFamilies[currentFamily.ID] = currentFamily
	}

	// Convert to models and build relationships
	personIDMap := make(map[string]string) // GEDCOM ID -> UUID

	// Create persons
	treeID := uuid.New()
	for _, gedcomPerson := range gedcomPersons {
		gender := models.GenderMale
		if gedcomPerson.Gender == "F" {
			gender = models.GenderFemale
		}

		person := models.Person{
			ID:         uuid.New(),
			TreeID:     treeID,
			FirstName:  gedcomPerson.FirstName,
			LastName:   gedcomPerson.LastName,
			Patronymic: gedcomPerson.Patronymic,
			Gender:     gender,
		}

		personIDMap[gedcomPerson.ID] = person.ID.String()
		result.Persons = append(result.Persons, person)
	}

	// Create relationships from families
	for _, family := range gedcomFamilies {
		// Add parent-child relationships
		childPersonIDs := family.Children

		husbandID := ""
		if family.Husband != "" {
			if pid, ok := personIDMap[family.Husband]; ok {
				husbandID = pid
			}
		}

		wifeID := ""
		if family.Wife != "" {
			if pid, ok := personIDMap[family.Wife]; ok {
				wifeID = pid
			}
		}

		// Parent -> Child relationships
		for _, childGID := range childPersonIDs {
			if childPID, ok := personIDMap[childGID]; ok {
				if husbandID != "" {
					parent := parseUUID(husbandID)
					child := parseUUID(childPID)
					result.Relationships = append(result.Relationships, models.Relationship{
						PersonIDFrom: parent,
						PersonIDTo:   child,
						Type:         models.RelationshipParentChild,
					})
				}
				if wifeID != "" {
					parent := parseUUID(wifeID)
					child := parseUUID(childPID)
					result.Relationships = append(result.Relationships, models.Relationship{
						PersonIDFrom: parent,
						PersonIDTo:   child,
						Type:         models.RelationshipParentChild,
					})
				}
			}
		}

		// Partner relationship (husband <-> wife)
		if husbandID != "" && wifeID != "" {
			husband := parseUUID(husbandID)
			wife := parseUUID(wifeID)
			// Married by default, no status info in basic GEDCOM
			result.Relationships = append(result.Relationships, models.Relationship{
				PersonIDFrom: husband,
				PersonIDTo:   wife,
				Type:         models.RelationshipPartnerMarried,
			})
		}
	}

	return result
}

// parseGEDCOMLine parses a single GEDCOM line
func parseGEDCOMLine(parts []string) (int, string, string) {
	level := 0
	tag := ""
	value := ""

	if len(parts) > 0 {
		fmt.Sscanf(parts[0], "%d", &level)
	}
	if len(parts) > 1 {
		tag = parts[1]
	}
	if len(parts) > 2 {
		value = parts[2]
	}

	return level, tag, value
}

// parseName parses GEDCOM NAME field (e.g., "John /Doe/ Patronymic")
func parseName(nameStr string, person *GEDCOMPerson) {
	nameStr = strings.TrimSpace(nameStr)

	// Format: FirstName LastName /Patronymic/
	// Or with slashes: FirstName /LastName/

	// Try to extract parts from NAME field
	if strings.Contains(nameStr, "/") {
		// Has formatted name
		parts := strings.Split(nameStr, "/")
		if len(parts) >= 3 {
			person.FirstName = strings.TrimSpace(parts[0])
			person.LastName = strings.TrimSpace(parts[1])
			if len(parts) > 2 {
				person.Patronymic = strings.TrimSpace(parts[2])
			}
		}
	} else {
		// Simple name without slashes
		nameParts := strings.Fields(nameStr)
		if len(nameParts) >= 1 {
			person.FirstName = nameParts[0]
		}
		if len(nameParts) >= 2 {
			person.LastName = nameParts[1]
		}
		if len(nameParts) >= 3 {
			person.Patronymic = nameParts[2]
		}
	}
}

// extractID extracts GEDCOM ID from reference (e.g., "@I1@" -> "I1")
func extractID(ref string) string {
	ref = strings.TrimSpace(ref)
	if strings.HasPrefix(ref, "@") && strings.HasSuffix(ref, "@") {
		return strings.Trim(ref, "@")
	}
	return ""
}

// parseUUID converts string UUID to uuid.UUID
func parseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}
