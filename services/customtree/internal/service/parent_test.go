package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Bad-Utya/myforebears-backend/services/customtree/internal/domain"
	store "github.com/Bad-Utya/myforebears-backend/services/customtree/internal/storage"
	"github.com/google/uuid"
)

type parentStore struct {
	store.Store
	tree      domain.Tree
	child     domain.Entity
	created   domain.Entity
	createErr error
}

func (s *parentStore) GetTree(context.Context, uuid.UUID) (domain.Tree, error) {
	return s.tree, nil
}

func (s *parentStore) GetEntity(_ context.Context, id uuid.UUID) (domain.Entity, error) {
	if id != s.child.ID {
		return domain.Entity{}, store.ErrNotFound
	}
	return s.child, nil
}

func (s *parentStore) CreateParent(_ context.Context, treeID, childID uuid.UUID, entity domain.Entity) error {
	if treeID != s.tree.ID || childID != s.child.ID {
		return errors.New("unexpected parent attachment")
	}
	s.created = entity
	return s.createErr
}

func TestAddParentCreatesEntityAboveChild(t *testing.T) {
	treeID, childID := uuid.New(), uuid.New()
	db := &parentStore{
		tree:  domain.Tree{ID: treeID, RootEntityID: childID},
		child: domain.Entity{ID: childID, TreeID: treeID},
	}

	got, err := New(db, nil).AddParent(context.Background(), treeID.String(), childID.String(), " Parent ", " Description ")
	if err != nil {
		t.Fatalf("AddParent() error = %v", err)
	}
	if got.ID == uuid.Nil || got.TreeID != treeID || got.Name != "Parent" || got.Description != "Description" {
		t.Fatalf("AddParent() = %#v", got)
	}
	if db.created != got {
		t.Fatalf("stored parent = %#v, want %#v", db.created, got)
	}
}

func TestAddParentMapsStorageConflict(t *testing.T) {
	treeID, childID := uuid.New(), uuid.New()
	db := &parentStore{
		tree:      domain.Tree{ID: treeID, RootEntityID: childID},
		child:     domain.Entity{ID: childID, TreeID: treeID},
		createErr: store.ErrConflict,
	}

	_, err := New(db, nil).AddParent(context.Background(), treeID.String(), childID.String(), "Parent", "")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("AddParent() error = %v, want %v", err, ErrConflict)
	}
}
