package familytree

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"

	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/storage"
	"github.com/google/uuid"
)

func (s *Service) GetTreeAccessInfo(ctx context.Context, treeID string) (models.Tree, error) {
	const op = "service.familytree.GetTreeAccessInfo"
	log := s.log.With(slog.String("op", op))

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		log.Info("invalid tree id", slog.String("tree_id", treeID))
		return models.Tree{}, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	tree, err := s.personStorage.GetTree(ctx, parsedTreeID)
	if err != nil {
		if errors.Is(err, storage.ErrTreeNotFound) {
			return models.Tree{}, fmt.Errorf("%s: %w", op, ErrTreeNotFound)
		}
		return models.Tree{}, fmt.Errorf("%s: %w", op, err)
	}

	return tree, nil
}

func (s *Service) IsTreeAccessEmailAllowed(ctx context.Context, treeID string, email string) (bool, error) {
	const op = "service.familytree.IsTreeAccessEmailAllowed"
	log := s.log.With(slog.String("op", op))

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		log.Info("invalid tree id", slog.String("tree_id", treeID))
		return false, fmt.Errorf("%s: %w", op, ErrInvalidTreeID)
	}

	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		log.Info("invalid email", slog.String("email", email))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	if _, err := s.personStorage.GetTree(ctx, parsedTreeID); err != nil {
		if errors.Is(err, storage.ErrTreeNotFound) {
			return false, fmt.Errorf("%s: %w", op, ErrTreeNotFound)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	allowed, err := s.personStorage.IsTreeAccessEmailAllowed(ctx, parsedTreeID, normalizedEmail)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return allowed, nil
}

func (s *Service) AddTreeAccessEmail(ctx context.Context, requestUserID int, treeID string, email string) error {
	const op = "service.familytree.AddTreeAccessEmail"
	log := s.log.With(slog.String("op", op))

	log.Info("adding tree access email", slog.Int("request_user_id", requestUserID), slog.String("tree_id", treeID))

	parsedTreeID, err := s.authorizeTree(ctx, requestUserID, treeID)
	if err != nil {
		log.Error("failed to authorize tree", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		log.Info("invalid email", slog.String("email", email))
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.personStorage.AddTreeAccessEmail(ctx, parsedTreeID, normalizedEmail); err != nil {
		if errors.Is(err, storage.ErrTreeAccessEmailExists) {
			log.Info("tree access email already exists", slog.String("email", normalizedEmail))
			return fmt.Errorf("%s: %w", op, ErrTreeAccessEmailExists)
		}
		log.Error("failed to add tree access email", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) ListTreeAccessEmails(ctx context.Context, requestUserID int, treeID string) ([]string, error) {
	const op = "service.familytree.ListTreeAccessEmails"
	log := s.log.With(slog.String("op", op))

	log.Info("listing tree access emails", slog.Int("request_user_id", requestUserID), slog.String("tree_id", treeID))

	parsedTreeID, err := s.authorizeTree(ctx, requestUserID, treeID)
	if err != nil {
		log.Error("failed to authorize tree", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	emails, err := s.personStorage.ListTreeAccessEmails(ctx, parsedTreeID)
	if err != nil {
		log.Error("failed to list tree access emails", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return emails, nil
}

func (s *Service) DeleteTreeAccessEmail(ctx context.Context, requestUserID int, treeID string, email string) error {
	const op = "service.familytree.DeleteTreeAccessEmail"
	log := s.log.With(slog.String("op", op))

	log.Info("deleting tree access email", slog.Int("request_user_id", requestUserID), slog.String("tree_id", treeID))

	parsedTreeID, err := s.authorizeTree(ctx, requestUserID, treeID)
	if err != nil {
		log.Error("failed to authorize tree", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		log.Info("invalid email", slog.String("email", email))
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.personStorage.DeleteTreeAccessEmail(ctx, parsedTreeID, normalizedEmail); err != nil {
		if errors.Is(err, storage.ErrTreeAccessEmailNotFound) {
			log.Info("tree access email not found", slog.String("email", normalizedEmail))
			return fmt.Errorf("%s: %w", op, ErrTreeAccessEmailNotFound)
		}
		log.Error("failed to delete tree access email", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func normalizeEmail(email string) (string, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if normalizedEmail == "" {
		return "", ErrInvalidEmail
	}

	parsed, err := mail.ParseAddress(normalizedEmail)
	if err != nil || !strings.EqualFold(parsed.Address, normalizedEmail) {
		return "", ErrInvalidEmail
	}

	return normalizedEmail, nil
}
