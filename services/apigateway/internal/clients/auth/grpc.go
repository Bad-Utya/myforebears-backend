package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	authpb "github.com/Bad-Utya/myforebears-backend/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api  authpb.AuthClient
	conn *grpc.ClientConn
	log  *slog.Logger
}

func New(ctx context.Context, log *slog.Logger, addr string, timeout time.Duration, retriesCount int) (*Client, error) {
	const op = "clients.auth.New"

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		api:  authpb.NewAuthClient(conn),
		conn: conn,
		log:  log,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) SendCode(ctx context.Context, email, password string) error {
	const op = "clients.auth.SendCode"

	_, err := c.api.SendCode(ctx, &authpb.SendCodeRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) Register(ctx context.Context, email, code string) (string, string, error) {
	const op = "clients.auth.Register"

	resp, err := c.api.Register(ctx, &authpb.RegisterRequest{
		Email: email,
		Code:  code,
	})
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return resp.GetAccessToken(), resp.GetRefreshToken(), nil
}

func (c *Client) Login(ctx context.Context, email, password string) (string, string, error) {
	const op = "clients.auth.Login"

	resp, err := c.api.Login(ctx, &authpb.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return resp.GetAccessToken(), resp.GetRefreshToken(), nil
}

func (c *Client) SendLinkForResetPassword(ctx context.Context, email string) error {
	const op = "clients.auth.SendLinkForResetPassword"

	_, err := c.api.SendLinkForResetPassword(ctx, &authpb.SendLinkForResetPasswordRequest{
		Email: email,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) ResetPasswordWithLink(ctx context.Context, link, password string) error {
	const op = "clients.auth.ResetPasswordWithLink"

	_, err := c.api.ResetPasswordWithLink(ctx, &authpb.ResetPasswordWithLinkRequest{
		Link:     link,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) ResetPasswordWithToken(ctx context.Context, accessToken, password string) error {
	const op = "clients.auth.ResetPasswordWithToken"

	_, err := c.api.ResetPasswordWithToken(ctx, &authpb.ResetPasswordWithTokenRequest{
		AccessToken: accessToken,
		Password:    password,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	const op = "clients.auth.RefreshTokens"

	resp, err := c.api.RefreshTokens(ctx, &authpb.RefreshTokensRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return resp.GetAccessToken(), resp.GetRefreshToken(), nil
}

func (c *Client) Logout(ctx context.Context, accessToken string) error {
	const op = "clients.auth.Logout"

	_, err := c.api.Logout(ctx, &authpb.LogoutRequest{
		AccessToken: accessToken,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) LogoutFromAllDevices(ctx context.Context, accessToken string) error {
	const op = "clients.auth.LogoutFromAllDevices"

	_, err := c.api.LogoutFromAllDevices(ctx, &authpb.LogoutFromAllDevicesRequest{
		AccessToken: accessToken,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
