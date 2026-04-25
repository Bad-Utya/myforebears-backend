package familytree

import (
	"context"

	familytreepb "github.com/Bad-Utya/myforebears-backend/gen/go/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/lib/grpcerr"
	personsvc "github.com/Bad-Utya/myforebears-backend/services/familytree/internal/services/familytree"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type FamilyTreeService interface {
	CreateTree(ctx context.Context, requestUserID int) (models.Tree, models.Person, error)
	ListTreesByCreator(ctx context.Context, requestUserID int) ([]models.Tree, error)
	ListPublicTreesByCreator(ctx context.Context, creatorID int) ([]models.Tree, error)
	ListRandomPublicTrees(ctx context.Context, limit int) ([]models.Tree, error)
	GetTree(ctx context.Context, treeID string) (models.Tree, error)
	GetTreeContent(ctx context.Context, treeID string) ([]models.Person, []models.Relationship, error)
	GetTreeAccessInfo(ctx context.Context, treeID string) (models.Tree, error)
	IsTreeAccessEmailAllowed(ctx context.Context, treeID string, email string) (bool, error)
	AddTreeAccessEmail(ctx context.Context, treeID string, email string) error
	ListTreeAccessEmails(ctx context.Context, treeID string) ([]string, error)
	DeleteTreeAccessEmail(ctx context.Context, treeID string, email string) error
	UpdateTreeSettings(ctx context.Context, treeID string, isViewRestricted bool, isPublicOnMainPage bool, name string) (models.Tree, error)
	ListPersonsByTree(ctx context.Context, treeID string) ([]models.Person, error)
	AddParent(ctx context.Context, treeID string, childID string, role personsvc.ParentRole, firstName string, lastName string, patronymic string) (models.Person, *models.Person, error)
	AddChild(ctx context.Context, treeID string, parent1ID string, parent2ID string, firstName string, lastName string, patronymic string, gender models.Gender) (models.Person, *models.Person, error)
	AddPartner(ctx context.Context, treeID string, personID string, firstName string, lastName string, patronymic string) (models.Person, error)
	UpdatePersonName(ctx context.Context, treeID string, personID string, firstName string, lastName string, patronymic string) (models.Person, error)
	UpdatePersonAvatarPhoto(ctx context.Context, personID string, avatarPhotoID string) (models.Person, error)
	DeletePersonInTree(ctx context.Context, treeID string, personID string) error

	CreatePerson(ctx context.Context, treeID string, firstName string, lastName string, patronymic string, gender models.Gender) (models.Person, error)
	GetPerson(ctx context.Context, treeID string, personID string) (models.Person, error)
	UpdatePerson(ctx context.Context, treeID string, personID string, firstName string, lastName string, patronymic string, gender models.Gender) (models.Person, error)
	DeletePerson(ctx context.Context, treeID string, personID string) error
	AddRelationship(ctx context.Context, treeID string, personIDFrom string, personIDTo string, relType models.RelationshipType) error
	RemoveRelationship(ctx context.Context, treeID string, personIDFrom string, personIDTo string, relType models.RelationshipType) error
	GetRelatives(ctx context.Context, treeID string, personID string) ([]models.Relative, error)
	ValidatePersonsInTree(ctx context.Context, treeID string, personIDs []string) error
	UpdatePartnerRelationshipStatus(ctx context.Context, treeID string, personID1 string, personID2 string, status models.PartnerRelationshipStatus) error
}

type Handler struct {
	familytreepb.UnimplementedFamilyTreeServiceServer
	service FamilyTreeService
}

func New(service FamilyTreeService) *Handler {
	return &Handler{service: service}
}

func Register(gRPC *grpc.Server, service FamilyTreeService) {
	familytreepb.RegisterFamilyTreeServiceServer(gRPC, New(service))
}

func (h *Handler) CreatePerson(ctx context.Context, req *familytreepb.CreatePersonRequest) (*familytreepb.CreatePersonResponse, error) {
	person, err := h.service.CreatePerson(
		ctx,
		req.GetTreeId(),
		req.GetFirstName(),
		req.GetLastName(),
		req.GetPatronymic(),
		toModelGender(req.GetGender()),
	)
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.CreatePersonResponse{Person: toProtoPerson(person)}, nil
}

func (h *Handler) GetPerson(ctx context.Context, req *familytreepb.GetPersonRequest) (*familytreepb.GetPersonResponse, error) {
	person, err := h.service.GetPerson(ctx, req.GetTreeId(), req.GetPersonId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.GetPersonResponse{Person: toProtoPerson(person)}, nil
}

func (h *Handler) UpdatePerson(ctx context.Context, req *familytreepb.UpdatePersonRequest) (*familytreepb.UpdatePersonResponse, error) {
	person, err := h.service.UpdatePerson(
		ctx,
		req.GetTreeId(),
		req.GetPersonId(),
		req.GetFirstName(),
		req.GetLastName(),
		req.GetPatronymic(),
		toModelGender(req.GetGender()),
	)
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.UpdatePersonResponse{Person: toProtoPerson(person)}, nil
}

func (h *Handler) DeletePerson(ctx context.Context, req *familytreepb.DeletePersonRequest) (*familytreepb.DeletePersonResponse, error) {
	err := h.service.DeletePerson(ctx, req.GetTreeId(), req.GetPersonId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &familytreepb.DeletePersonResponse{}, nil
}

func toProtoPerson(person models.Person) *familytreepb.Person {
	return &familytreepb.Person{
		Id:            person.ID.String(),
		TreeId:        person.TreeID.String(),
		FirstName:     person.FirstName,
		LastName:      person.LastName,
		Patronymic:    person.Patronymic,
		Gender:        toProtoGender(person.Gender),
		AvatarPhotoId: toProtoAvatarPhotoID(person.AvatarPhotoID),
	}
}

func toProtoAvatarPhotoID(avatarPhotoID *uuid.UUID) string {
	if avatarPhotoID == nil {
		return ""
	}

	return avatarPhotoID.String()
}

func toProtoGender(gender models.Gender) familytreepb.Gender {
	switch gender {
	case models.GenderMale:
		return familytreepb.Gender_GENDER_MALE
	case models.GenderFemale:
		return familytreepb.Gender_GENDER_FEMALE
	default:
		return familytreepb.Gender_GENDER_UNSPECIFIED
	}
}

func toModelGender(gender familytreepb.Gender) models.Gender {
	switch gender {
	case familytreepb.Gender_GENDER_MALE:
		return models.GenderMale
	case familytreepb.Gender_GENDER_FEMALE:
		return models.GenderFemale
	default:
		return ""
	}
}
