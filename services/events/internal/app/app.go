package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/Bad-Utya/myforebears-backend/services/events/internal/app/grpc"
	familytreeclient "github.com/Bad-Utya/myforebears-backend/services/events/internal/clients/familytree"
	eventssvc "github.com/Bad-Utya/myforebears-backend/services/events/internal/services/events"
	"github.com/Bad-Utya/myforebears-backend/services/events/internal/storage/postgres"
)

type App struct {
	GRPCServer *grpcapp.App
	postgres   *postgres.Storage
	familyTree *familytreeclient.Client
}

func New(
	log *slog.Logger,
	grpcPort int,
	pgHost string,
	pgPort int,
	pgUser string,
	pgPassword string,
	pgDBName string,
	familyTreeAddr string,
	familyTreeTimeout time.Duration,
	familyTreeRetries int,
) *App {
	eventsStorage, err := postgres.New(pgHost, pgPort, pgUser, pgPassword, pgDBName)
	if err != nil {
		panic(err)
	}

	familyTree, err := familytreeclient.New(log, familyTreeAddr, familyTreeTimeout, familyTreeRetries)
	if err != nil {
		panic(err)
	}

	eventsService := eventssvc.New(log, eventsStorage, familyTree)
	grpcServer := grpcapp.New(log, eventsService, grpcPort)

	return &App{
		GRPCServer: grpcServer,
		postgres:   eventsStorage,
		familyTree: familyTree,
	}
}

func (a *App) Stop() {
	a.GRPCServer.Stop()
	a.postgres.Close()
	_ = a.familyTree.Close()
}
