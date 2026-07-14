package photos

import (
	"context"

	photospb "github.com/Bad-Utya/myforebears-backend/gen/go/photos"
	"github.com/Bad-Utya/myforebears-backend/services/photos/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/photos/internal/lib/grpcerr"
	photossvc "github.com/Bad-Utya/myforebears-backend/services/photos/internal/services/photos"
	"github.com/google/uuid"
)

func (h *Handler) CopyPersonMediaToPublic(ctx context.Context, req *photospb.CopyPersonMediaToPublicRequest) (*photospb.CopyMediaResponse, error) {
	photos, err := h.service.CopyPersonMediaToPublic(ctx, int(req.GetRequestUserId()), req.GetTreeId(), req.GetPersonId(), req.GetPublicPersonId(), toMappings(req.GetEventMappings()))
	if err != nil {
		return nil, grpcerr.Map(err)
	}
	return mediaResponse(photos), nil
}
func (h *Handler) CopyPublicPersonMediaToTree(ctx context.Context, req *photospb.CopyPublicPersonMediaToTreeRequest) (*photospb.CopyMediaResponse, error) {
	photos, err := h.service.CopyPublicPersonMediaToTree(ctx, int(req.GetRequestUserId()), req.GetPublicPersonId(), req.GetTreeId(), req.GetPersonId(), toMappings(req.GetEventMappings()))
	if err != nil {
		return nil, grpcerr.Map(err)
	}
	return mediaResponse(photos), nil
}
func (h *Handler) UploadPublicPersonPhoto(ctx context.Context, req *photospb.UploadPublicPersonPhotoRequest) (*photospb.UploadPersonPhotoResponse, error) {
	p, err := h.service.UploadPublicPersonPhoto(ctx, int(req.GetRequestUserId()), req.GetPublicPersonId(), req.GetFileName(), req.GetMimeType(), req.GetContent(), req.GetIsAvatar())
	if err != nil {
		return nil, grpcerr.Map(err)
	}
	return &photospb.UploadPersonPhotoResponse{Photo: toProtoPhoto(p)}, nil
}
func (h *Handler) ListPublicPersonPhotos(ctx context.Context, req *photospb.ListPublicPersonPhotosRequest) (*photospb.ListPersonPhotosResponse, error) {
	items, err := h.service.ListPublicPersonPhotos(ctx, req.GetPublicPersonId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}
	out := make([]*photospb.Photo, 0, len(items))
	for _, p := range items {
		out = append(out, toProtoPhoto(p))
	}
	return &photospb.ListPersonPhotosResponse{Photos: out}, nil
}
func (h *Handler) GetPublicPersonPhoto(ctx context.Context, req *photospb.GetPublicPersonPhotoRequest) (*photospb.GetPhotoContentResponse, error) {
	p, data, err := h.service.GetPublicPersonPhoto(ctx, req.GetPublicPersonId(), req.GetPhotoId())
	if err != nil {
		return nil, grpcerr.Map(err)
	}
	return &photospb.GetPhotoContentResponse{Photo: toProtoPhoto(p), Content: data}, nil
}
func (h *Handler) DeletePublicPersonPhoto(ctx context.Context, req *photospb.DeletePublicPersonPhotoRequest) (*photospb.DeletePhotoByIDResponse, error) {
	if err := h.service.DeletePublicPersonPhoto(ctx, int(req.GetRequestUserId()), req.GetPublicPersonId(), req.GetPhotoId()); err != nil {
		return nil, grpcerr.Map(err)
	}
	return &photospb.DeletePhotoByIDResponse{}, nil
}
func (h *Handler) DeletePublicPersonMedia(ctx context.Context, req *photospb.DeletePublicPersonMediaRequest) (*photospb.DeletePhotoByIDResponse, error) {
	if err := h.service.DeletePublicPersonMedia(ctx, int(req.GetRequestUserId()), req.GetPublicPersonId()); err != nil {
		return nil, grpcerr.Map(err)
	}
	return &photospb.DeletePhotoByIDResponse{}, nil
}

func toMappings(items []*photospb.EventPhotoMapping) []photossvc.EventPhotoMapping {
	out := make([]photossvc.EventPhotoMapping, 0, len(items))
	for _, m := range items {
		source, e1 := uuid.Parse(m.GetSourceEventId())
		target, e2 := uuid.Parse(m.GetTargetEventId())
		if e1 == nil && e2 == nil {
			out = append(out, photossvc.EventPhotoMapping{SourceEventID: source, TargetEventID: target})
		}
	}
	return out
}
func mediaResponse(items []models.Photo) *photospb.CopyMediaResponse {
	out := make([]*photospb.Photo, 0, len(items))
	for _, p := range items {
		out = append(out, toProtoPhoto(p))
	}
	return &photospb.CopyMediaResponse{Photos: out}
}
