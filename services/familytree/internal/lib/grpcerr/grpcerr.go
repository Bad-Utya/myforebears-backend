package grpcerr

import (
	"errors"

	personsvc "github.com/Bad-Utya/myforebears-backend/services/familytree/internal/services/familytree"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Map(err error) error {
	switch {
	case errors.Is(err, personsvc.ErrInvalidPersonID),
		errors.Is(err, personsvc.ErrInvalidTreeID),
		errors.Is(err, personsvc.ErrInvalidUserID),
		errors.Is(err, personsvc.ErrInvalidName),
		errors.Is(err, personsvc.ErrInvalidGender),
		errors.Is(err, personsvc.ErrDeleteNotAllowed),
		errors.Is(err, personsvc.ErrCannotDeleteLast),
		errors.Is(err, personsvc.ErrInvalidParentRole),
		errors.Is(err, personsvc.ErrParentExists),
		errors.Is(err, personsvc.ErrParentLimitReached),
		errors.Is(err, personsvc.ErrAtLeastOneParent),
		errors.Is(err, personsvc.ErrTreeMismatch),
		errors.Is(err, personsvc.ErrUnknownPersonGender),
		errors.Is(err, personsvc.ErrInvalidRelationType),
		errors.Is(err, personsvc.ErrSelfRelationship),
		errors.Is(err, personsvc.ErrPersonNotInSameTree):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, personsvc.ErrPersonNotFound), errors.Is(err, personsvc.ErrRelationshipMissing), errors.Is(err, personsvc.ErrTreeNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, personsvc.ErrRelationshipExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, personsvc.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
