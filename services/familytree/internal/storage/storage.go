package storage

import (
	"context"
	"errors"

	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/google/uuid"
)

var (
	ErrPersonNotFound      = errors.New("person not found")
	ErrTreeNotFound        = errors.New("tree not found")
	ErrRelationshipExists  = errors.New("relationship already exists")
	ErrRelationshipMissing = errors.New("relationship not found")
)

type PersonStorage interface {
	CreateTree(ctx context.Context, tree models.Tree) error
	GetTree(ctx context.Context, treeID uuid.UUID) (models.Tree, error)
	GetTreesByCreator(ctx context.Context, creatorID int) ([]models.Tree, error)
	CreatePerson(ctx context.Context, person models.Person) error
	GetPerson(ctx context.Context, personID uuid.UUID) (models.Person, error)
	UpdatePerson(ctx context.Context, person models.Person) error
	DeletePerson(ctx context.Context, personID uuid.UUID) error
	GetPersonsByTree(ctx context.Context, treeID uuid.UUID) ([]models.Person, error)
	Close()
}

type RelationshipStorage interface {
	EnsurePersonNode(ctx context.Context, personID uuid.UUID, treeID uuid.UUID) error
	DeletePersonNode(ctx context.Context, personID uuid.UUID) error
	CreateRelationship(ctx context.Context, personIDFrom uuid.UUID, personIDTo uuid.UUID, relType models.RelationshipType) error
	RemoveRelationship(ctx context.Context, personIDFrom uuid.UUID, personIDTo uuid.UUID, relType models.RelationshipType) error
	GetRelatives(ctx context.Context, personID uuid.UUID) ([]models.Relative, error)
	GetTreeRelationships(ctx context.Context, treeID uuid.UUID) ([]models.Relationship, error)
	Close(ctx context.Context) error
}
