package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Bad-Utya/myforebears-backend/services/emailprovider/internal/app"
	"github.com/Bad-Utya/myforebears-backend/services/emailprovider/internal/config"
	"github.com/Bad-Utya/myforebears-backend/utility/pkg/log"
)

func main() {
	cfg := config.MustLoad()

	log := log.SetupLogger(cfg.Env)

	log.Info("starting emailprovider", slog.Any("config", cfg))

	application := app.New(
		log,
		cfg.RabbitMQ.URL,
		cfg.RabbitMQ.Exchange,
		cfg.RabbitMQ.Queue,
		cfg.RabbitMQ.RoutingKey,
		cfg.RabbitMQ.ConsumerTag,
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.From,
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
