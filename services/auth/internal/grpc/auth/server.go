package auth

import (
	"context"
	"errors"

	authpb "github.com/Bad-Utya/myforebears-backend/gen/go/auth"
	"github.com/Bad-Utya/myforebears-backend/services/auth/internal/domain/models"
	"github.com/Bad-Utya/myforebears-backend/services/auth/internal/services/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	SendCode(ctx context.Context, email string, password string) error
	Register(ctx context.Context, email string, code string) (string, string, error)
	Login(ctx context.Context, email string, password string) (string, string, error)
	SendLinkForResetPassword(ctx context.Context, email string) error
	ResetPasswordWithLink(ctx context.Context, link string, password string) error
	ResetPasswordWithToken(ctx context.Context, accessToken string, password string) error
	RefreshTokens(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, accessToken string) error
	LogoutFromAllDevices(ctx context.Context, accessToken string) error
	GetUserInfo(ctx context.Context, userID int) (models.User, error)
	UpdateNickname(ctx context.Context, userID int, nickname string) (models.User, error)
}

type ServerAPI struct {
	authpb.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	authpb.RegisterAuthServer(gRPC, &ServerAPI{auth: auth})
}

func (s *ServerAPI) SendCode(ctx context.Context, req *authpb.SendCodeRequest) (*authpb.SendCodeResponse, error) {
	err := s.auth.SendCode(ctx, req.GetEmail(), req.GetPassword())
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

	// Code is delivered via email; the response field is intentionally empty.
	return &authpb.SendCodeResponse{}, nil
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

func (s *ServerAPI) SendLinkForResetPassword(ctx context.Context, req *authpb.SendLinkForResetPasswordRequest) (*authpb.SendLinkForResetPasswordResponse, error) {
	err := s.auth.SendLinkForResetPassword(ctx, req.GetEmail())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authpb.SendLinkForResetPasswordResponse{}, nil
}

func (s *ServerAPI) ResetPasswordWithLink(ctx context.Context, req *authpb.ResetPasswordWithLinkRequest) (*authpb.ResetPasswordWithLinkResponse, error) {
	err := s.auth.ResetPasswordWithLink(ctx, req.GetLink(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidLink) {
			return nil, status.Error(codes.InvalidArgument, "invalid link")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authpb.ResetPasswordWithLinkResponse{}, nil
}

func (s *ServerAPI) ResetPasswordWithToken(ctx context.Context, req *authpb.ResetPasswordWithTokenRequest) (*authpb.ResetPasswordWithTokenResponse, error) {
	err := s.auth.ResetPasswordWithToken(ctx, req.GetAccessToken(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) {
			return nil, status.Error(codes.InvalidArgument, "invalid token")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authpb.ResetPasswordWithTokenResponse{}, nil
}

func (s *ServerAPI) RefreshTokens(ctx context.Context, req *authpb.RefreshTokensRequest) (*authpb.RefreshTokensResponse, error) {
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

func (s *ServerAPI) GetUserInfo(ctx context.Context, req *authpb.GetUserInfoRequest) (*authpb.GetUserInfoResponse, error) {
	user, err := s.auth.GetUserInfo(ctx, int(req.GetUserId()))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidUserID) {
			return nil, status.Error(codes.InvalidArgument, "invalid user id")
		}
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authpb.GetUserInfoResponse{User: toProtoUser(user)}, nil
}

func (s *ServerAPI) UpdateNickname(ctx context.Context, req *authpb.UpdateNicknameRequest) (*authpb.UpdateNicknameResponse, error) {
	user, err := s.auth.UpdateNickname(ctx, int(req.GetUserId()), req.GetNickname())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidUserID) {
			return nil, status.Error(codes.InvalidArgument, "invalid user id")
		}
		if errors.Is(err, auth.ErrInvalidNickname) {
			return nil, status.Error(codes.InvalidArgument, "invalid nickname")
		}
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authpb.UpdateNicknameResponse{User: toProtoUser(user)}, nil
}

func toProtoUser(user models.User) *authpb.User {
	return &authpb.User{
		Id:       int32(user.ID),
		Nickname: user.Nickname,
	}
}

func (s *ServerAPI) Logout(ctx context.Context, req *authpb.LogoutRequest) (*authpb.LogoutResponse, error) {
	err := s.auth.Logout(ctx, req.GetAccessToken())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) {
			return nil, status.Error(codes.InvalidArgument, "invalid token")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authpb.LogoutResponse{}, nil
}

func (s *ServerAPI) LogoutFromAllDevices(ctx context.Context, req *authpb.LogoutFromAllDevicesRequest) (*authpb.LogoutFromAllDevicesResponse, error) {
	err := s.auth.LogoutFromAllDevices(ctx, req.GetAccessToken())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) {
			return nil, status.Error(codes.InvalidArgument, "invalid token")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authpb.LogoutFromAllDevicesResponse{}, nil
}
