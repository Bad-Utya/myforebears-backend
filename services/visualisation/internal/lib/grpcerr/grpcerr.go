package grpcerr

import (
	"errors"

	visualisationsvc "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/services/visualisation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Map(err error) error {
	switch {
	case errors.Is(err, visualisationsvc.ErrInvalidVisualisationID),
		errors.Is(err, visualisationsvc.ErrInvalidTreeID),
		errors.Is(err, visualisationsvc.ErrInvalidRootPersonID),
		errors.Is(err, visualisationsvc.ErrInvalidIncludedPersonID),
		errors.Is(err, visualisationsvc.ErrIncludedPersonsRequired):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, visualisationsvc.ErrVisualisationNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, visualisationsvc.ErrVisualisationNotReady):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, visualisationsvc.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
