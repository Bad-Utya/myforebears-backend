package models

import "github.com/google/uuid"

type RelationshipType string

const (
	RelationshipParentChild      RelationshipType = "PARENT_CHILD"
	RelationshipPartner          RelationshipType = "PARTNER"
	RelationshipPartnerUnmarried RelationshipType = "PARTNER_UNMARRIED"
	RelationshipPartnerMarried   RelationshipType = "PARTNER_MARRIED"
	RelationshipPartnerDivorced  RelationshipType = "PARTNER_DIVORCED"
)

type PartnerRelationshipStatus string

const (
	PartnerRelationshipStatusUnmarried PartnerRelationshipStatus = "UNMARRIED"
	PartnerRelationshipStatusMarried   PartnerRelationshipStatus = "MARRIED"
	PartnerRelationshipStatusDivorced  PartnerRelationshipStatus = "DIVORCED"
)

type RelationDirection string

const (
	DirectionOutgoing RelationDirection = "OUTGOING"
	DirectionIncoming RelationDirection = "INCOMING"
)

type Relationship struct {
	PersonIDFrom uuid.UUID
	PersonIDTo   uuid.UUID
	Type         RelationshipType
}

type Relative struct {
	Person           Person
	RelationshipType RelationshipType
	Direction        RelationDirection
}
