package postgres

import (
	"context"
	"fmt"

	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/storage"
	"github.com/google/uuid"
)

func (s *Storage) ListTags(ctx context.Context) ([]models.Tag, error) {
	rows, err := s.pool.Query(ctx, `SELECT code,name,description FROM tags ORDER BY sort_order`)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer rows.Close()
	result := make([]models.Tag, 0)
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.Code, &tag.Name, &tag.Description); err != nil {
			return nil, err
		}
		result = append(result, tag)
	}
	return result, rows.Err()
}

func (s *Storage) validateTagCodes(ctx context.Context, codes []string) error {
	if len(codes) == 0 {
		return nil
	}
	var count int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM tags WHERE code=ANY($1::text[])`, codes).Scan(&count); err != nil {
		return err
	}
	if count != len(codes) {
		return storage.ErrUnknownTag
	}
	return nil
}

func (s *Storage) SetTreeTags(ctx context.Context, treeID uuid.UUID, codes []string) error {
	if err := s.validateTagCodes(ctx, codes); err != nil {
		return err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `DELETE FROM tree_tags WHERE tree_id=$1`, treeID); err != nil {
		return err
	}
	for _, code := range codes {
		if _, err := tx.Exec(ctx, `INSERT INTO tree_tags(tree_id,tag_code) VALUES($1,$2)`, treeID, code); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Storage) SetPublicPersonTags(ctx context.Context, personID uuid.UUID, codes []string) error {
	if err := s.validateTagCodes(ctx, codes); err != nil {
		return err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `DELETE FROM public_person_tags WHERE public_person_id=$1`, personID); err != nil {
		return err
	}
	for _, code := range codes {
		if _, err := tx.Exec(ctx, `INSERT INTO public_person_tags(public_person_id,tag_code) VALUES($1,$2)`, personID, code); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Storage) listTreeTags(ctx context.Context, id uuid.UUID) ([]models.Tag, error) {
	return s.listObjectTags(ctx, `SELECT t.code,t.name,t.description FROM tags t JOIN tree_tags x ON x.tag_code=t.code WHERE x.tree_id=$1 ORDER BY t.sort_order`, id)
}

func (s *Storage) listPublicPersonTags(ctx context.Context, id uuid.UUID) ([]models.Tag, error) {
	return s.listObjectTags(ctx, `SELECT t.code,t.name,t.description FROM tags t JOIN public_person_tags x ON x.tag_code=t.code WHERE x.public_person_id=$1 ORDER BY t.sort_order`, id)
}

func (s *Storage) listObjectTags(ctx context.Context, query string, id uuid.UUID) ([]models.Tag, error) {
	rows, err := s.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]models.Tag, 0)
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.Code, &tag.Name, &tag.Description); err != nil {
			return nil, err
		}
		result = append(result, tag)
	}
	return result, rows.Err()
}

func (s *Storage) SearchPublicTrees(ctx context.Context, query string, codes []string, limit int) ([]models.Tree, error) {
	if err := s.validateTagCodes(ctx, codes); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT tr.id,
		CASE WHEN cardinality($2::text[])=0 THEN 0::float8 ELSE
			COUNT(tt.tag_code) FILTER (WHERE tt.tag_code=ANY($2::text[]))::float8 /
			(cardinality($2::text[]) + COUNT(tt.tag_code) - COUNT(tt.tag_code) FILTER (WHERE tt.tag_code=ANY($2::text[])))
		END AS similarity
		FROM trees tr LEFT JOIN tree_tags tt ON tt.tree_id=tr.id
		WHERE tr.is_public_on_main_page=TRUE
		AND ($1='' OR concat_ws(' ',tr.name,tr.description) ILIKE '%'||$1||'%')
		AND (cardinality($2::text[])=0 OR EXISTS(SELECT 1 FROM tree_tags hit WHERE hit.tree_id=tr.id AND hit.tag_code=ANY($2::text[])))
		GROUP BY tr.id
		ORDER BY similarity DESC,
			CASE WHEN lower(tr.name)=lower($1) THEN 3 WHEN tr.name ILIKE $1||'%' THEN 2 WHEN tr.name ILIKE '%'||$1||'%' THEN 1 ELSE 0 END DESC,
			tr.created_at DESC LIMIT $3`, query, codes, limit)
	if err != nil {
		return nil, fmt.Errorf("search public trees: %w", err)
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
	result := make([]models.Tree, 0, len(hits))
	for _, h := range hits {
		tree, err := s.GetTree(ctx, h.id)
		if err != nil {
			return nil, err
		}
		tree.SimilarityScore = h.score
		result = append(result, tree)
	}
	return result, nil
}

func (s *Storage) SearchPublicPersonsByTags(ctx context.Context, query string, codes []string, limit int) ([]models.PublicPerson, error) {
	if err := s.validateTagCodes(ctx, codes); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT p.id,
		CASE WHEN cardinality($2::text[])=0 THEN 0::float8 ELSE
			COUNT(pt.tag_code) FILTER (WHERE pt.tag_code=ANY($2::text[]))::float8 /
			(cardinality($2::text[]) + COUNT(pt.tag_code) - COUNT(pt.tag_code) FILTER (WHERE pt.tag_code=ANY($2::text[])))
		END AS similarity
		FROM public_persons p LEFT JOIN public_person_tags pt ON pt.public_person_id=p.id
		WHERE ($1='' OR concat_ws(' ',p.first_name,p.last_name,p.patronymic,p.biography) ILIKE '%'||$1||'%')
		AND (cardinality($2::text[])=0 OR EXISTS(SELECT 1 FROM public_person_tags hit WHERE hit.public_person_id=p.id AND hit.tag_code=ANY($2::text[])))
		GROUP BY p.id
		ORDER BY similarity DESC,
			CASE WHEN lower(concat_ws(' ',p.first_name,p.last_name))=lower($1) THEN 3 WHEN concat_ws(' ',p.first_name,p.last_name) ILIKE $1||'%' THEN 2 ELSE 0 END DESC,
			p.updated_at DESC LIMIT $3`, query, codes, limit)
	if err != nil {
		return nil, fmt.Errorf("search public persons by tags: %w", err)
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
	result := make([]models.PublicPerson, 0, len(hits))
	for _, h := range hits {
		person, err := s.GetPublicPerson(ctx, h.id)
		if err != nil {
			return nil, err
		}
		person.SimilarityScore = h.score
		result = append(result, person)
	}
	return result, nil
}
