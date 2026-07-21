package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/Bad-Utya/myforebears-backend/services/customtree/internal/domain"
	"github.com/Bad-Utya/myforebears-backend/services/customtree/internal/layout"
	store "github.com/Bad-Utya/myforebears-backend/services/customtree/internal/storage"
	"github.com/google/uuid"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrInvalid         = errors.New("invalid input")
	ErrNotFound        = errors.New("not found")
	ErrForbidden       = errors.New("forbidden")
	ErrConflict        = errors.New("conflict")
	ErrDeleteForbidden = errors.New("root entities and entities with children cannot be deleted")
)

type Service struct {
	db      store.Store
	objects store.Objects
}

func New(db store.Store, objects store.Objects) *Service { return &Service{db, objects} }
func (s *Service) Close()                                { s.db.Close() }
func parseID(v string) (uuid.UUID, error) {
	id, err := uuid.Parse(v)
	if err != nil {
		return id, ErrInvalid
	}
	return id, nil
}
func clean(v string) string { return strings.TrimSpace(v) }
func (s *Service) CreateTree(ctx context.Context, user int, name, description, down, up, rootName string) (domain.Tree, domain.Entity, error) {
	if user <= 0 || clean(name) == "" || clean(down) == "" || clean(up) == "" {
		return domain.Tree{}, domain.Entity{}, ErrInvalid
	}
	now := time.Now()
	tid, eid := uuid.New(), uuid.New()
	t := domain.Tree{ID: tid, CreatorID: user, Name: clean(name), Description: clean(description), RelationDown: clean(down), RelationUp: clean(up), RootEntityID: eid, IsViewRestricted: true, CreatedAt: now}
	e := domain.Entity{ID: eid, TreeID: tid, Name: clean(rootName), CreatedAt: now}
	if e.Name == "" {
		e.Name = "Root"
	}
	if err := s.db.CreateTreeWithRoot(ctx, t, e); err != nil {
		return t, e, err
	}
	return t, e, nil
}
func (s *Service) GetTree(ctx context.Context, id string) (domain.Tree, error) {
	x, err := parseID(id)
	if err != nil {
		return domain.Tree{}, err
	}
	t, err := s.db.GetTree(ctx, x)
	if errors.Is(err, store.ErrNotFound) {
		return t, ErrNotFound
	}
	return t, err
}
func (s *Service) ListTreesByOwner(ctx context.Context, user int) ([]domain.Tree, error) {
	if user <= 0 {
		return nil, ErrInvalid
	}
	return s.db.ListTreesByOwner(ctx, user, false)
}
func (s *Service) ListPublicTreesByOwner(ctx context.Context, user int) ([]domain.Tree, error) {
	if user <= 0 {
		return nil, ErrInvalid
	}
	return s.db.ListTreesByOwner(ctx, user, true)
}
func (s *Service) ListRandomPublicTrees(ctx context.Context, n int) ([]domain.Tree, error) {
	if n <= 0 || n > 100 {
		return nil, ErrInvalid
	}
	return s.db.ListRandomPublicTrees(ctx, n)
}
func (s *Service) SearchPublicTrees(ctx context.Context, q string, tagCodes []string, n int) ([]domain.Tree, error) {
	tagCodes = normalizeTags(tagCodes)
	if (clean(q) == "" && len(tagCodes) == 0) || n <= 0 || n > 100 {
		return nil, ErrInvalid
	}
	items, err := s.db.SearchPublicTrees(ctx, clean(q), tagCodes, n)
	if errors.Is(err, store.ErrConflict) {
		return nil, ErrInvalid
	}
	return items, err
}

func normalizeTags(codes []string) []string {
	seen := make(map[string]struct{}, len(codes))
	result := make([]string, 0, len(codes))
	for _, code := range codes {
		code = strings.ToLower(clean(code))
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		result = append(result, code)
	}
	return result
}

