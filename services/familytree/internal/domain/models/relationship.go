package models

import "github.com/google/uuid"

type RelationshipType string

const (
	RelationshipParentChild RelationshipType = "PARENT_CHILD"
	RelationshipPartner     RelationshipType = "PARTNER"
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
