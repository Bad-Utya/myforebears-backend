package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	authclient "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/clients/auth"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/config"
	authhandler "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/handlers/auth"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/middleware"
)

type App struct {
	log        *slog.Logger
	httpServer *http.Server
	authClient *authclient.Client
}

func New(log *slog.Logger, cfg *config.Config) *App {
	ctx := context.Background()

	// Connect to auth gRPC service.
	authGRPC, err := authclient.New(
		ctx,
		log,
		cfg.Clients.Auth.Address,
		cfg.Clients.Auth.Timeout,
		cfg.Clients.Auth.RetriesCount,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to auth service: %v", err))
	}

	// Build HTTP router.
	router := chi.NewRouter()

	// Global middleware.
	router.Use(chimw.RequestID)
	router.Use(chimw.RealIP)
	router.Use(middleware.Logging(log))
	router.Use(chimw.Recoverer)

	// Auth routes.
	authHandler := authhandler.New(log, authGRPC)

	router.Route("/api/auth", func(r chi.Router) {
		r.Post("/send-code", authHandler.SendCode)
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/send-link-for-reset-password", authHandler.SendLinkForResetPassword)
		r.Post("/reset-password-with-link", authHandler.ResetPasswordWithLink)
		r.Post("/reset-password-with-token", authHandler.ResetPasswordWithToken)
		r.Post("/refresh", authHandler.RefreshTokens)
		r.Post("/logout", authHandler.Logout)
		r.Post("/logout-all", authHandler.LogoutFromAllDevices)
	})

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:      router,
		ReadTimeout:  cfg.HTTP.Timeout,
		WriteTimeout: cfg.HTTP.Timeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	return &App{
		log:        log,
		httpServer: httpServer,
		authClient: authGRPC,
	}
}

func (a *App) Run() error {
	const op = "app.Run"

	a.log.Info("http server is running", slog.String("addr", a.httpServer.Addr))

	if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Stop(ctx context.Context) {
	const op = "app.Stop"

	a.log.Info("stopping http server")

	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.log.Error("failed to shutdown http server", slog.String("op", op), slog.String("error", err.Error()))
	}

	if err := a.authClient.Close(); err != nil {
		a.log.Error("failed to close auth grpc connection", slog.String("op", op), slog.String("error", err.Error()))
	}
}
