package grpcerr

import (
	"errors"

	eventssvc "github.com/Bad-Utya/myforebears-backend/services/events/internal/services/events"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Map(err error) error {
	switch {
	case errors.Is(err, eventssvc.ErrInvalidEventTypeID),
		errors.Is(err, eventssvc.ErrInvalidEventID),
		errors.Is(err, eventssvc.ErrInvalidTreeID),
		errors.Is(err, eventssvc.ErrInvalidUserID),
		errors.Is(err, eventssvc.ErrInvalidEventTypeName),
		errors.Is(err, eventssvc.ErrInvalidPrimaryPersonsMode),
		errors.Is(err, eventssvc.ErrInvalidPrimaryPersonsCount),
		errors.Is(err, eventssvc.ErrInvalidEventDate),
		errors.Is(err, eventssvc.ErrInvalidEventDatePrecision),
		errors.Is(err, eventssvc.ErrInvalidEventDateBound),
		errors.Is(err, eventssvc.ErrInvalidPrimaryPersons),
		errors.Is(err, eventssvc.ErrDuplicatePersonInParticipants),
		errors.Is(err, eventssvc.ErrParticipantListsOverlap):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, eventssvc.ErrEventTypeNotFound),
		errors.Is(err, eventssvc.ErrEventNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, eventssvc.ErrEventTypeAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, eventssvc.ErrEventTypeInUse),
		errors.Is(err, eventssvc.ErrCannotDeleteSystemEventType):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, eventssvc.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
