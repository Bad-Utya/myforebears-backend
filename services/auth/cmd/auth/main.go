package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Bad-Utya/myforebears-backend/services/auth/internal/app"
	"github.com/Bad-Utya/myforebears-backend/services/auth/internal/config"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting app", slog.Any("config", cfg))

	application := app.New(
		log,
		cfg.GRPC.Port,
		cfg.UserStorage.Host,
		cfg.UserStorage.Port,
		cfg.UserStorage.Username,
		cfg.UserStorage.Password,
		cfg.UserStorage.Database,
		cfg.VerificationStorage.Address,
		cfg.VerificationStorage.Password,
		cfg.VerificationStorage.Database,
		cfg.JWTSecret,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
		cfg.LinkForResetPassword,
		cfg.LinkTTL,
	)

	// Run gRPC server in background; shutdown is handled by OS signals.
	go application.GRPCServer.MustRun()

	// Block until a termination signal is received.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	log.Info("stopping app", slog.String("signal", sign.String()))

	application.GRPCServer.Stop()

	log.Info("app stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
