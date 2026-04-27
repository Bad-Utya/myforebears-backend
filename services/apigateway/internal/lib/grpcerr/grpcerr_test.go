package grpcerr_test

import (
	"net/http"
	"testing"

	grpcerrpkg "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/lib/grpcerr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHTTPStatus_MappedCodes(t *testing.T) {
	cases := []struct {
		code codes.Code
		want int
	}{
		{codes.InvalidArgument, http.StatusBadRequest},
		{codes.NotFound, http.StatusNotFound},
		{codes.AlreadyExists, http.StatusConflict},
		{codes.FailedPrecondition, http.StatusTooManyRequests},
		{codes.ResourceExhausted, http.StatusTooManyRequests},
		{codes.Unauthenticated, http.StatusUnauthorized},
		{codes.PermissionDenied, http.StatusForbidden},
	}

	for _, c := range cases {
		err := status.Error(c.code, "msg")
		gotCode, gotMsg := grpcerrpkg.HTTPStatus(err)
		if gotCode != c.want {
			t.Fatalf("code %v: want %d, got %d", c.code, c.want, gotCode)
		}
		if gotMsg != "msg" {
			t.Fatalf("code %v: expected message propagated, got %q", c.code, gotMsg)
		}
	}
}

func TestHTTPStatus_DefaultAndNonStatus(t *testing.T) {
	// Unknown -> default branch returns internal error message
	err := status.Error(codes.Unknown, "secret")
	gotCode, gotMsg := grpcerrpkg.HTTPStatus(err)
	if gotCode != http.StatusInternalServerError {
		t.Fatalf("unknown: want %d, got %d", http.StatusInternalServerError, gotCode)
	}
	if gotMsg != "internal error" {
		t.Fatalf("unknown: want internal error, got %q", gotMsg)
	}

	// non-gRPC error
	gotCode, gotMsg = grpcerrpkg.HTTPStatus(nil)
	if gotCode != http.StatusInternalServerError || gotMsg != "internal error" {
		t.Fatalf("nil error: want (500, 'internal error'), got (%d, %q)", gotCode, gotMsg)
	}
}
