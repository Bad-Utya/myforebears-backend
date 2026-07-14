package domain

import (
	"github.com/google/uuid"
	"time"
)

type Tree struct {
	ID                                          uuid.UUID
	CreatorID                                   int
	Name, Description, RelationDown, RelationUp string
	RootEntityID                                uuid.UUID
	IsViewRestricted, IsPublicOnMainPage        bool
	CreatedAt                                   time.Time
	Tags                                        []Tag
	SimilarityScore                             float64
}
type Tag struct{ Code, Name, Description string }
type Entity struct {
	ID, TreeID        uuid.UUID
	Name, Description string
	AvatarPhotoID     *uuid.UUID
	CreatedAt         time.Time
}
type Edge struct{ ParentID, ChildID uuid.UUID }
type Photo struct {
	ID, EntityID                  uuid.UUID
	FileName, MIMEType, ObjectKey string
	SizeBytes                     int64
	IsAvatar                      bool
	CreatedAt                     time.Time
}
