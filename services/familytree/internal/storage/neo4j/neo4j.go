package neo4j

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/storage"
	"github.com/google/uuid"
	neo4j "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Storage struct {
	driver neo4j.DriverWithContext
}

func New(uri string, username string, password string) (*Storage, error) {
	const op = "storage.neo4j.New"

	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := driver.VerifyConnectivity(context.Background()); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{driver: driver}, nil
}

func (s *Storage) EnsurePersonNode(ctx context.Context, personID uuid.UUID, treeID uuid.UUID) error {
	const op = "storage.neo4j.EnsurePersonNode"

	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.Run(
		ctx,
		`MERGE (p:Person {id: $id})
		 SET p.tree_id = $tree_id`,
		map[string]any{
			"id":      personID.String(),
			"tree_id": treeID.String(),
		},
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) DeletePersonNode(ctx context.Context, personID uuid.UUID) error {
	const op = "storage.neo4j.DeletePersonNode"

	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.Run(
		ctx,
		`MATCH (p:Person {id: $id}) DETACH DELETE p`,
		map[string]any{"id": personID.String()},
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) CreateRelationship(ctx context.Context, personIDFrom uuid.UUID, personIDTo uuid.UUID, relType models.RelationshipType) error {
	const op = "storage.neo4j.CreateRelationship"

	relName, err := toNeo4jRelationship(relType)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	partnerStatus, hasPartnerStatus, err := toNeo4jPartnerStatus(relType)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	query := fmt.Sprintf(
		`MATCH (a:Person {id: $from_id})
		 MATCH (b:Person {id: $to_id})
		 OPTIONAL MATCH (a)-[existing:%s]->(b)
		 WITH a, b, existing
		 WHERE existing IS NULL
		 CREATE (a)-[r:%s]->(b)
		 SET r.partner_status = CASE WHEN $has_partner_status THEN $partner_status ELSE r.partner_status END
		 RETURN COUNT(*)`,
		relName,
		relName,
	)

	res, err := session.Run(
		ctx,
		query,
		map[string]any{
			"from_id":            personIDFrom.String(),
			"to_id":              personIDTo.String(),
			"has_partner_status": hasPartnerStatus,
			"partner_status":     partnerStatus,
		},
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !res.Next(ctx) {
		if err := res.Err(); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return fmt.Errorf("%s: %w", op, storage.ErrRelationshipExists)
	}

	if res.Record().Values[0].(int64) == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrRelationshipExists)
	}

	return nil
}

func (s *Storage) SetPartnerRelationshipStatus(ctx context.Context, personID1 uuid.UUID, personID2 uuid.UUID, status models.PartnerRelationshipStatus) error {
	const op = "storage.neo4j.SetPartnerRelationshipStatus"

	neoStatus, err := toNeo4jPartnerStatusFromModel(status)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if personID1.String() > personID2.String() {
		personID1, personID2 = personID2, personID1
	}

	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err = session.Run(
		ctx,
		`MATCH (a:Person {id: $id1})
		 MATCH (b:Person {id: $id2})
		 MERGE (a)-[r:PARTNER_OF]->(b)
		 SET r.partner_status = $status`,
		map[string]any{
			"id1":    personID1.String(),
			"id2":    personID2.String(),
			"status": neoStatus,
		},
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) RemoveRelationship(ctx context.Context, personIDFrom uuid.UUID, personIDTo uuid.UUID, relType models.RelationshipType) error {
	const op = "storage.neo4j.RemoveRelationship"

	relName, err := toNeo4jRelationship(relType)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	query := fmt.Sprintf(
		`MATCH (a:Person {id: $from_id})-[r:%s]->(b:Person {id: $to_id})
		 DELETE r
		 RETURN COUNT(*)`,
		relName,
	)

	res, err := session.Run(
		ctx,
		query,
		map[string]any{
			"from_id": personIDFrom.String(),
			"to_id":   personIDTo.String(),
		},
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !res.Next(ctx) {
		if err := res.Err(); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return fmt.Errorf("%s: %w", op, storage.ErrRelationshipMissing)
	}

	if res.Record().Values[0].(int64) == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrRelationshipMissing)
	}

	return nil
}

func (s *Storage) GetRelatives(ctx context.Context, personID uuid.UUID) ([]models.Relative, error) {
	const op = "storage.neo4j.GetRelatives"

	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	res, err := session.Run(
		ctx,
		`MATCH (root:Person {id: $id})
		OPTIONAL MATCH (root)-[r1:PARENT_OF|PARTNER_OF]->(other1:Person)
		WITH root, collect({id: other1.id, type: type(r1), partner_status: r1.partner_status, dir: 'OUTGOING'}) AS outgoing

		OPTIONAL MATCH (other2:Person)-[r2:PARENT_OF|PARTNER_OF]->(root)
		WITH outgoing, collect({id: other2.id, type: type(r2), partner_status: r2.partner_status, dir: 'INCOMING'}) AS incoming

		WITH outgoing + incoming AS rels
		UNWIND rels AS rel
		WITH rel WHERE rel.id IS NOT NULL
		RETURN rel.id AS rel_id, rel.type AS rel_type, rel.partner_status AS partner_status, rel.dir AS rel_dir`,
		map[string]any{"id": personID.String()},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	relatives := make([]models.Relative, 0)
	for res.Next(ctx) {
		rec := res.Record()
		idValue, _ := rec.Get("rel_id")
		typeValue, _ := rec.Get("rel_type")
		partnerStatusValue, _ := rec.Get("partner_status")
		dirValue, _ := rec.Get("rel_dir")

		relativeID, err := uuid.Parse(idValue.(string))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		partnerStatus := ""
		if partnerStatusValue != nil {
			if value, ok := partnerStatusValue.(string); ok {
				partnerStatus = value
			}
		}

		relType, err := fromNeo4jRelationship(typeValue.(string), partnerStatus)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		direction, err := fromNeo4jDirection(dirValue.(string))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		relatives = append(relatives, models.Relative{
			Person:           models.Person{ID: relativeID},
			RelationshipType: relType,
			Direction:        direction,
		})
	}

	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return relatives, nil
}

func (s *Storage) GetTreeRelationships(ctx context.Context, treeID uuid.UUID) ([]models.Relationship, error) {
	const op = "storage.neo4j.GetTreeRelationships"

	session := s.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	res, err := session.Run(
		ctx,
		`MATCH (a:Person {tree_id: $tree_id})-[r:PARENT_OF|PARTNER_OF]->(b:Person {tree_id: $tree_id})
		 RETURN a.id AS from_id, b.id AS to_id, type(r) AS rel_type, r.partner_status AS partner_status`,
		map[string]any{"tree_id": treeID.String()},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rels := make([]models.Relationship, 0)
	for res.Next(ctx) {
		rec := res.Record()
		fromValue, _ := rec.Get("from_id")
		toValue, _ := rec.Get("to_id")
		typeValue, _ := rec.Get("rel_type")
		partnerStatusValue, _ := rec.Get("partner_status")

		fromID, err := uuid.Parse(fromValue.(string))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		toID, err := uuid.Parse(toValue.(string))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		partnerStatus := ""
		if partnerStatusValue != nil {
			if value, ok := partnerStatusValue.(string); ok {
				partnerStatus = value
			}
		}

		relType, err := fromNeo4jRelationship(typeValue.(string), partnerStatus)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		rels = append(rels, models.Relationship{
			PersonIDFrom: fromID,
			PersonIDTo:   toID,
			Type:         relType,
		})
	}

	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return rels, nil
}

func (s *Storage) Close(ctx context.Context) error {
	return s.driver.Close(ctx)
}

func toNeo4jRelationship(relType models.RelationshipType) (string, error) {
	switch relType {
	case models.RelationshipParentChild:
		return "PARENT_OF", nil
	case models.RelationshipPartner,
		models.RelationshipPartnerUnmarried,
		models.RelationshipPartnerMarried,
		models.RelationshipPartnerDivorced:
		return "PARTNER_OF", nil
	default:
		return "", errors.New("unknown relationship type")
	}
}

func toNeo4jPartnerStatus(relType models.RelationshipType) (string, bool, error) {
	switch relType {
	case models.RelationshipPartnerUnmarried:
		return "UNMARRIED", true, nil
	case models.RelationshipPartnerMarried:
		return "MARRIED", true, nil
	case models.RelationshipPartnerDivorced:
		return "DIVORCED", true, nil
	case models.RelationshipPartner:
		return "UNMARRIED", true, nil
	case models.RelationshipParentChild:
		return "", false, nil
	default:
		return "", false, errors.New("unknown relationship type")
	}
}

func toNeo4jPartnerStatusFromModel(status models.PartnerRelationshipStatus) (string, error) {
	switch status {
	case models.PartnerRelationshipStatusUnmarried:
		return "UNMARRIED", nil
	case models.PartnerRelationshipStatusMarried:
		return "MARRIED", nil
	case models.PartnerRelationshipStatusDivorced:
		return "DIVORCED", nil
	default:
		return "", errors.New("unknown partner relationship status")
	}
}

func fromNeo4jRelationship(relType string, partnerStatus string) (models.RelationshipType, error) {
	switch relType {
	case "PARENT_OF":
		return models.RelationshipParentChild, nil
	case "PARTNER_OF":
		switch partnerStatus {
		case "MARRIED":
			return models.RelationshipPartnerMarried, nil
		case "DIVORCED":
			return models.RelationshipPartnerDivorced, nil
		default:
			return models.RelationshipPartnerUnmarried, nil
		}
	default:
		return "", errors.New("unknown relationship type")
	}
}

func fromNeo4jDirection(dir string) (models.RelationDirection, error) {
	switch dir {
	case "OUTGOING":
		return models.DirectionOutgoing, nil
	case "INCOMING":
		return models.DirectionIncoming, nil
	default:
		return "", errors.New("unknown relation direction")
	}
}
