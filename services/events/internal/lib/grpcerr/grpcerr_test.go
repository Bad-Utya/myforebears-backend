package grpcerr_test

import (
	"errors"
	"testing"

	grpcerrpkg "github.com/Bad-Utya/myforebears-backend/services/events/internal/lib/grpcerr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMap_DefaultReturnsInternal(t *testing.T) {
	// passing an unrelated error should map to Internal with generic message
	out := grpcerrpkg.Map(errors.New("boom"))
	st, ok := status.FromError(out)
	if !ok {
		t.Fatalf("expected gRPC status error, got %T", out)
	}
	if st.Code() != codes.Internal || st.Message() != "internal error" {
		t.Fatalf("unexpected status: %v %q", st.Code(), st.Message())
	}
}
