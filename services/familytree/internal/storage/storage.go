package storage

import (
	"context"
	"errors"

	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/google/uuid"
)

var (
	ErrPersonNotFound          = errors.New("person not found")
	ErrTreeNotFound            = errors.New("tree not found")
	ErrRelationshipExists      = errors.New("relationship already exists")
	ErrRelationshipMissing     = errors.New("relationship not found")
	ErrTreeAccessEmailExists   = errors.New("tree access email already exists")
	ErrTreeAccessEmailNotFound = errors.New("tree access email not found")
	ErrPublicPersonNotFound    = errors.New("public person not found")
	ErrUnknownTag              = errors.New("unknown tag")
)

type PersonStorage interface {
	CreateTree(ctx context.Context, tree models.Tree) error
	GetTree(ctx context.Context, treeID uuid.UUID) (models.Tree, error)
	UpdateTreeRootPersonID(ctx context.Context, treeID uuid.UUID, rootPersonID uuid.UUID) error
	UpdateTreeSettings(ctx context.Context, treeID uuid.UUID, isViewRestricted bool, isPublicOnMainPage bool, name string, description *string) error
	AddTreeAccessEmail(ctx context.Context, treeID uuid.UUID, email string) error
	ListTreeAccessEmails(ctx context.Context, treeID uuid.UUID) ([]string, error)
	IsTreeAccessEmailAllowed(ctx context.Context, treeID uuid.UUID, email string) (bool, error)
	DeleteTreeAccessEmail(ctx context.Context, treeID uuid.UUID, email string) error
	GetTreesByCreator(ctx context.Context, creatorID int) ([]models.Tree, error)
	GetPublicTreesByCreator(ctx context.Context, creatorID int) ([]models.Tree, error)
	GetPublicTreesByName(ctx context.Context, nameQuery string, limit int) ([]models.Tree, error)
	GetRandomPublicTrees(ctx context.Context, limit int) ([]models.Tree, error)
	CreatePerson(ctx context.Context, person models.Person) error
	GetPerson(ctx context.Context, personID uuid.UUID) (models.Person, error)
	UpdatePerson(ctx context.Context, person models.Person) error
	UpdatePersonAvatarPhoto(ctx context.Context, personID uuid.UUID, avatarPhotoID *uuid.UUID) error
	DeletePerson(ctx context.Context, personID uuid.UUID) error
	GetPersonsByTree(ctx context.Context, treeID uuid.UUID) ([]models.Person, error)
	Close()
}

type PublicPersonStorage interface {
	CreatePublicPerson(ctx context.Context, person models.PublicPerson) error
	GetPublicPerson(ctx context.Context, personID uuid.UUID) (models.PublicPerson, error)
	ListRandomPublicPersons(ctx context.Context, limit int) ([]models.PublicPerson, error)
	ListPublicPersonsByOwner(ctx context.Context, ownerUserID int, limit int) ([]models.PublicPerson, error)
	SearchPublicPersons(ctx context.Context, query string, limit int) ([]models.PublicPerson, error)
	UpdatePublicPerson(ctx context.Context, person models.PublicPerson) error
	SetPublicPersonAvatarPhoto(ctx context.Context, personID uuid.UUID, avatarPhotoID *uuid.UUID) error
	DeletePublicPerson(ctx context.Context, personID uuid.UUID) error
}

type TagStorage interface {
	ListTags(context.Context) ([]models.Tag, error)
	SetTreeTags(context.Context, uuid.UUID, []string) error
	SetPublicPersonTags(context.Context, uuid.UUID, []string) error
	SearchPublicTrees(context.Context, string, []string, int) ([]models.Tree, error)
	SearchPublicPersonsByTags(context.Context, string, []string, int) ([]models.PublicPerson, error)
}

type RelationshipStorage interface {
	EnsurePersonNode(ctx context.Context, personID uuid.UUID, treeID uuid.UUID) error
	DeletePersonNode(ctx context.Context, personID uuid.UUID) error
	CreateRelationship(ctx context.Context, personIDFrom uuid.UUID, personIDTo uuid.UUID, relType models.RelationshipType) error
	SetPartnerRelationshipStatus(ctx context.Context, personID1 uuid.UUID, personID2 uuid.UUID, status models.PartnerRelationshipStatus) error
	RemoveRelationship(ctx context.Context, personIDFrom uuid.UUID, personIDTo uuid.UUID, relType models.RelationshipType) error
	GetRelatives(ctx context.Context, personID uuid.UUID) ([]models.Relative, error)
	GetTreeRelationships(ctx context.Context, treeID uuid.UUID) ([]models.Relationship, error)
	GetTreeRelationshipsWithinDepth(ctx context.Context, treeID uuid.UUID, rootPersonID uuid.UUID, maxDepth int) ([]models.Relationship, error)
	Close(ctx context.Context) error
}
