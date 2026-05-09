package grpcerr

import (
	"errors"

	photossvc "github.com/Bad-Utya/myforebears-backend/services/photos/internal/services/photos"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Map(err error) error {
	switch {
	case errors.Is(err, photossvc.ErrInvalidUserID),
		errors.Is(err, photossvc.ErrInvalidPhotoID),
		errors.Is(err, photossvc.ErrInvalidTreeID),
		errors.Is(err, photossvc.ErrInvalidPersonID),
		errors.Is(err, photossvc.ErrInvalidEventID),
		errors.Is(err, photossvc.ErrInvalidFileName),
		errors.Is(err, photossvc.ErrInvalidMIMEType),
		errors.Is(err, photossvc.ErrEmptyContent),
		errors.Is(err, photossvc.ErrTooLarge):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, photossvc.ErrPhotoNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, photossvc.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
