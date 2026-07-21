package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/Bad-Utya/myforebears-backend/services/customtree/internal/domain"
	store "github.com/Bad-Utya/myforebears-backend/services/customtree/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct{ pool *pgxpool.Pool }

func New(host string, port int, user, password, db string) (*Storage, error) {
	p, err := pgxpool.New(context.Background(), fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, db))
	if err != nil {
		return nil, err
	}
	if err = p.Ping(context.Background()); err != nil {
		return nil, err
	}
	return &Storage{p}, nil
}
func (s *Storage) Close() { s.pool.Close() }
func (s *Storage) CreateTreeWithRoot(ctx context.Context, t domain.Tree, e domain.Entity) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `INSERT INTO custom_trees(id,creator_id,name,description,relation_down,relation_up,root_entity_id,is_view_restricted,is_public_on_main_page,created_at)VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, t.ID, t.CreatorID, t.Name, t.Description, t.RelationDown, t.RelationUp, t.RootEntityID, t.IsViewRestricted, t.IsPublicOnMainPage, t.CreatedAt)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO custom_entities(id,tree_id,name,description,created_at)VALUES($1,$2,$3,$4,$5)`, e.ID, e.TreeID, e.Name, e.Description, e.CreatedAt)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func scanTree(row interface{ Scan(...any) error }) (domain.Tree, error) {
	var t domain.Tree
	err := row.Scan(&t.ID, &t.CreatorID, &t.Name, &t.Description, &t.RelationDown, &t.RelationUp, &t.RootEntityID, &t.IsViewRestricted, &t.IsPublicOnMainPage, &t.CreatedAt)
	return t, err
}

const treeCols = `id,creator_id,name,description,relation_down,relation_up,root_entity_id,is_view_restricted,is_public_on_main_page,created_at`

func (s *Storage) GetTree(ctx context.Context, id uuid.UUID) (domain.Tree, error) {
	t, err := scanTree(s.pool.QueryRow(ctx, `SELECT `+treeCols+` FROM custom_trees WHERE id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return t, store.ErrNotFound
	}
	if err != nil {
		return t, err
	}
	t.Tags, err = s.listTreeTags(ctx, id)
	return t, err
}
func (s *Storage) listTrees(ctx context.Context, q string, args ...any) ([]domain.Tree, error) {
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Tree{}
	for rows.Next() {
		t, err := scanTree(rows)
		if err != nil {
			return nil, err
		}
		t.Tags, err = s.listTreeTags(ctx, t.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func (s *Storage) ListTreesByOwner(ctx context.Context, owner int, public bool) ([]domain.Tree, error) {
	if public {
		return s.listTrees(ctx, `SELECT `+treeCols+` FROM custom_trees WHERE creator_id=$1 AND is_public_on_main_page ORDER BY created_at DESC`, owner)
	}
	return s.listTrees(ctx, `SELECT `+treeCols+` FROM custom_trees WHERE creator_id=$1 ORDER BY created_at DESC`, owner)
}
func (s *Storage) ListRandomPublicTrees(ctx context.Context, n int) ([]domain.Tree, error) {
	return s.listTrees(ctx, `SELECT `+treeCols+` FROM custom_trees WHERE is_public_on_main_page ORDER BY random() LIMIT $1`, n)
}
func (s *Storage) UpdateTree(ctx context.Context, t domain.Tree) error {
	tag, err := s.pool.Exec(ctx, `UPDATE custom_trees SET name=$1,description=$2,relation_down=$3,relation_up=$4,root_entity_id=$5,is_view_restricted=$6,is_public_on_main_page=$7,updated_at=NOW() WHERE id=$8`, t.Name, t.Description, t.RelationDown, t.RelationUp, t.RootEntityID, t.IsViewRestricted, t.IsPublicOnMainPage, t.ID)
	if err == nil && tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return err
}
func (s *Storage) DeleteTree(ctx context.Context, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM custom_trees WHERE id=$1`, id)
	if err == nil && tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return err
}
func (s *Storage) AddAccessEmail(ctx context.Context, id uuid.UUID, email string) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO custom_tree_access_emails(tree_id,email)VALUES($1,$2)`, id, email)
	if isUnique(err) {
		return store.ErrConflict
	}
	return err
}
func (s *Storage) ListAccessEmails(ctx context.Context, id uuid.UUID) ([]string, error) {
	rows, err := s.pool.Query(ctx, `SELECT email FROM custom_tree_access_emails WHERE tree_id=$1 ORDER BY email`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var x string
		if err := rows.Scan(&x); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (s *Storage) DeleteAccessEmail(ctx context.Context, id uuid.UUID, email string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM custom_tree_access_emails WHERE tree_id=$1 AND email=$2`, id, email)
	if err == nil && tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return err
}
func (s *Storage) IsAccessEmailAllowed(ctx context.Context, id uuid.UUID, email string) (bool, error) {
	var ok bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM custom_tree_access_emails WHERE tree_id=$1 AND email=$2)`, id, email).Scan(&ok)
	return ok, err
}
func (s *Storage) CreateEntity(ctx context.Context, e domain.Entity) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO custom_entities(id,tree_id,name,description,created_at)VALUES($1,$2,$3,$4,$5)`, e.ID, e.TreeID, e.Name, e.Description, e.CreatedAt)
	return err
}
func (s *Storage) CreateParent(ctx context.Context, treeID, childID uuid.UUID, e domain.Entity) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `INSERT INTO custom_entities(id,tree_id,name,description,created_at)VALUES($1,$2,$3,$4,$5)`, e.ID, e.TreeID, e.Name, e.Description, e.CreatedAt); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO custom_edges(tree_id,parent_id,child_id)VALUES($1,$2,$3)`, treeID, e.ID, childID); err != nil {
		if isUnique(err) {
			return store.ErrConflict
		}
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE custom_trees SET root_entity_id=$1,updated_at=NOW() WHERE id=$2 AND root_entity_id=$3`, e.ID, treeID, childID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func scanEntity(row interface{ Scan(...any) error }) (domain.Entity, error) {
	var e domain.Entity
	err := row.Scan(&e.ID, &e.TreeID, &e.Name, &e.Description, &e.AvatarPhotoID, &e.CreatedAt)
	return e, err
}

