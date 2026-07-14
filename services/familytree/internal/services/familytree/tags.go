package familytree

import (
	"context"
	"errors"
	"strings"

	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/storage"
)

func normalizeTagCodes(codes []string) []string {
	seen := make(map[string]struct{}, len(codes))
	result := make([]string, 0, len(codes))
	for _, code := range codes {
		code = strings.ToLower(strings.TrimSpace(code))
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

func (s *Service) ListTags(ctx context.Context) ([]models.Tag, error) {
	return s.tagStorage.ListTags(ctx)
}

func (s *Service) SetTreeTags(ctx context.Context, userID int, treeID string, codes []string) (models.Tree, error) {
	if userID <= 0 {
		return models.Tree{}, ErrInvalidUserID
	}
	tree, err := s.GetTree(ctx, treeID)
	if err != nil {
		return models.Tree{}, err
	}
	if tree.CreatorID != userID {
		return models.Tree{}, ErrForbidden
	}
	if err := s.tagStorage.SetTreeTags(ctx, tree.ID, normalizeTagCodes(codes)); err != nil {
		if errors.Is(err, storage.ErrUnknownTag) {
			return models.Tree{}, ErrUnknownTag
		}
		return models.Tree{}, err
	}
	return s.GetTree(ctx, treeID)
}

func (s *Service) SetPublicPersonTags(ctx context.Context, userID int, personID string, codes []string) (models.PublicPerson, error) {
	if userID <= 0 {
		return models.PublicPerson{}, ErrInvalidUserID
	}
	person, err := s.GetPublicPerson(ctx, personID)
	if err != nil {
		return models.PublicPerson{}, err
	}
	if person.OwnerUserID != userID {
		return models.PublicPerson{}, ErrForbidden
	}
	if err := s.tagStorage.SetPublicPersonTags(ctx, person.ID, normalizeTagCodes(codes)); err != nil {
		if errors.Is(err, storage.ErrUnknownTag) {
			return models.PublicPerson{}, ErrUnknownTag
		}
		return models.PublicPerson{}, err
	}
	return s.GetPublicPerson(ctx, personID)
}
