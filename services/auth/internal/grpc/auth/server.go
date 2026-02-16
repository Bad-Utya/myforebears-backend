package auth

import (
	"context"
	"errors"

	authpb "github.com/Bad-Utya/myforebears-backend/gen/go/auth"
	"github.com/Bad-Utya/myforebears-backend/services/auth/internal/services/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	SendCode(ctx context.Context, email string, password string) (string, error)
	Register(ctx context.Context, email string, code string) (string, string, error)
	Login(ctx context.Context, email string, password string) (string, string, error)
	RefreshTokens(ctx context.Context, refreshToken string) (string, string, error)
}

type ServerAPI struct {
	authpb.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	authpb.RegisterAuthServer(gRPC, &ServerAPI{auth: auth})
}

func (s *ServerAPI) SendCode(ctx context.Context, req *authpb.SendCodeRequest) (*authpb.SendCodeResponse, error) {
	code, err := s.auth.SendCode(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		if errors.Is(err, auth.ErrCodeNotFound) {
			return nil, status.Error(codes.NotFound, "no pending verification found")
		}
		if errors.Is(err, auth.ErrCodeCooldown) {
			return nil, status.Error(codes.FailedPrecondition, "please wait before requesting a new code")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authpb.SendCodeResponse{Code: code}, nil
}

func (s *ServerAPI) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	accessToken, refreshToken, err := s.auth.Register(ctx, req.GetEmail(), req.GetCode())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCode) {
			return nil, status.Error(codes.InvalidArgument, "invalid verification code")
		}
		if errors.Is(err, auth.ErrNoAttemptsLeft) {
			return nil, status.Error(codes.ResourceExhausted, "no attempts left, please request a new code")
		}
		if errors.Is(err, auth.ErrCodeNotFound) {
			return nil, status.Error(codes.NotFound, "no pending verification found")
		}
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authpb.RegisterResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *ServerAPI) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	accessToken, refreshToken, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authpb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *ServerAPI) RefreshToken(ctx context.Context, req *authpb.RefreshTokensRequest) (*authpb.RefreshTokensResponse, error) {
	accessToken, refreshToken, err := s.auth.RefreshTokens(ctx, req.GetRefreshToken())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) {
			return nil, status.Error(codes.InvalidArgument, "invalid token")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authpb.RefreshTokensResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
