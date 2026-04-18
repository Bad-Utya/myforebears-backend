package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"utility/pkg/log"

	"github.com/Bad-Utya/myforebears-backend/services/person/internal/app"
	"github.com/Bad-Utya/myforebears-backend/services/person/internal/config"
)

func main() {
	cfg := config.MustLoad()

	log := log.SetupLogger(cfg.Env)
	log.Info("starting person service", slog.Any("config", cfg))

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
	)

	go application.GRPCServer.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	log.Info("stopping person service", slog.String("signal", sign.String()))

	application.Stop()
	log.Info("person service stopped")
}
