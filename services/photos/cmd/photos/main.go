package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Bad-Utya/myforebears-backend/services/photos/internal/app"
	"github.com/Bad-Utya/myforebears-backend/services/photos/internal/config"
	"github.com/Bad-Utya/myforebears-backend/utility/pkg/log"
)

func main() {
	cfg := config.MustLoad()

	logger := log.SetupLogger(cfg.Env)
	logger.Info("starting photos service", slog.Any("config", cfg))

	application := app.New(
		logger,
		cfg.GRPC.Port,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Username,
		cfg.Postgres.Password,
		cfg.Postgres.Database,
		cfg.S3,
		cfg.FamilyTree.Address,
		cfg.FamilyTree.Timeout,
		cfg.FamilyTree.RetriesCount,
		cfg.Events.Address,
		cfg.Events.Timeout,
		cfg.Events.RetriesCount,
	)

	go application.GRPCServer.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	logger.Info("stopping photos service", slog.String("signal", sign.String()))

	application.Stop()
	logger.Info("photos service stopped")
}
