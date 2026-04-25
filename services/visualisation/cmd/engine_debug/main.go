package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	familytreepb "github.com/Bad-Utya/myforebears-backend/gen/go/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/engine"
	"github.com/google/uuid"
)

const rootPersonID = "8239e07f-199c-4eb1-bb52-379ddf560da4"
const outputFileName = "debug_render.svg"

const testJSON = `{
  "data": {
    "persons": [
      {
        "avatar_photo_id": "a78b8c84-77af-4920-a9af-c0e9ebe5a803",
        "first_name": "Данил",
        "gender": "GENDER_MALE",
        "id": "8239e07f-199c-4eb1-bb52-379ddf560da4",
        "last_name": "Стадник",
        "patronymic": "Сергеевич",
        "tree_id": "62cfa487-c7da-490d-8a24-0f188f76fd71"
      },
      {
        "avatar_photo_id": "",
        "first_name": "Юлия",
        "gender": "GENDER_FEMALE",
        "id": "e0b0c555-ca30-455e-9db1-7323069022aa",
        "last_name": "Стадник",
        "patronymic": "Александровна",
        "tree_id": "62cfa487-c7da-490d-8a24-0f188f76fd71"
      },
      {
        "avatar_photo_id": "",
        "first_name": "Сергей",
        "gender": "GENDER_MALE",
        "id": "3cd68c81-087b-4e0b-b39a-e674d285a070",
        "last_name": "Стадник",
        "patronymic": "Александрович",
        "tree_id": "62cfa487-c7da-490d-8a24-0f188f76fd71"
      }
    ],
    "relationships": [
      {
        "person_id_from": "e0b0c555-ca30-455e-9db1-7323069022aa",
        "person_id_to": "8239e07f-199c-4eb1-bb52-379ddf560da4",
        "type": "RELATIONSHIP_PARENT_CHILD"
      },
      {
        "person_id_from": "3cd68c81-087b-4e0b-b39a-e674d285a070",
        "person_id_to": "e0b0c555-ca30-455e-9db1-7323069022aa",
        "type": "RELATIONSHIP_PARTNER_MARRIED"
      },
      {
        "person_id_from": "3cd68c81-087b-4e0b-b39a-e674d285a070",
        "person_id_to": "8239e07f-199c-4eb1-bb52-379ddf560da4",
        "type": "RELATIONSHIP_PARENT_CHILD"
      }
    ]
  }
}`

type inputPayload struct {
	Data struct {
		Persons       []inputPerson       `json:"persons"`
		Relationships []inputRelationship `json:"relationships"`
	} `json:"data"`
}

type inputPerson struct {
	ID         string `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Patronymic string `json:"patronymic"`
	Gender     string `json:"gender"`
}

type inputRelationship struct {
	From string `json:"person_id_from"`
	To   string `json:"person_id_to"`
	Type string `json:"type"`
}

func main() {
	payload := inputPayload{}
	if err := json.Unmarshal([]byte(testJSON), &payload); err != nil {
		panic(fmt.Errorf("failed to parse embedded test json: %w", err))
	}

	content := &familytreepb.GetTreeContentResponse{
		Persons:       make([]*familytreepb.Person, 0, len(payload.Data.Persons)),
		Relationships: make([]*familytreepb.Relationship, 0, len(payload.Data.Relationships)),
	}

	for _, person := range payload.Data.Persons {
		content.Persons = append(content.Persons, &familytreepb.Person{
			Id:         person.ID,
			FirstName:  person.FirstName,
			LastName:   person.LastName,
			Patronymic: person.Patronymic,
			Gender:     parseGender(person.Gender),
		})
	}

	for _, rel := range payload.Data.Relationships {
		content.Relationships = append(content.Relationships, &familytreepb.Relationship{
			PersonIdFrom: rel.From,
			PersonIdTo:   rel.To,
			Type:         parseRelationshipType(rel.Type),
		})
	}

	rootID, err := uuid.Parse(rootPersonID)
	if err != nil {
		panic(fmt.Errorf("invalid root id: %w", err))
	}

	fmt.Println("Running RenderSVGWithTrace with embedded test data...")
	svg, err := engine.RenderSVGWithTrace(
		models.VisualisationTypeAncestorsAndDescendants,
		rootID,
		nil,
		content,
		os.Stdout,
	)
	if err != nil {
		panic(fmt.Errorf("render failed: %w", err))
	}

	if err := os.WriteFile(outputFileName, svg, 0644); err != nil {
		panic(fmt.Errorf("failed to save svg: %w", err))
	}

	absPath, err := filepath.Abs(outputFileName)
	if err != nil {
		fmt.Printf("SVG saved to %s\n", outputFileName)
		return
	}
	fmt.Printf("SVG saved to %s\n", absPath)
}

func parseGender(raw string) familytreepb.Gender {
	switch raw {
	case "GENDER_FEMALE":
		return familytreepb.Gender_GENDER_FEMALE
	case "GENDER_MALE":
		return familytreepb.Gender_GENDER_MALE
	default:
		return familytreepb.Gender_GENDER_UNSPECIFIED
	}
}

func parseRelationshipType(raw string) familytreepb.RelationshipType {
	switch raw {
	case "RELATIONSHIP_PARENT_CHILD":
		return familytreepb.RelationshipType_RELATIONSHIP_PARENT_CHILD
	case "RELATIONSHIP_PARTNER":
		return familytreepb.RelationshipType_RELATIONSHIP_PARTNER
	case "RELATIONSHIP_PARTNER_UNMARRIED":
		return familytreepb.RelationshipType_RELATIONSHIP_PARTNER_UNMARRIED
	case "RELATIONSHIP_PARTNER_MARRIED":
		return familytreepb.RelationshipType_RELATIONSHIP_PARTNER_MARRIED
	case "RELATIONSHIP_PARTNER_DIVORCED":
		return familytreepb.RelationshipType_RELATIONSHIP_PARTNER_DIVORCED
	default:
		return familytreepb.RelationshipType_RELATIONSHIP_TYPE_UNSPECIFIED
	}
}