func (s *Service) SetTreeTags(ctx context.Context, user int, id string, codes []string) (domain.Tree, error) {
	t, err := s.GetTree(ctx, id)
	if err != nil {
		return t, err
	}
	if user <= 0 || t.CreatorID != user {
		return t, ErrForbidden
	}
	if err := s.db.SetTreeTags(ctx, t.ID, normalizeTags(codes)); err != nil {
		if errors.Is(err, store.ErrConflict) {
			return t, ErrInvalid
		}
		return t, err
	}
	return s.GetTree(ctx, id)
}
func (s *Service) UpdateTree(ctx context.Context, user int, t domain.Tree) (domain.Tree, error) {
	old, err := s.GetTree(ctx, t.ID.String())
	if err != nil {
		return old, err
	}
	if old.CreatorID != user {
		return old, ErrForbidden
	}
	if clean(t.Name) == "" || clean(t.RelationDown) == "" || clean(t.RelationUp) == "" {
		return old, ErrInvalid
	}
	root, err := s.db.GetEntity(ctx, t.RootEntityID)
	if err != nil || root.TreeID != old.ID {
		return old, ErrInvalid
	}
	old.Name, old.Description, old.RelationDown, old.RelationUp = clean(t.Name), clean(t.Description), clean(t.RelationDown), clean(t.RelationUp)
	old.RootEntityID = t.RootEntityID
	old.IsViewRestricted = t.IsViewRestricted
	old.IsPublicOnMainPage = t.IsPublicOnMainPage
	if err := s.db.UpdateTree(ctx, old); err != nil {
		return old, err
	}
	return old, nil
}
func (s *Service) DeleteTree(ctx context.Context, user int, id string) error {
	t, err := s.GetTree(ctx, id)
	if err != nil {
		return err
	}
	if t.CreatorID != user {
		return ErrForbidden
	}
	entities, _ := s.db.ListEntities(ctx, t.ID)
	for _, e := range entities {
		photos, _ := s.db.ListPhotos(ctx, e.ID)
		for _, p := range photos {
			_ = s.objects.Delete(ctx, p.ObjectKey)
		}
	}
	return s.db.DeleteTree(ctx, t.ID)
}
func (s *Service) AddAccessEmail(ctx context.Context, id, email string) error {
	t, err := s.GetTree(ctx, id)
	if err != nil {
		return err
	}
	_ = t
	if clean(email) == "" {
		return ErrInvalid
	}
	err = s.db.AddAccessEmail(ctx, t.ID, strings.ToLower(clean(email)))
	if errors.Is(err, store.ErrConflict) {
		return ErrConflict
	}
	return err
}
func (s *Service) ListAccessEmails(ctx context.Context, id string) ([]string, error) {
	t, err := s.GetTree(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.db.ListAccessEmails(ctx, t.ID)
}
func (s *Service) DeleteAccessEmail(ctx context.Context, id, email string) error {
	t, err := s.GetTree(ctx, id)
	if err != nil {
		return err
	}
	return s.db.DeleteAccessEmail(ctx, t.ID, strings.ToLower(clean(email)))
}
func (s *Service) IsAccessEmailAllowed(ctx context.Context, id, email string) (bool, error) {
	t, err := s.GetTree(ctx, id)
	if err != nil {
		return false, err
	}
	return s.db.IsAccessEmailAllowed(ctx, t.ID, strings.ToLower(clean(email)))
}
func (s *Service) CreateEntity(ctx context.Context, treeID, parentID, name, description string) (domain.Entity, error) {
	t, err := s.GetTree(ctx, treeID)
	if err != nil {
		return domain.Entity{}, err
	}
	if clean(name) == "" {
		return domain.Entity{}, ErrInvalid
	}
	if clean(parentID) == "" {
		return domain.Entity{}, ErrInvalid
	}
	e := domain.Entity{ID: uuid.New(), TreeID: t.ID, Name: clean(name), Description: clean(description), CreatedAt: time.Now()}
	if err := s.db.CreateEntity(ctx, e); err != nil {
		return e, err
	}
	p, err := s.GetEntity(ctx, treeID, parentID)
	if err != nil {
		_ = s.db.DeleteEntity(ctx, e.ID)
		return e, err
	}
	if err := s.AddEdge(ctx, treeID, p.ID.String(), e.ID.String()); err != nil {
		_ = s.db.DeleteEntity(ctx, e.ID)
		return e, err
	}
	return e, nil
}
func (s *Service) AddParent(ctx context.Context, treeID, childID, name, description string) (domain.Entity, error) {
	t, err := s.GetTree(ctx, treeID)
	if err != nil {
		return domain.Entity{}, err
	}
	if clean(name) == "" || clean(childID) == "" {
		return domain.Entity{}, ErrInvalid
	}
	child, err := s.GetEntity(ctx, treeID, childID)
	if err != nil {
		return domain.Entity{}, err
	}
	e := domain.Entity{ID: uuid.New(), TreeID: t.ID, Name: clean(name), Description: clean(description), CreatedAt: time.Now()}
	if err := s.db.CreateParent(ctx, t.ID, child.ID, e); err != nil {
		if errors.Is(err, store.ErrConflict) {
			return domain.Entity{}, ErrConflict
		}
		return domain.Entity{}, err
	}
	return e, nil
}
func (s *Service) GetEntity(ctx context.Context, treeID, entityID string) (domain.Entity, error) {
	tid, err := parseID(treeID)
	if err != nil {
		return domain.Entity{}, err
	}
	id, err := parseID(entityID)
	if err != nil {
		return domain.Entity{}, err
	}
	e, err := s.db.GetEntity(ctx, id)
	if errors.Is(err, store.ErrNotFound) {
		return e, ErrNotFound
	}
	if err == nil && e.TreeID != tid {
		return e, ErrNotFound
	}
	return e, err
}
func (s *Service) ListEntities(ctx context.Context, id string) ([]domain.Entity, error) {
	t, err := s.GetTree(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.db.ListEntities(ctx, t.ID)
}
func (s *Service) UpdateEntity(ctx context.Context, treeID, id, name, description string) (domain.Entity, error) {
	e, err := s.GetEntity(ctx, treeID, id)
	if err != nil {
		return e, err
	}
	if clean(name) == "" {
		return e, ErrInvalid
	}
	e.Name, e.Description = clean(name), clean(description)
	return e, s.db.UpdateEntity(ctx, e)
}
func (s *Service) DeleteEntity(ctx context.Context, treeID, id string) error {
	t, err := s.GetTree(ctx, treeID)
	if err != nil {
		return err
	}
	e, err := s.GetEntity(ctx, treeID, id)
	if err != nil {
		return err
	}
	if e.ID == t.RootEntityID {
		return ErrDeleteForbidden
	}
	has, err := s.db.HasChildren(ctx, e.ID)
	if err != nil {
		return err
	}
	if has {
		return ErrDeleteForbidden
	}
	photos, _ := s.db.ListPhotos(ctx, e.ID)
	if err := s.db.DeleteEntity(ctx, e.ID); err != nil {
		return err
	}
	for _, p := range photos {
		_ = s.objects.Delete(ctx, p.ObjectKey)
	}
	return nil
}
func (s *Service) AddEdge(ctx context.Context, treeID, parentID, childID string) error {
	t, err := s.GetTree(ctx, treeID)
	if err != nil {
		return err
	}
	p, err := s.GetEntity(ctx, treeID, parentID)
	if err != nil {
		return err
	}
	c, err := s.GetEntity(ctx, treeID, childID)
	if err != nil {
		return err
	}
	if p.ID == c.ID {
		return ErrInvalid
	}
	edges, err := s.db.ListEdges(ctx, t.ID)
	if err != nil {
		return err
	}
	children := map[uuid.UUID][]uuid.UUID{}
	for _, e := range edges {
		children[e.ParentID] = append(children[e.ParentID], e.ChildID)
	}
	queue := []uuid.UUID{c.ID}
	seen := map[uuid.UUID]bool{}
	for len(queue) > 0 {
		x := queue[0]
		queue = queue[1:]
		if x == p.ID {
			return ErrConflict
		}
		if seen[x] {
			continue
		}
		seen[x] = true
		queue = append(queue, children[x]...)
	}
	err = s.db.AddEdge(ctx, t.ID, p.ID, c.ID)
	if errors.Is(err, store.ErrConflict) {
		return ErrConflict
	}
	return err
}
func (s *Service) RemoveEdge(ctx context.Context, treeID, parentID, childID string) error {
	_, err := s.GetTree(ctx, treeID)
	if err != nil {
		return err
	}
	p, err := parseID(parentID)
	if err != nil {
		return err
	}
	c, err := parseID(childID)
	if err != nil {
		return err
	}
	return s.db.RemoveEdge(ctx, p, c)
}
func (s *Service) GetContent(ctx context.Context, id string) (domain.Tree, []domain.Entity, []domain.Edge, error) {
	t, err := s.GetTree(ctx, id)
	if err != nil {
		return t, nil, nil, err
	}
	entities, err := s.db.ListEntities(ctx, t.ID)
	if err != nil {
		return t, nil, nil, err
	}
	edges, err := s.db.ListEdges(ctx, t.ID)
	return t, entities, edges, err
}
func (s *Service) UploadPhoto(ctx context.Context, treeID, entityID, file, mime string, data []byte, avatar bool) (domain.Photo, error) {
	e, err := s.GetEntity(ctx, treeID, entityID)
	if err != nil {
		return domain.Photo{}, err
	}
	if len(data) == 0 || len(data) > 15<<20 || !strings.HasPrefix(strings.ToLower(mime), "image/") {
		return domain.Photo{}, ErrInvalid
	}
	if avatar {
		if err := s.db.UnsetAvatar(ctx, e.ID); err != nil {
			return domain.Photo{}, err
		}
	}
	p := domain.Photo{ID: uuid.New(), EntityID: e.ID, FileName: filepath.Base(file), MIMEType: mime, SizeBytes: int64(len(data)), IsAvatar: avatar, CreatedAt: time.Now()}
	p.ObjectKey = fmt.Sprintf("custom-trees/%s/entities/%s/%s_%s", e.TreeID, e.ID, p.ID, p.FileName)
	if err := s.objects.Put(ctx, p.ObjectKey, data, mime); err != nil {
		return p, err
	}
	if err := s.db.CreatePhoto(ctx, p); err != nil {
		_ = s.objects.Delete(ctx, p.ObjectKey)
		return p, err
	}
	if avatar {
		_ = s.db.SetEntityAvatar(ctx, e.ID, &p.ID)
	}
	return p, nil
}
func (s *Service) ListPhotos(ctx context.Context, treeID, entityID string) ([]domain.Photo, error) {
	e, err := s.GetEntity(ctx, treeID, entityID)
	if err != nil {
		return nil, err
	}
	return s.db.ListPhotos(ctx, e.ID)
}
func (s *Service) GetPhoto(ctx context.Context, treeID, entityID, photoID string) (domain.Photo, []byte, error) {
	e, err := s.GetEntity(ctx, treeID, entityID)
	if err != nil {
		return domain.Photo{}, nil, err
	}
	id, err := parseID(photoID)
	if err != nil {
		return domain.Photo{}, nil, err
	}
	p, err := s.db.GetPhoto(ctx, id)
	if err != nil || p.EntityID != e.ID {
		return p, nil, ErrNotFound
	}
	data, err := s.objects.Get(ctx, p.ObjectKey)
	return p, data, err
}
func (s *Service) DeletePhoto(ctx context.Context, treeID, entityID, photoID string) error {
	e, err := s.GetEntity(ctx, treeID, entityID)
	if err != nil {
		return err
	}
	id, err := parseID(photoID)
	if err != nil {
		return err
	}
	p, err := s.db.GetPhoto(ctx, id)
	if err != nil || p.EntityID != e.ID {
		return ErrNotFound
	}
	p, err = s.db.DeletePhoto(ctx, id)
	if err != nil {
		return err
	}
	if p.IsAvatar {
		_ = s.db.SetEntityAvatar(ctx, e.ID, nil)
	}
	_ = s.objects.Delete(ctx, p.ObjectKey)
	return nil
}
func (s *Service) Render(ctx context.Context, treeID, rootID string) (layout.Result, error) {
	t, entities, edges, err := s.GetContent(ctx, treeID)
	if err != nil {
		return layout.Result{}, err
	}
	root := t.RootEntityID
	if clean(rootID) != "" {
		root, err = parseID(rootID)
		if err != nil {
			return layout.Result{}, err
		}
	}
	return layout.Build(root, entities, edges, t.RelationDown, t.RelationUp), nil
}
