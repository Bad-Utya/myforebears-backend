package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"utility/pkg/log"

	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/app"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/config"
)

func main() {
	cfg := config.MustLoad()

	log := log.SetupLogger(cfg.Env)
	log.Info("starting familytree service", slog.Any("config", cfg))

	application := app.New(
		log,
		cfg.GRPC.Port,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Username,
		cfg.Postgres.Password,
		cfg.Postgres.Database,
		cfg.Neo4j.URI,
		cfg.Neo4j.Username,
		cfg.Neo4j.Password,
		cfg.Events.Address,
		cfg.Events.Timeout,
		cfg.Events.RetriesCount,
	)

	go application.GRPCServer.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	log.Info("stopping familytree service", slog.String("signal", sign.String()))

	application.Stop()
	log.Info("familytree service stopped")
}
