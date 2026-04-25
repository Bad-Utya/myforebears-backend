package grpcapp

import (
	"fmt"
	"log/slog"
	"net"

	visualisationhandler "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/handlers/visualisation"
	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *slog.Logger, visualisationService visualisationhandler.VisualisationService, port int) *App {
	gRPCServer := grpc.NewServer()
	visualisationhandler.Register(gRPCServer, visualisationService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"
	log := a.log.With(slog.String("op", op), slog.Int("port", a.port))

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("grpc server is running", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).Info("stopping grpc server", slog.Int("port", a.port))
	a.gRPCServer.GracefulStop()
}
