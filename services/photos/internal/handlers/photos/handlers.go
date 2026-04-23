package photos

import (
	"context"

	photospb "github.com/Bad-Utya/myforebears-backend/gen/go/photos"
	"github.com/Bad-Utya/myforebears-backend/services/photos/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/photos/internal/lib/grpcerr"
	"google.golang.org/grpc"
)

type PhotosService interface {
	UploadUserAvatar(ctx context.Context, requestUserID int, fileName string, mimeType string, content []byte) (models.Photo, error)
	GetUserAvatar(ctx context.Context, requestUserID int) (models.Photo, []byte, error)
	UploadPersonAvatar(ctx context.Context, requestUserID int, personID string, fileName string, mimeType string, content []byte) (models.Photo, error)
	GetPersonAvatar(ctx context.Context, requestUserID int, personID string) (models.Photo, []byte, error)
	UploadPersonPhoto(ctx context.Context, requestUserID int, personID string, fileName string, mimeType string, content []byte) (models.Photo, error)
	ListPersonPhotos(ctx context.Context, requestUserID int, personID string) ([]models.Photo, error)
	UploadEventPhoto(ctx context.Context, requestUserID int, eventID string, fileName string, mimeType string, content []byte) (models.Photo, error)
	ListEventPhotos(ctx context.Context, requestUserID int, eventID string) ([]models.Photo, error)
	GetPhotoByID(ctx context.Context, requestUserID int, photoID string) (models.Photo, []byte, error)
	DeletePhotoByID(ctx context.Context, requestUserID int, photoID string) error
}

type Handler struct {
	photospb.UnimplementedPhotosServiceServer
	service PhotosService
}

func New(service PhotosService) *Handler {
	return &Handler{service: service}
}

func Register(gRPC *grpc.Server, service PhotosService) {
	photospb.RegisterPhotosServiceServer(gRPC, New(service))
}

func (h *Handler) UploadUserAvatar(ctx context.Context, req *photospb.UploadUserAvatarRequest) (*photospb.UploadUserAvatarResponse, error) {
	photo, err := h.service.UploadUserAvatar(
		ctx,
		int(req.GetRequestUserId()),
		req.GetFileName(),
		req.GetMimeType(),
		req.GetContent(),
	)
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &photospb.UploadUserAvatarResponse{Photo: toProtoPhoto(photo)}, nil
}

func (h *Handler) GetUserAvatar(ctx context.Context, req *photospb.GetUserAvatarRequest) (*photospb.GetPhotoContentResponse, error) {
	photo, content, err := h.service.GetUserAvatar(ctx, int(req.GetRequestUserId()))
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &photospb.GetPhotoContentResponse{Photo: toProtoPhoto(photo), Content: content}, nil
}

func (h *Handler) UploadPersonAvatar(ctx context.Context, req *photospb.UploadPersonAvatarRequest) (*photospb.UploadPersonAvatarResponse, error) {
	photo, err := h.service.UploadPersonAvatar(
		ctx,
		int(req.GetRequestUserId()),
		req.GetPersonId(),
		req.GetFileName(),
		req.GetMimeType(),
		req.GetContent(),
	)
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &photospb.UploadPersonAvatarResponse{Photo: toProtoPhoto(photo)}, nil
}

func (h *Handler) GetPersonAvatar(ctx context.Context, req *photospb.GetPersonAvatarRequest) (*photospb.GetPhotoContentResponse, error) {
	photo, content, err := h.service.GetPersonAvatar(ctx, int(req.GetRequestUserId()), req.GetPersonId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &photospb.GetPhotoContentResponse{Photo: toProtoPhoto(photo), Content: content}, nil
}

func (h *Handler) UploadPersonPhoto(ctx context.Context, req *photospb.UploadPersonPhotoRequest) (*photospb.UploadPersonPhotoResponse, error) {
	photo, err := h.service.UploadPersonPhoto(
		ctx,
		int(req.GetRequestUserId()),
		req.GetPersonId(),
		req.GetFileName(),
		req.GetMimeType(),
		req.GetContent(),
	)
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &photospb.UploadPersonPhotoResponse{Photo: toProtoPhoto(photo)}, nil
}

func (h *Handler) ListPersonPhotos(ctx context.Context, req *photospb.ListPersonPhotosRequest) (*photospb.ListPersonPhotosResponse, error) {
	photos, err := h.service.ListPersonPhotos(
		ctx,
		int(req.GetRequestUserId()),
		req.GetPersonId(),
	)
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	out := make([]*photospb.Photo, 0, len(photos))
	for _, photo := range photos {
		out = append(out, toProtoPhoto(photo))
	}

	return &photospb.ListPersonPhotosResponse{Photos: out}, nil
}

func (h *Handler) UploadEventPhoto(ctx context.Context, req *photospb.UploadEventPhotoRequest) (*photospb.UploadEventPhotoResponse, error) {
	photo, err := h.service.UploadEventPhoto(
		ctx,
		int(req.GetRequestUserId()),
		req.GetEventId(),
		req.GetFileName(),
		req.GetMimeType(),
		req.GetContent(),
	)
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &photospb.UploadEventPhotoResponse{Photo: toProtoPhoto(photo)}, nil
}

func (h *Handler) ListEventPhotos(ctx context.Context, req *photospb.ListEventPhotosRequest) (*photospb.ListEventPhotosResponse, error) {
	photos, err := h.service.ListEventPhotos(
		ctx,
		int(req.GetRequestUserId()),
		req.GetEventId(),
	)
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	out := make([]*photospb.Photo, 0, len(photos))
	for _, photo := range photos {
		out = append(out, toProtoPhoto(photo))
	}

	return &photospb.ListEventPhotosResponse{Photos: out}, nil
}

func (h *Handler) GetPhotoByID(ctx context.Context, req *photospb.GetPhotoByIDRequest) (*photospb.GetPhotoContentResponse, error) {
	photo, content, err := h.service.GetPhotoByID(ctx, int(req.GetRequestUserId()), req.GetPhotoId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &photospb.GetPhotoContentResponse{Photo: toProtoPhoto(photo), Content: content}, nil
}

func (h *Handler) DeletePhotoByID(ctx context.Context, req *photospb.DeletePhotoByIDRequest) (*photospb.DeletePhotoByIDResponse, error) {
	err := h.service.DeletePhotoByID(ctx, int(req.GetRequestUserId()), req.GetPhotoId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}

	return &photospb.DeletePhotoByIDResponse{}, nil
}

func toProtoPhoto(photo models.Photo) *photospb.Photo {
	personID := ""
	eventID := ""
	treeID := ""
	if photo.PersonID != nil {
		personID = photo.PersonID.String()
	}
	if photo.EventID != nil {
		eventID = photo.EventID.String()
	}
	if photo.TreeID != nil {
		treeID = photo.TreeID.String()
	}

	return &photospb.Photo{
		Id:             photo.ID.String(),
		OwnerUserId:    int32(photo.OwnerUserID),
		PersonId:       personID,
		EventId:        eventID,
		IsUserAvatar:   photo.IsUserAvatar,
		IsPersonAvatar: photo.IsPersonAvatar,
		FileName:       photo.FileName,
		MimeType:       photo.MIMEType,
		SizeBytes:      photo.SizeBytes,
		CreatedAtUnix:  photo.CreatedAt.Unix(),
		TreeId:         treeID,
	}
}