const entityCols = `id,tree_id,name,description,avatar_photo_id,created_at`

func (s *Storage) GetEntity(ctx context.Context, id uuid.UUID) (domain.Entity, error) {
	e, err := scanEntity(s.pool.QueryRow(ctx, `SELECT `+entityCols+` FROM custom_entities WHERE id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return e, store.ErrNotFound
	}
	return e, err
}
func (s *Storage) ListEntities(ctx context.Context, id uuid.UUID) ([]domain.Entity, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+entityCols+` FROM custom_entities WHERE tree_id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Entity{}
	for rows.Next() {
		e, err := scanEntity(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
func (s *Storage) UpdateEntity(ctx context.Context, e domain.Entity) error {
	tag, err := s.pool.Exec(ctx, `UPDATE custom_entities SET name=$1,description=$2,updated_at=NOW() WHERE id=$3`, e.Name, e.Description, e.ID)
	if err == nil && tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return err
}
func (s *Storage) DeleteEntity(ctx context.Context, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM custom_entities WHERE id=$1`, id)
	if err == nil && tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return err
}
func (s *Storage) AddEdge(ctx context.Context, tree, parent, child uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO custom_edges(tree_id,parent_id,child_id)VALUES($1,$2,$3)`, tree, parent, child)
	if isUnique(err) {
		return store.ErrConflict
	}
	return err
}
func (s *Storage) RemoveEdge(ctx context.Context, parent, child uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM custom_edges WHERE parent_id=$1 AND child_id=$2`, parent, child)
	if err == nil && tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return err
}
func (s *Storage) ListEdges(ctx context.Context, id uuid.UUID) ([]domain.Edge, error) {
	rows, err := s.pool.Query(ctx, `SELECT parent_id,child_id FROM custom_edges WHERE tree_id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Edge{}
	for rows.Next() {
		var e domain.Edge
		if err := rows.Scan(&e.ParentID, &e.ChildID); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
func (s *Storage) HasChildren(ctx context.Context, id uuid.UUID) (bool, error) {
	var ok bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM custom_edges WHERE parent_id=$1)`, id).Scan(&ok)
	return ok, err
}
func (s *Storage) CreatePhoto(ctx context.Context, p domain.Photo) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO custom_photos(id,entity_id,file_name,mime_type,size_bytes,object_key,is_avatar,created_at)VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, p.ID, p.EntityID, p.FileName, p.MIMEType, p.SizeBytes, p.ObjectKey, p.IsAvatar, p.CreatedAt)
	return err
}
func scanPhoto(row interface{ Scan(...any) error }) (domain.Photo, error) {
	var p domain.Photo
	err := row.Scan(&p.ID, &p.EntityID, &p.FileName, &p.MIMEType, &p.SizeBytes, &p.ObjectKey, &p.IsAvatar, &p.CreatedAt)
	return p, err
}

const photoCols = `id,entity_id,file_name,mime_type,size_bytes,object_key,is_avatar,created_at`

func (s *Storage) GetPhoto(ctx context.Context, id uuid.UUID) (domain.Photo, error) {
	p, err := scanPhoto(s.pool.QueryRow(ctx, `SELECT `+photoCols+` FROM custom_photos WHERE id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return p, store.ErrNotFound
	}
	return p, err
}
func (s *Storage) ListPhotos(ctx context.Context, id uuid.UUID) ([]domain.Photo, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+photoCols+` FROM custom_photos WHERE entity_id=$1 ORDER BY created_at`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Photo{}
	for rows.Next() {
		p, err := scanPhoto(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
func (s *Storage) UnsetAvatar(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `UPDATE custom_photos SET is_avatar=FALSE WHERE entity_id=$1 AND is_avatar`, id)
	return err
}
func (s *Storage) SetEntityAvatar(ctx context.Context, id uuid.UUID, pid *uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `UPDATE custom_entities SET avatar_photo_id=$1 WHERE id=$2`, pid, id)
	return err
}
func (s *Storage) DeletePhoto(ctx context.Context, id uuid.UUID) (domain.Photo, error) {
	p, err := scanPhoto(s.pool.QueryRow(ctx, `DELETE FROM custom_photos WHERE id=$1 RETURNING `+photoCols, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return p, store.ErrNotFound
	}
	return p, err
}
func isUnique(err error) bool { var p *pgconn.PgError; return errors.As(err, &p) && p.Code == "23505" }
