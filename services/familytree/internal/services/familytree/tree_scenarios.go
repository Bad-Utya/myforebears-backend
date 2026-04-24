package familytree

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	eventspb "github.com/Bad-Utya/myforebears-backend/gen/go/events"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/storage"
	"github.com/google/uuid"
)

type ParentRole string

const (
	ParentRoleFather    ParentRole = "FATHER"
	ParentRoleMother    ParentRole = "MOTHER"
	birthEventTypeID    string     = "4af8f935-180f-4be6-8f7a-f6ecf90af4b2"
	marriageEventTypeID string     = "2c2f5f12-5476-4ef4-89df-85d0a6f4a6bc"
)

func (s *Service) CreateTree(ctx context.Context, requestUserID int) (models.Tree, models.Person, error) {
	const op = "service.familytree.CreateTree"
	log := s.log.With(slog.String("op", op))

	log.Info("creating tree", slog.Int("request_user_id", requestUserID))

	if requestUserID <= 0 {
		return models.Tree{}, models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}

	tree := models.Tree{
		ID:                 uuid.New(),
		CreatorID:          requestUserID,
		CreatedAt:          time.Now(),
		IsViewRestricted:   true,
		IsPublicOnMainPage: false,
	}

	if err := s.personStorage.CreateTree(ctx, tree); err != nil {
		log.Error("failed to create tree", slog.String("error", err.Error()))
		return models.Tree{}, models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	root, err := s.createPersonRecord(ctx, requestUserID, tree.ID, models.GenderMale, "", "", "")
	if err != nil {
		log.Error("failed to create root person", slog.String("error", err.Error()))
		return models.Tree{}, models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("tree created", slog.String("tree_id", tree.ID.String()), slog.String("root_person_id", root.ID.String()))

	return tree, root, nil
}

func (s *Service) ListTreesByCreator(ctx context.Context, requestUserID int) ([]models.Tree, error) {
	const op = "service.familytree.ListTreesByCreator"
	log := s.log.With(slog.String("op", op))

	log.Info("listing trees by creator", slog.Int("request_user_id", requestUserID))

	if requestUserID <= 0 {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}

	trees, err := s.personStorage.GetTreesByCreator(ctx, requestUserID)
	if err != nil {
		log.Error("failed to list trees", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("trees listed", slog.Int("count", len(trees)))

	return trees, nil
}

func (s *Service) GetTree(ctx context.Context, treeID string) (models.Tree, error) {
	const op = "service.familytree.GetTree"
	log := s.log.With(slog.String("op", op))

	log.Info("getting tree", slog.String("tree_id", treeID))

	parsedTreeID, err := s.authorizeTree(ctx, treeID)
	if err != nil {
		log.Error("failed to validate tree", slog.String("error", err.Error()))
		return models.Tree{}, fmt.Errorf("%s: %w", op, err)
	}

	tree, err := s.personStorage.GetTree(ctx, parsedTreeID)
	if err != nil {
		if errors.Is(err, storage.ErrTreeNotFound) {
			return models.Tree{}, fmt.Errorf("%s: %w", op, ErrTreeNotFound)
		}
		return models.Tree{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("tree loaded", slog.String("tree_id", tree.ID.String()))

	return tree, nil
}

func (s *Service) GetTreeContent(ctx context.Context, treeID string) ([]models.Person, []models.Relationship, error) {
	const op = "service.familytree.GetTreeContent"
	log := s.log.With(slog.String("op", op))

	log.Info("getting tree content", slog.String("tree_id", treeID))

	parsedTreeID, err := s.authorizeTree(ctx, treeID)
	if err != nil {
		log.Error("failed to validate tree", slog.String("error", err.Error()))
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	persons, err := s.personStorage.GetPersonsByTree(ctx, parsedTreeID)
	if err != nil {
		log.Error("failed to load persons by tree", slog.String("error", err.Error()))
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	relationships, err := s.relationStorage.GetTreeRelationships(ctx, parsedTreeID)
	if err != nil {
		log.Error("failed to load tree relationships", slog.String("error", err.Error()))
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("tree content loaded", slog.Int("persons_count", len(persons)), slog.Int("relationships_count", len(relationships)))

	return persons, relationships, nil
}

func (s *Service) UpdateTreeSettings(ctx context.Context, treeID string, isViewRestricted bool, isPublicOnMainPage bool) (models.Tree, error) {
	const op = "service.familytree.UpdateTreeSettings"
	log := s.log.With(slog.String("op", op))

	log.Info(
		"updating tree settings",
		slog.String("tree_id", treeID),
		slog.Bool("is_view_restricted", isViewRestricted),
		slog.Bool("is_public_on_main_page", isPublicOnMainPage),
	)

	parsedTreeID, err := s.authorizeTree(ctx, treeID)
	if err != nil {
		log.Error("failed to validate tree", slog.String("error", err.Error()))
		return models.Tree{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.personStorage.UpdateTreeSettings(ctx, parsedTreeID, isViewRestricted, isPublicOnMainPage); err != nil {
		if errors.Is(err, storage.ErrTreeNotFound) {
			log.Info("tree not found", slog.String("tree_id", parsedTreeID.String()))
			return models.Tree{}, fmt.Errorf("%s: %w", op, ErrTreeNotFound)
		}
		log.Error("failed to update tree settings", slog.String("error", err.Error()))
		return models.Tree{}, fmt.Errorf("%s: %w", op, err)
	}

	tree, err := s.personStorage.GetTree(ctx, parsedTreeID)
	if err != nil {
		if errors.Is(err, storage.ErrTreeNotFound) {
			log.Info("tree not found", slog.String("tree_id", parsedTreeID.String()))
			return models.Tree{}, fmt.Errorf("%s: %w", op, ErrTreeNotFound)
		}
		log.Error("failed to load updated tree", slog.String("error", err.Error()))
		return models.Tree{}, fmt.Errorf("%s: %w", op, err)
	}

	return tree, nil
}

func (s *Service) ListPersonsByTree(ctx context.Context, treeID string) ([]models.Person, error) {
	const op = "service.familytree.ListPersonsByTree"
	log := s.log.With(slog.String("op", op))

	log.Info("listing persons by tree", slog.String("tree_id", treeID))

	parsedTreeID, err := s.authorizeTree(ctx, treeID)
	if err != nil {
		log.Error("failed to validate tree", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	persons, err := s.personStorage.GetPersonsByTree(ctx, parsedTreeID)
	if err != nil {
		log.Error("failed to list persons by tree", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("persons listed", slog.Int("count", len(persons)))

	return persons, nil
}

func (s *Service) AddParent(
	ctx context.Context,
	treeID string,
	childID string,
	role ParentRole,
	firstName string,
	lastName string,
	patronymic string,
) (models.Person, *models.Person, error) {
	const op = "service.familytree.AddParent"
	log := s.log.With(slog.String("op", op))

	log.Info("adding parent", slog.String("tree_id", treeID), slog.String("child_id", childID), slog.String("role", string(role)))

	parsedTreeID, err := s.authorizeTree(ctx, treeID)
	if err != nil {
		log.Error("failed to validate tree", slog.String("error", err.Error()))
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	tree, err := s.personStorage.GetTree(ctx, parsedTreeID)
	if err != nil {
		log.Error("failed to load tree", slog.String("error", err.Error()))
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	parsedChildID, err := uuid.Parse(childID)
	if err != nil {
		log.Info("invalid child id", slog.String("child_id", childID))
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	child, err := s.personStorage.GetPerson(ctx, parsedChildID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Info("child not found", slog.String("child_id", parsedChildID.String()))
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		log.Error("failed to load child", slog.String("error", err.Error()))
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}
	if child.TreeID != parsedTreeID {
		log.Info("child tree mismatch", slog.String("child_tree_id", child.TreeID.String()), slog.String("requested_tree_id", parsedTreeID.String()))
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, ErrTreeMismatch)
	}

	targetGender, err := parentRoleGender(role)
	if err != nil {
		log.Info("invalid parent role", slog.String("role", string(role)))
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	parents, err := s.fetchIncomingParents(ctx, child.ID)
	if err != nil {
		log.Error("failed to fetch incoming parents", slog.String("error", err.Error()))
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(parents) >= 2 {
		log.Info("parent limit reached", slog.Int("parents_count", len(parents)))
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, ErrParentLimitReached)
	}

	for _, p := range parents {
		if p.Gender == targetGender {
			log.Info("parent with same role already exists", slog.String("gender", string(targetGender)))
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, ErrParentExists)
		}
	}

	newParent, err := s.createPersonRecord(ctx, tree.CreatorID, parsedTreeID, targetGender, firstName, lastName, patronymic)
	if err != nil {
		log.Error("failed to create new parent", slog.String("error", err.Error()))
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.relationStorage.CreateRelationship(ctx, newParent.ID, child.ID, models.RelationshipParentChild); err != nil {
		if errors.Is(err, storage.ErrRelationshipExists) {
			log.Info("parent-child relationship already exists")
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, ErrRelationshipExists)
		}
		log.Error("failed to create parent-child relationship", slog.String("error", err.Error()))
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(parents) == 1 {
		other := parents[0]
		_, err := s.ensureRelationship(ctx, newParent.ID, other.ID, models.RelationshipPartnerUnmarried)
		if err != nil {
			log.Error("failed to ensure partner relationship", slog.String("error", err.Error()))
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
		}
		return newParent, nil, nil
	}

	autoGender, err := oppositeGender(newParent.Gender)
	if err != nil {
		log.Info("failed to derive opposite gender", slog.String("gender", string(newParent.Gender)))
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	autoParent, err := s.createPersonRecord(ctx, tree.CreatorID, parsedTreeID, autoGender, "", "", "")
	if err != nil {
		log.Error("failed to create auto parent", slog.String("error", err.Error()))
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.relationStorage.CreateRelationship(ctx, autoParent.ID, child.ID, models.RelationshipParentChild); err != nil {
		if errors.Is(err, storage.ErrRelationshipExists) {
			log.Info("auto parent-child relationship already exists")
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, ErrRelationshipExists)
		}
		log.Error("failed to create auto parent-child relationship", slog.String("error", err.Error()))
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = s.ensureRelationship(ctx, newParent.ID, autoParent.ID, models.RelationshipPartnerUnmarried)
	if err != nil {
		log.Error("failed to ensure partner relationship", slog.String("error", err.Error()))
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("parent added", slog.String("parent_id", newParent.ID.String()), slog.String("auto_parent_id", autoParent.ID.String()))

	return newParent, &autoParent, nil
}

func (s *Service) AddChild(
	ctx context.Context,
	treeID string,
	parent1ID string,
	parent2ID string,
	firstName string,
	lastName string,
	patronymic string,
	gender models.Gender,
) (models.Person, *models.Person, error) {
	const op = "service.familytree.AddChild"
	log := s.log.With(slog.String("op", op))

	log.Info("adding child", slog.String("tree_id", treeID), slog.String("parent1_id", parent1ID), slog.String("parent2_id", parent2ID))

	parsedTreeID, err := s.authorizeTree(ctx, treeID)
	if err != nil {
		log.Error("failed to validate tree", slog.String("error", err.Error()))
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	tree, err := s.personStorage.GetTree(ctx, parsedTreeID)
	if err != nil {
		log.Error("failed to load tree", slog.String("error", err.Error()))
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	if strings.TrimSpace(parent1ID) == "" && strings.TrimSpace(parent2ID) == "" {
		log.Info("at least one parent id is required")
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, ErrAtLeastOneParent)
	}

	if !isValidGender(gender) {
		log.Info("invalid child gender", slog.String("gender", string(gender)))
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, ErrInvalidGender)
	}

	if firstName == "" || lastName == "" {
		log.Info("invalid child name")
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, ErrInvalidName)
	}

	parent1, hasParent1, err := s.resolveOptionalParent(ctx, parsedTreeID, parent1ID)
	if err != nil {
		log.Error("failed to resolve first parent", slog.String("error", err.Error()))
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}
	parent2, hasParent2, err := s.resolveOptionalParent(ctx, parsedTreeID, parent2ID)
	if err != nil {
		log.Error("failed to resolve second parent", slog.String("error", err.Error()))
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	var autoCreatedParent *models.Person
	if hasParent1 && !hasParent2 {
		opposite, err := oppositeGender(parent1.Gender)
		if err != nil {
			log.Info("failed to derive opposite gender", slog.String("gender", string(parent1.Gender)))
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
		}
		auto, err := s.createPersonRecord(ctx, tree.CreatorID, parsedTreeID, opposite, "", "", "")
		if err != nil {
			log.Error("failed to create auto parent", slog.String("error", err.Error()))
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
		}
		parent2 = auto
		hasParent2 = true
		autoCreatedParent = &auto
	}
	if !hasParent1 && hasParent2 {
		opposite, err := oppositeGender(parent2.Gender)
		if err != nil {
			log.Info("failed to derive opposite gender", slog.String("gender", string(parent2.Gender)))
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
		}
		auto, err := s.createPersonRecord(ctx, tree.CreatorID, parsedTreeID, opposite, "", "", "")
		if err != nil {
			log.Error("failed to create auto parent", slog.String("error", err.Error()))
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
		}
		parent1 = auto
		hasParent1 = true
		autoCreatedParent = &auto
	}

	child, err := s.createPersonRecord(ctx, tree.CreatorID, parsedTreeID, gender, firstName, lastName, patronymic)
	if err != nil {
		log.Error("failed to create child record", slog.String("error", err.Error()))
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	if hasParent1 {
		if _, err := s.ensureRelationship(ctx, parent1.ID, child.ID, models.RelationshipParentChild); err != nil {
			log.Error("failed to ensure first parent-child relationship", slog.String("error", err.Error()))
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
		}
	}
	if hasParent2 {
		if _, err := s.ensureRelationship(ctx, parent2.ID, child.ID, models.RelationshipParentChild); err != nil {
			log.Error("failed to ensure second parent-child relationship", slog.String("error", err.Error()))
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
		}
	}
	if hasParent1 && hasParent2 {
		_, err := s.ensureRelationship(ctx, parent1.ID, parent2.ID, models.RelationshipPartnerUnmarried)
		if err != nil {
			log.Error("failed to ensure parents partner relationship", slog.String("error", err.Error()))
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	log.Info("child added", slog.String("child_id", child.ID.String()), slog.Bool("auto_parent_created", autoCreatedParent != nil))

	return child, autoCreatedParent, nil
}

func (s *Service) AddPartner(
	ctx context.Context,
	treeID string,
	personID string,
	firstName string,
	lastName string,
	patronymic string,
) (models.Person, error) {
	const op = "service.familytree.AddPartner"
	log := s.log.With(slog.String("op", op))

	log.Info("adding partner", slog.String("tree_id", treeID), slog.String("person_id", personID))

	parsedTreeID, err := s.authorizeTree(ctx, treeID)
	if err != nil {
		log.Error("failed to validate tree", slog.String("error", err.Error()))
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	tree, err := s.personStorage.GetTree(ctx, parsedTreeID)
	if err != nil {
		log.Error("failed to load tree", slog.String("error", err.Error()))
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	if firstName == "" || lastName == "" {
		log.Info("invalid partner name")
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidName)
	}

	baseID, err := uuid.Parse(personID)
	if err != nil {
		log.Info("invalid person id", slog.String("person_id", personID))
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	basePerson, err := s.personStorage.GetPerson(ctx, baseID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Info("person not found", slog.String("person_id", baseID.String()))
			return models.Person{}, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		log.Error("failed to load person", slog.String("error", err.Error()))
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}
	if basePerson.TreeID != parsedTreeID {
		log.Info("person tree mismatch", slog.String("person_tree_id", basePerson.TreeID.String()), slog.String("requested_tree_id", parsedTreeID.String()))
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrTreeMismatch)
	}

	partnerGender, err := oppositeGender(basePerson.Gender)
	if err != nil {
		log.Info("failed to derive opposite gender", slog.String("gender", string(basePerson.Gender)))
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	partner, err := s.createPersonRecord(ctx, tree.CreatorID, parsedTreeID, partnerGender, firstName, lastName, patronymic)
	if err != nil {
		log.Error("failed to create partner", slog.String("error", err.Error()))
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	created, err := s.ensureRelationship(ctx, basePerson.ID, partner.ID, models.RelationshipPartnerMarried)
	if err != nil {
		log.Error("failed to ensure partner relationship", slog.String("error", err.Error()))
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}
	if created {
		if err := s.createAutogeneratedMarriageEvent(ctx, tree.CreatorID, parsedTreeID, basePerson.ID, partner.ID); err != nil {
			log.Error("failed to create autogenerated marriage event", slog.String("error", err.Error()))
			return models.Person{}, fmt.Errorf("%s: %w", op, err)
		}
	}

	log.Info("partner added", slog.String("partner_id", partner.ID.String()))

	return partner, nil
}

func (s *Service) UpdatePersonName(
	ctx context.Context,
	treeID string,
	personID string,
	firstName string,
	lastName string,
	patronymic string,
) (models.Person, error) {
	const op = "service.familytree.UpdatePersonName"
	log := s.log.With(slog.String("op", op))

	log.Info("updating person name", slog.String("tree_id", treeID), slog.String("person_id", personID))

	parsedTreeID, err := s.authorizeTree(ctx, treeID)
	if err != nil {
		log.Error("failed to validate tree", slog.String("error", err.Error()))
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	if firstName == "" || lastName == "" {
		log.Info("invalid person name")
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidName)
	}

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		log.Info("invalid person id", slog.String("person_id", personID))
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	person, err := s.personStorage.GetPerson(ctx, parsedPersonID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Info("person not found", slog.String("person_id", parsedPersonID.String()))
			return models.Person{}, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		log.Error("failed to load person", slog.String("error", err.Error()))
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}
	if person.TreeID != parsedTreeID {
		log.Info("person tree mismatch", slog.String("person_tree_id", person.TreeID.String()), slog.String("requested_tree_id", parsedTreeID.String()))
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrTreeMismatch)
	}

	person.FirstName = firstName
	person.LastName = lastName
	person.Patronymic = patronymic

	if err := s.personStorage.UpdatePerson(ctx, person); err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Info("person not found during update", slog.String("person_id", parsedPersonID.String()))
			return models.Person{}, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		log.Error("failed to update person", slog.String("error", err.Error()))
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("person name updated", slog.String("person_id", person.ID.String()))

	return person, nil
}

func (s *Service) UpdatePersonAvatarPhoto(
	ctx context.Context,
	personID string,
	avatarPhotoID string,
) (models.Person, error) {
	const op = "service.familytree.UpdatePersonAvatarPhoto"
	log := s.log.With(slog.String("op", op))

	log.Info("updating person avatar photo", slog.String("person_id", personID))

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		log.Info("invalid person id", slog.String("person_id", personID))
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	person, err := s.personStorage.GetPerson(ctx, parsedPersonID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Info("person not found", slog.String("person_id", parsedPersonID.String()))
			return models.Person{}, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		log.Error("failed to load person", slog.String("error", err.Error()))
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	var parsedAvatarPhotoID *uuid.UUID
	trimmedAvatar := strings.TrimSpace(avatarPhotoID)
	if trimmedAvatar != "" {
		avatarID, err := uuid.Parse(trimmedAvatar)
		if err != nil {
			log.Info("invalid avatar photo id", slog.String("avatar_photo_id", trimmedAvatar))
			return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
		}
		parsedAvatarPhotoID = &avatarID
	}

	if err := s.personStorage.UpdatePersonAvatarPhoto(ctx, parsedPersonID, parsedAvatarPhotoID); err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Info("person not found during avatar update", slog.String("person_id", parsedPersonID.String()))
			return models.Person{}, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		log.Error("failed to update person avatar photo", slog.String("error", err.Error()))
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	person.AvatarPhotoID = parsedAvatarPhotoID

	log.Info("person avatar photo updated", slog.String("person_id", person.ID.String()))

	return person, nil
}

func (s *Service) DeletePersonInTree(ctx context.Context, treeID string, personID string) error {
	const op = "service.familytree.DeletePersonInTree"
	log := s.log.With(slog.String("op", op))

	log.Info("deleting person in tree", slog.String("tree_id", treeID), slog.String("person_id", personID))

	parsedTreeID, err := s.authorizeTree(ctx, treeID)
	if err != nil {
		log.Error("failed to validate tree", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		log.Info("invalid person id", slog.String("person_id", personID))
		return fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	person, err := s.personStorage.GetPerson(ctx, parsedPersonID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Info("person not found", slog.String("person_id", parsedPersonID.String()))
			return fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		log.Error("failed to load person", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}
	if person.TreeID != parsedTreeID {
		log.Info("person tree mismatch", slog.String("person_tree_id", person.TreeID.String()), slog.String("requested_tree_id", parsedTreeID.String()))
		return fmt.Errorf("%s: %w", op, ErrTreeMismatch)
	}

	if err := s.DeletePerson(ctx, treeID, personID); err != nil {
		log.Error("failed to delete person", slog.String("error", err.Error()))
		return err
	}

	log.Info("person deleted", slog.String("person_id", parsedPersonID.String()))

	return nil
}

func (s *Service) UpdatePartnerRelationshipStatus(
	ctx context.Context,
	treeID string,
	personID1 string,
	personID2 string,
	status models.PartnerRelationshipStatus,
) error {
	const op = "service.familytree.UpdatePartnerRelationshipStatus"
	log := s.log.With(slog.String("op", op))

	parsedTreeID, err := s.authorizeTree(ctx, treeID)
	if err != nil {
		log.Error("failed to validate tree", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	if status != models.PartnerRelationshipStatusUnmarried &&
		status != models.PartnerRelationshipStatusMarried &&
		status != models.PartnerRelationshipStatusDivorced {
		return fmt.Errorf("%s: %w", op, ErrInvalidRelationType)
	}

	id1, err := uuid.Parse(personID1)
	if err != nil {
		return fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}
	id2, err := uuid.Parse(personID2)
	if err != nil {
		return fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}
	if id1 == id2 {
		return fmt.Errorf("%s: %w", op, ErrSelfRelationship)
	}

	person1, err := s.personStorage.GetPerson(ctx, id1)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			return fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	person2, err := s.personStorage.GetPerson(ctx, id2)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			return fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if person1.TreeID != parsedTreeID || person2.TreeID != parsedTreeID {
		return fmt.Errorf("%s: %w", op, ErrTreeMismatch)
	}

	if err := s.relationStorage.SetPartnerRelationshipStatus(ctx, id1, id2, status); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) authorizeTree(ctx context.Context, treeID string) (uuid.UUID, error) {
	const op = "service.familytree.authorizeTree"
	log := s.log.With(slog.String("op", op))

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		log.Info("invalid tree id", slog.String("tree_id", treeID))
		return uuid.Nil, ErrInvalidTreeID
	}

	_, err = s.personStorage.GetTree(ctx, parsedTreeID)
	if err != nil {
		if errors.Is(err, storage.ErrTreeNotFound) {
			log.Info("tree not found", slog.String("tree_id", parsedTreeID.String()))
			return uuid.Nil, ErrTreeNotFound
		}
		log.Error("failed to load tree", slog.String("error", err.Error()))
		return uuid.Nil, err
	}

	return parsedTreeID, nil
}

func (s *Service) fetchIncomingParents(ctx context.Context, childID uuid.UUID) ([]models.Person, error) {
	const op = "service.familytree.fetchIncomingParents"
	log := s.log.With(slog.String("op", op), slog.String("child_id", childID.String()))

	rels, err := s.relationStorage.GetRelatives(ctx, childID)
	if err != nil {
		log.Error("failed to load relatives", slog.String("error", err.Error()))
		return nil, err
	}

	parents := make([]models.Person, 0)
	for _, rel := range rels {
		if rel.RelationshipType != models.RelationshipParentChild || rel.Direction != models.DirectionIncoming {
			continue
		}
		parent, err := s.personStorage.GetPerson(ctx, rel.Person.ID)
		if err != nil {
			log.Error("failed to load parent person", slog.String("parent_id", rel.Person.ID.String()), slog.String("error", err.Error()))
			return nil, err
		}
		parents = append(parents, parent)
	}

	return parents, nil
}

func (s *Service) resolveOptionalParent(ctx context.Context, treeID uuid.UUID, personID string) (models.Person, bool, error) {
	const op = "service.familytree.resolveOptionalParent"
	log := s.log.With(slog.String("op", op), slog.String("tree_id", treeID.String()), slog.String("person_id", personID))

	id := strings.TrimSpace(personID)
	if id == "" {
		return models.Person{}, false, nil
	}

	parsed, err := uuid.Parse(id)
	if err != nil {
		log.Info("invalid person id")
		return models.Person{}, false, ErrInvalidPersonID
	}

	person, err := s.personStorage.GetPerson(ctx, parsed)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			log.Info("person not found", slog.String("person_id", parsed.String()))
			return models.Person{}, false, ErrPersonNotFound
		}
		log.Error("failed to load person", slog.String("error", err.Error()))
		return models.Person{}, false, err
	}
	if person.TreeID != treeID {
		log.Info("person tree mismatch", slog.String("person_tree_id", person.TreeID.String()), slog.String("expected_tree_id", treeID.String()))
		return models.Person{}, false, ErrTreeMismatch
	}

	return person, true, nil
}

func (s *Service) createPersonRecord(
	ctx context.Context,
	requestUserID int,
	treeID uuid.UUID,
	gender models.Gender,
	firstName string,
	lastName string,
	patronymic string,
) (models.Person, error) {
	const op = "service.familytree.createPersonRecord"
	log := s.log.With(slog.String("op", op), slog.String("tree_id", treeID.String()), slog.Int("request_user_id", requestUserID))

	if !isValidGender(gender) {
		log.Info("invalid gender", slog.String("gender", string(gender)))
		return models.Person{}, ErrInvalidGender
	}

	person := models.Person{
		ID:         uuid.New(),
		TreeID:     treeID,
		FirstName:  firstName,
		LastName:   lastName,
		Patronymic: patronymic,
		Gender:     gender,
	}

	if err := s.personStorage.CreatePerson(ctx, person); err != nil {
		log.Error("failed to create person", slog.String("error", err.Error()))
		return models.Person{}, err
	}
	if err := s.relationStorage.EnsurePersonNode(ctx, person.ID, person.TreeID); err != nil {
		log.Error("failed to ensure person node", slog.String("person_id", person.ID.String()), slog.String("error", err.Error()))
		return models.Person{}, err
	}

	if err := s.createAutogeneratedBirthEvent(ctx, requestUserID, treeID, person.ID); err != nil {
		log.Error("failed to create autogenerated birth event", slog.String("person_id", person.ID.String()), slog.String("error", err.Error()))
		return models.Person{}, err
	}

	return person, nil
}

func (s *Service) ensureRelationship(ctx context.Context, fromID uuid.UUID, toID uuid.UUID, relType models.RelationshipType) (bool, error) {
	const op = "service.familytree.ensureRelationship"
	log := s.log.With(
		slog.String("op", op),
		slog.String("from_id", fromID.String()),
		slog.String("to_id", toID.String()),
		slog.String("relationship_type", string(relType)),
	)

	if isPartnerRelationshipType(relType) && fromID.String() > toID.String() {
		fromID, toID = toID, fromID
	}
	if err := s.relationStorage.CreateRelationship(ctx, fromID, toID, relType); err != nil {
		if errors.Is(err, storage.ErrRelationshipExists) {
			log.Info("relationship already exists")
			return false, nil
		}
		log.Error("failed to create relationship", slog.String("error", err.Error()))
		return false, err
	}
	return true, nil
}

func (s *Service) createAutogeneratedBirthEvent(ctx context.Context, requestUserID int, treeID uuid.UUID, personID uuid.UUID) error {
	const op = "service.familytree.createAutogeneratedBirthEvent"
	log := s.log.With(slog.String("op", op), slog.Int("request_user_id", requestUserID), slog.String("tree_id", treeID.String()), slog.String("person_id", personID.String()))

	if s.eventsClient == nil {
		log.Info("events client is not configured, skip autogenerated birth event")
		return nil
	}

	_, err := s.eventsClient.CreateEvent(ctx, &eventspb.CreateEventRequest{
		RequestUserId:       int32(requestUserID),
		TreeId:              treeID.String(),
		EventTypeId:         birthEventTypeID,
		PrimaryPersonIds:    []string{personID.String()},
		AdditionalPersonIds: nil,
		DateIso:             "",
		DateUnknown:         true,
		DatePrecision:       eventspb.EventDatePrecision_EVENT_DATE_PRECISION_DAY,
		DateBound:           eventspb.EventDateBound_EVENT_DATE_BOUND_EXACT,
		IsAutogenerated:     true,
	})

	if err != nil {
		log.Error("failed to create autogenerated birth event", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (s *Service) createAutogeneratedMarriageEvent(ctx context.Context, requestUserID int, treeID uuid.UUID, person1ID uuid.UUID, person2ID uuid.UUID) error {
	const op = "service.familytree.createAutogeneratedMarriageEvent"
	log := s.log.With(slog.String("op", op), slog.Int("request_user_id", requestUserID), slog.String("tree_id", treeID.String()), slog.String("person1_id", person1ID.String()), slog.String("person2_id", person2ID.String()))

	if s.eventsClient == nil {
		log.Info("events client is not configured, skip autogenerated marriage event")
		return nil
	}

	primary := []string{person1ID.String(), person2ID.String()}
	if primary[0] > primary[1] {
		primary[0], primary[1] = primary[1], primary[0]
	}

	_, err := s.eventsClient.CreateEvent(ctx, &eventspb.CreateEventRequest{
		RequestUserId:       int32(requestUserID),
		TreeId:              treeID.String(),
		EventTypeId:         marriageEventTypeID,
		PrimaryPersonIds:    primary,
		AdditionalPersonIds: nil,
		DateIso:             "",
		DateUnknown:         true,
		DatePrecision:       eventspb.EventDatePrecision_EVENT_DATE_PRECISION_DAY,
		DateBound:           eventspb.EventDateBound_EVENT_DATE_BOUND_EXACT,
		IsAutogenerated:     true,
	})

	if err != nil {
		log.Error("failed to create autogenerated marriage event", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func parentRoleGender(role ParentRole) (models.Gender, error) {
	switch role {
	case ParentRoleFather:
		return models.GenderMale, nil
	case ParentRoleMother:
		return models.GenderFemale, nil
	default:
		return "", ErrInvalidParentRole
	}
}

func oppositeGender(g models.Gender) (models.Gender, error) {
	switch g {
	case models.GenderMale:
		return models.GenderFemale, nil
	case models.GenderFemale:
		return models.GenderMale, nil
	default:
		return "", ErrUnknownPersonGender
	}
}
