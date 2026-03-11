package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/app"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/config"
	"github.com/Bad-Utya/myforebears-backend/utility/pkg/log"
)

func main() {
	cfg := config.MustLoad()

	log := log.SetupLogger(cfg.Env)

	log.Info("starting api gateway", slog.String("env", cfg.Env))

	application := app.New(log, cfg)

	// Run HTTP server in background.
	go application.MustRun()

	// Block until a termination signal is received.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	log.Info("stopping api gateway", slog.String("signal", sign.String()))

	// Give the server a grace period to finish active requests.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	application.Stop(ctx)

	log.Info("api gateway stopped")
}
