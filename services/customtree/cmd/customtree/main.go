package main

import (
	"github.com/Bad-Utya/myforebears-backend/services/customtree/internal/app"
	"github.com/Bad-Utya/myforebears-backend/services/customtree/internal/config"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	sharedlog "utility/pkg/log"
)

func main() {
	c := config.MustLoad()
	log := sharedlog.SetupLogger(c.Env)
	a := app.New(c)
	go a.MustRun()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	sig := <-ch
	log.Info("stopping customtree", slog.String("signal", sig.String()))
	a.Stop()
}
