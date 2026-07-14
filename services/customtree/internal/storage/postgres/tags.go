package postgres

import (
	"context"
	"fmt"

	"github.com/Bad-Utya/myforebears-backend/services/customtree/internal/domain"
	store "github.com/Bad-Utya/myforebears-backend/services/customtree/internal/storage"
	"github.com/google/uuid"
)

func (s *Storage) validateTagCodes(ctx context.Context, codes []string) error {
	if len(codes) == 0 {
		return nil
	}
	var count int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM tags WHERE code=ANY($1::text[])`, codes).Scan(&count); err != nil {
		return err
	}
	if count != len(codes) {
		return store.ErrConflict
	}
	return nil
}

func (s *Storage) SetTreeTags(ctx context.Context, id uuid.UUID, codes []string) error {
	if err := s.validateTagCodes(ctx, codes); err != nil {
		return err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `DELETE FROM custom_tree_tags WHERE tree_id=$1`, id); err != nil {
		return err
	}
	for _, code := range codes {
		if _, err := tx.Exec(ctx, `INSERT INTO custom_tree_tags(tree_id,tag_code)VALUES($1,$2)`, id, code); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Storage) listTreeTags(ctx context.Context, id uuid.UUID) ([]domain.Tag, error) {
	rows, err := s.pool.Query(ctx, `SELECT t.code,t.name,t.description FROM tags t JOIN custom_tree_tags x ON x.tag_code=t.code WHERE x.tree_id=$1 ORDER BY t.sort_order`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.Tag, 0)
	for rows.Next() {
		var tag domain.Tag
		if err := rows.Scan(&tag.Code, &tag.Name, &tag.Description); err != nil {
			return nil, err
		}
		result = append(result, tag)
	}
	return result, rows.Err()
}

func (s *Storage) SearchPublicTrees(ctx context.Context, q string, codes []string, n int) ([]domain.Tree, error) {
	if err := s.validateTagCodes(ctx, codes); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT tr.id,
		CASE WHEN cardinality($2::text[])=0 THEN 0::float8 ELSE
		COUNT(tt.tag_code) FILTER (WHERE tt.tag_code=ANY($2::text[]))::float8 /
		(cardinality($2::text[]) + COUNT(tt.tag_code) - COUNT(tt.tag_code) FILTER (WHERE tt.tag_code=ANY($2::text[]))) END similarity
		FROM custom_trees tr LEFT JOIN custom_tree_tags tt ON tt.tree_id=tr.id
		WHERE tr.is_public_on_main_page
		AND ($1='' OR concat_ws(' ',tr.name,tr.description) ILIKE '%'||$1||'%')
		AND (cardinality($2::text[])=0 OR EXISTS(SELECT 1 FROM custom_tree_tags hit WHERE hit.tree_id=tr.id AND hit.tag_code=ANY($2::text[])))
		GROUP BY tr.id
		ORDER BY similarity DESC,
		CASE WHEN lower(tr.name)=lower($1) THEN 3 WHEN tr.name ILIKE $1||'%' THEN 2 WHEN tr.name ILIKE '%'||$1||'%' THEN 1 ELSE 0 END DESC,
		tr.updated_at DESC LIMIT $3`, q, codes, n)
	if err != nil {
		return nil, fmt.Errorf("search custom trees: %w", err)
	}
	defer rows.Close()
	type hit struct {
		id    uuid.UUID
		score float64
	}
	hits := make([]hit, 0)
	for rows.Next() {
		var h hit
		if err := rows.Scan(&h.id, &h.score); err != nil {
			return nil, err
		}
		hits = append(hits, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	result := make([]domain.Tree, 0, len(hits))
	for _, h := range hits {
		t, err := s.GetTree(ctx, h.id)
		if err != nil {
			return nil, err
		}
		t.SimilarityScore = h.score
		result = append(result, t)
	}
	return result, nil
}
