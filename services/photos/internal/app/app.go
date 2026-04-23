package app

import (
	"context"
	"log/slog"
	"time"

	grpcapp "github.com/Bad-Utya/myforebears-backend/services/photos/internal/app/grpc"
	eventsclient "github.com/Bad-Utya/myforebears-backend/services/photos/internal/clients/events"
	familytreeclient "github.com/Bad-Utya/myforebears-backend/services/photos/internal/clients/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/photos/internal/config"
	photossvc "github.com/Bad-Utya/myforebears-backend/services/photos/internal/services/photos"
	"github.com/Bad-Utya/myforebears-backend/services/photos/internal/storage/postgres"
	photos3 "github.com/Bad-Utya/myforebears-backend/services/photos/internal/storage/s3"
)

type App struct {
	GRPCServer *grpcapp.App
	postgres   *postgres.Storage
	familyTree *familytreeclient.Client
	events     *eventsclient.Client
}

func New(
	log *slog.Logger,
	grpcPort int,
	pgHost string,
	pgPort int,
	pgUser string,
	pgPassword string,
	pgDBName string,
	s3cfg config.S3Config,
	familyTreeAddr string,
	familyTreeTimeout time.Duration,
	familyTreeRetries int,
	eventsAddr string,
	eventsTimeout time.Duration,
	eventsRetries int,
) *App {
	photoStorage, err := postgres.New(pgHost, pgPort, pgUser, pgPassword, pgDBName)
	if err != nil {
		panic(err)
	}

	s3storage, err := photos3.New(context.Background(), s3cfg)
	if err != nil {
		panic(err)
	}

	familyTree, err := familytreeclient.New(log, familyTreeAddr, familyTreeTimeout, familyTreeRetries)
	if err != nil {
		panic(err)
	}

	events, err := eventsclient.New(log, eventsAddr, eventsTimeout, eventsRetries)
	if err != nil {
		panic(err)
	}

	photosService := photossvc.New(log, photoStorage, s3storage, familyTree, events)
	grpcServer := grpcapp.New(log, photosService, grpcPort)

	return &App{
		GRPCServer: grpcServer,
		postgres:   photoStorage,
		familyTree: familyTree,
		events:     events,
	}
}

func (a *App) Stop() {
	a.GRPCServer.Stop()
	a.postgres.Close()
	_ = a.familyTree.Close()
	_ = a.events.Close()
}
