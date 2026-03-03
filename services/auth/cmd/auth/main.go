package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Bad-Utya/myforebears-backend/services/auth/internal/app"
	"github.com/Bad-Utya/myforebears-backend/services/auth/internal/config"
	"github.com/Bad-Utya/myforebears-backend/utility/pkg/log"
)

func main() {
	cfg := config.MustLoad()

	log := log.SetupLogger(cfg.Env)

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
		cfg.RabbitMQ.URL,
		cfg.RabbitMQ.Exchange,
		cfg.RabbitMQ.RoutingKey,
	)

	// Run gRPC server in background; shutdown is handled by OS signals.
	go application.GRPCServer.MustRun()

	// Block until a termination signal is received.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	log.Info("stopping app", slog.String("signal", sign.String()))

	application.Stop()

	log.Info("app stopped")
}
