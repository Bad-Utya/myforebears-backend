package familytree

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/storage"
	"github.com/google/uuid"
)

type ParentRole string

const (
	ParentRoleFather ParentRole = "FATHER"
	ParentRoleMother ParentRole = "MOTHER"
)

func (s *Service) CreateTree(ctx context.Context, requestUserID int) (models.Tree, models.Person, error) {
	const op = "service.familytree.CreateTree"

	if requestUserID <= 0 {
		return models.Tree{}, models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}

	tree := models.Tree{
		ID:        uuid.New(),
		CreatorID: requestUserID,
		CreatedAt: time.Now(),
	}

	if err := s.personStorage.CreateTree(ctx, tree); err != nil {
		return models.Tree{}, models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	root, err := s.createPersonRecord(ctx, tree.ID, models.GenderMale, "", "", "")
	if err != nil {
		return models.Tree{}, models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	return tree, root, nil
}

func (s *Service) ListTreesByCreator(ctx context.Context, requestUserID int) ([]models.Tree, error) {
	const op = "service.familytree.ListTreesByCreator"

	if requestUserID <= 0 {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}

	trees, err := s.personStorage.GetTreesByCreator(ctx, requestUserID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return trees, nil
}

func (s *Service) GetTreeForUser(ctx context.Context, requestUserID int, treeID string) ([]models.Person, []models.Relationship, error) {
	const op = "service.familytree.GetTreeForUser"

	parsedTreeID, err := s.authorizeTree(ctx, requestUserID, treeID)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	persons, err := s.personStorage.GetPersonsByTree(ctx, parsedTreeID)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	relationships, err := s.relationStorage.GetTreeRelationships(ctx, parsedTreeID)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	return persons, relationships, nil
}

func (s *Service) ListPersonsByTree(ctx context.Context, requestUserID int, treeID string) ([]models.Person, error) {
	const op = "service.familytree.ListPersonsByTree"

	parsedTreeID, err := s.authorizeTree(ctx, requestUserID, treeID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	persons, err := s.personStorage.GetPersonsByTree(ctx, parsedTreeID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return persons, nil
}

func (s *Service) GetPersonInTree(ctx context.Context, requestUserID int, treeID string, personID string) (models.Person, error) {
	const op = "service.familytree.GetPersonInTree"

	parsedTreeID, err := s.authorizeTree(ctx, requestUserID, treeID)
	if err != nil {
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	person, err := s.personStorage.GetPerson(ctx, parsedPersonID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			return models.Person{}, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	if person.TreeID != parsedTreeID {
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrTreeMismatch)
	}

	return person, nil
}

func (s *Service) AddParent(
	ctx context.Context,
	requestUserID int,
	treeID string,
	childID string,
	role ParentRole,
	firstName string,
	lastName string,
	patronymic string,
) (models.Person, *models.Person, error) {
	const op = "service.familytree.AddParent"

	parsedTreeID, err := s.authorizeTree(ctx, requestUserID, treeID)
	if err != nil {
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	parsedChildID, err := uuid.Parse(childID)
	if err != nil {
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	child, err := s.personStorage.GetPerson(ctx, parsedChildID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}
	if child.TreeID != parsedTreeID {
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, ErrTreeMismatch)
	}

	targetGender, err := parentRoleGender(role)
	if err != nil {
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	parents, err := s.fetchIncomingParents(ctx, child.ID)
	if err != nil {
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(parents) >= 2 {
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, ErrParentLimitReached)
	}

	for _, p := range parents {
		if p.Gender == targetGender {
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, ErrParentExists)
		}
	}

	newParent, err := s.createPersonRecord(ctx, parsedTreeID, targetGender, firstName, lastName, patronymic)
	if err != nil {
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.relationStorage.CreateRelationship(ctx, newParent.ID, child.ID, models.RelationshipParentChild); err != nil {
		if errors.Is(err, storage.ErrRelationshipExists) {
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, ErrRelationshipExists)
		}
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(parents) == 1 {
		other := parents[0]
		if err := s.ensureRelationship(ctx, newParent.ID, other.ID, models.RelationshipPartner); err != nil {
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
		}
		return newParent, nil, nil
	}

	autoGender, err := oppositeGender(newParent.Gender)
	if err != nil {
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	autoParent, err := s.createPersonRecord(ctx, parsedTreeID, autoGender, "", "", "")
	if err != nil {
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.relationStorage.CreateRelationship(ctx, autoParent.ID, child.ID, models.RelationshipParentChild); err != nil {
		if errors.Is(err, storage.ErrRelationshipExists) {
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, ErrRelationshipExists)
		}
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.ensureRelationship(ctx, newParent.ID, autoParent.ID, models.RelationshipPartner); err != nil {
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	return newParent, &autoParent, nil
}

func (s *Service) AddChild(
	ctx context.Context,
	requestUserID int,
	treeID string,
	parent1ID string,
	parent2ID string,
	firstName string,
	lastName string,
	patronymic string,
	gender models.Gender,
) (models.Person, *models.Person, error) {
	const op = "service.familytree.AddChild"

	parsedTreeID, err := s.authorizeTree(ctx, requestUserID, treeID)
	if err != nil {
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	if strings.TrimSpace(parent1ID) == "" && strings.TrimSpace(parent2ID) == "" {
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, ErrAtLeastOneParent)
	}

	if !isValidGender(gender) {
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, ErrInvalidGender)
	}

	if firstName == "" || lastName == "" {
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, ErrInvalidName)
	}

	parent1, hasParent1, err := s.resolveOptionalParent(ctx, parsedTreeID, parent1ID)
	if err != nil {
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}
	parent2, hasParent2, err := s.resolveOptionalParent(ctx, parsedTreeID, parent2ID)
	if err != nil {
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	var autoCreatedParent *models.Person
	if hasParent1 && !hasParent2 {
		opposite, err := oppositeGender(parent1.Gender)
		if err != nil {
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
		}
		auto, err := s.createPersonRecord(ctx, parsedTreeID, opposite, "", "", "")
		if err != nil {
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
		}
		parent2 = auto
		hasParent2 = true
		autoCreatedParent = &auto
	}
	if !hasParent1 && hasParent2 {
		opposite, err := oppositeGender(parent2.Gender)
		if err != nil {
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
		}
		auto, err := s.createPersonRecord(ctx, parsedTreeID, opposite, "", "", "")
		if err != nil {
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
		}
		parent1 = auto
		hasParent1 = true
		autoCreatedParent = &auto
	}

	child, err := s.createPersonRecord(ctx, parsedTreeID, gender, firstName, lastName, patronymic)
	if err != nil {
		return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
	}

	if hasParent1 {
		if err := s.ensureRelationship(ctx, parent1.ID, child.ID, models.RelationshipParentChild); err != nil {
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
		}
	}
	if hasParent2 {
		if err := s.ensureRelationship(ctx, parent2.ID, child.ID, models.RelationshipParentChild); err != nil {
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
		}
	}
	if hasParent1 && hasParent2 {
		if err := s.ensureRelationship(ctx, parent1.ID, parent2.ID, models.RelationshipPartner); err != nil {
			return models.Person{}, nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return child, autoCreatedParent, nil
}

func (s *Service) AddPartner(
	ctx context.Context,
	requestUserID int,
	treeID string,
	personID string,
	firstName string,
	lastName string,
	patronymic string,
) (models.Person, error) {
	const op = "service.familytree.AddPartner"

	parsedTreeID, err := s.authorizeTree(ctx, requestUserID, treeID)
	if err != nil {
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	if firstName == "" || lastName == "" {
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidName)
	}

	baseID, err := uuid.Parse(personID)
	if err != nil {
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	basePerson, err := s.personStorage.GetPerson(ctx, baseID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			return models.Person{}, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}
	if basePerson.TreeID != parsedTreeID {
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrTreeMismatch)
	}

	partnerGender, err := oppositeGender(basePerson.Gender)
	if err != nil {
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	partner, err := s.createPersonRecord(ctx, parsedTreeID, partnerGender, firstName, lastName, patronymic)
	if err != nil {
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.ensureRelationship(ctx, basePerson.ID, partner.ID, models.RelationshipPartner); err != nil {
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	return partner, nil
}

func (s *Service) UpdatePersonName(
	ctx context.Context,
	requestUserID int,
	treeID string,
	personID string,
	firstName string,
	lastName string,
	patronymic string,
) (models.Person, error) {
	const op = "service.familytree.UpdatePersonName"

	parsedTreeID, err := s.authorizeTree(ctx, requestUserID, treeID)
	if err != nil {
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	if firstName == "" || lastName == "" {
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidName)
	}

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	person, err := s.personStorage.GetPerson(ctx, parsedPersonID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			return models.Person{}, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}
	if person.TreeID != parsedTreeID {
		return models.Person{}, fmt.Errorf("%s: %w", op, ErrTreeMismatch)
	}

	person.FirstName = firstName
	person.LastName = lastName
	person.Patronymic = patronymic

	if err := s.personStorage.UpdatePerson(ctx, person); err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			return models.Person{}, fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		return models.Person{}, fmt.Errorf("%s: %w", op, err)
	}

	return person, nil
}

func (s *Service) DeletePersonInTree(ctx context.Context, requestUserID int, treeID string, personID string) error {
	const op = "service.familytree.DeletePersonInTree"

	parsedTreeID, err := s.authorizeTree(ctx, requestUserID, treeID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	parsedPersonID, err := uuid.Parse(personID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, ErrInvalidPersonID)
	}

	person, err := s.personStorage.GetPerson(ctx, parsedPersonID)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			return fmt.Errorf("%s: %w", op, ErrPersonNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	if person.TreeID != parsedTreeID {
		return fmt.Errorf("%s: %w", op, ErrTreeMismatch)
	}

	return s.DeletePerson(ctx, personID)
}

func (s *Service) authorizeTree(ctx context.Context, requestUserID int, treeID string) (uuid.UUID, error) {
	if requestUserID <= 0 {
		return uuid.Nil, ErrInvalidUserID
	}

	parsedTreeID, err := uuid.Parse(treeID)
	if err != nil {
		return uuid.Nil, ErrInvalidTreeID
	}

	tree, err := s.personStorage.GetTree(ctx, parsedTreeID)
	if err != nil {
		if errors.Is(err, storage.ErrTreeNotFound) {
			return uuid.Nil, ErrTreeNotFound
		}
		return uuid.Nil, err
	}

	if tree.CreatorID != requestUserID {
		return uuid.Nil, ErrForbidden
	}

	return parsedTreeID, nil
}

func (s *Service) fetchIncomingParents(ctx context.Context, childID uuid.UUID) ([]models.Person, error) {
	rels, err := s.relationStorage.GetRelatives(ctx, childID)
	if err != nil {
		return nil, err
	}

	parents := make([]models.Person, 0)
	for _, rel := range rels {
		if rel.RelationshipType != models.RelationshipParentChild || rel.Direction != models.DirectionIncoming {
			continue
		}
		parent, err := s.personStorage.GetPerson(ctx, rel.Person.ID)
		if err != nil {
			return nil, err
		}
		parents = append(parents, parent)
	}

	return parents, nil
}

func (s *Service) resolveOptionalParent(ctx context.Context, treeID uuid.UUID, personID string) (models.Person, bool, error) {
	id := strings.TrimSpace(personID)
	if id == "" {
		return models.Person{}, false, nil
	}

	parsed, err := uuid.Parse(id)
	if err != nil {
		return models.Person{}, false, ErrInvalidPersonID
	}

	person, err := s.personStorage.GetPerson(ctx, parsed)
	if err != nil {
		if errors.Is(err, storage.ErrPersonNotFound) {
			return models.Person{}, false, ErrPersonNotFound
		}
		return models.Person{}, false, err
	}
	if person.TreeID != treeID {
		return models.Person{}, false, ErrTreeMismatch
	}

	return person, true, nil
}

func (s *Service) createPersonRecord(
	ctx context.Context,
	treeID uuid.UUID,
	gender models.Gender,
	firstName string,
	lastName string,
	patronymic string,
) (models.Person, error) {
	if !isValidGender(gender) {
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
		return models.Person{}, err
	}
	if err := s.relationStorage.EnsurePersonNode(ctx, person.ID, person.TreeID); err != nil {
		return models.Person{}, err
	}

	return person, nil
}

func (s *Service) ensureRelationship(ctx context.Context, fromID uuid.UUID, toID uuid.UUID, relType models.RelationshipType) error {
	if relType == models.RelationshipPartner && fromID.String() > toID.String() {
		fromID, toID = toID, fromID
	}
	if err := s.relationStorage.CreateRelationship(ctx, fromID, toID, relType); err != nil {
		if errors.Is(err, storage.ErrRelationshipExists) {
			return nil
		}
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
