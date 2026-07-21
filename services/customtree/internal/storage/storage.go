package storage

import (
	"context"
	"errors"
	"github.com/Bad-Utya/myforebears-backend/services/customtree/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type Store interface {
	CreateTreeWithRoot(context.Context, domain.Tree, domain.Entity) error
	GetTree(context.Context, uuid.UUID) (domain.Tree, error)
	ListTreesByOwner(context.Context, int, bool) ([]domain.Tree, error)
	ListRandomPublicTrees(context.Context, int) ([]domain.Tree, error)
	SearchPublicTrees(context.Context, string, []string, int) ([]domain.Tree, error)
	SetTreeTags(context.Context, uuid.UUID, []string) error
	UpdateTree(context.Context, domain.Tree) error
	DeleteTree(context.Context, uuid.UUID) error
	AddAccessEmail(context.Context, uuid.UUID, string) error
	ListAccessEmails(context.Context, uuid.UUID) ([]string, error)
	DeleteAccessEmail(context.Context, uuid.UUID, string) error
	IsAccessEmailAllowed(context.Context, uuid.UUID, string) (bool, error)
	CreateEntity(context.Context, domain.Entity) error
	CreateParent(context.Context, uuid.UUID, uuid.UUID, domain.Entity) error
	GetEntity(context.Context, uuid.UUID) (domain.Entity, error)
	ListEntities(context.Context, uuid.UUID) ([]domain.Entity, error)
	UpdateEntity(context.Context, domain.Entity) error
	DeleteEntity(context.Context, uuid.UUID) error
	AddEdge(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
	RemoveEdge(context.Context, uuid.UUID, uuid.UUID) error
	ListEdges(context.Context, uuid.UUID) ([]domain.Edge, error)
	HasChildren(context.Context, uuid.UUID) (bool, error)
	CreatePhoto(context.Context, domain.Photo) error
	GetPhoto(context.Context, uuid.UUID) (domain.Photo, error)
	ListPhotos(context.Context, uuid.UUID) ([]domain.Photo, error)
	UnsetAvatar(context.Context, uuid.UUID) error
	SetEntityAvatar(context.Context, uuid.UUID, *uuid.UUID) error
	DeletePhoto(context.Context, uuid.UUID) (domain.Photo, error)
	Close()
}
type Objects interface {
	Put(context.Context, string, []byte, string) error
	Get(context.Context, string) ([]byte, error)
	Delete(context.Context, string) error
}
