package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Bad-Utya/myforebears-backend/services/emailprovider/internal/app"
	"github.com/Bad-Utya/myforebears-backend/services/emailprovider/internal/config"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting emailprovider", slog.Any("config", cfg))

	application := app.New(
		log,
		cfg.RabbitMQ.URL,
		cfg.RabbitMQ.Exchange,
		cfg.RabbitMQ.Queue,
		cfg.RabbitMQ.RoutingKey,
		cfg.RabbitMQ.ConsumerTag,
	)

	// Run consumer in background goroutine.
	go application.MustRun()

	// Block until OS termination signal.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	log.Info("stopping emailprovider")
	application.Stop()
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
