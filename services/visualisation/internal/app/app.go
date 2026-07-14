package app

import (
	"context"
	"log/slog"
	"time"

	grpcapp "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/app/grpc"
	eventsclient "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/clients/events"
	familytreeclient "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/clients/familytree"
	photosclient "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/clients/photos"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/config"
	visualisationsvc "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/services/visualisation"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/storage/postgres"
	visualisations3 "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/storage/s3"
)

type App struct {
	GRPCServer *grpcapp.App
	service    *visualisationsvc.Service
	postgres   *postgres.Storage
	familyTree *familytreeclient.Client
	photos     *photosclient.Client
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
	photosAddr string,
	photosTimeout time.Duration,
	photosRetries int,
	eventsAddr string,
	eventsTimeout time.Duration,
	eventsRetries int,
) *App {
	metaStorage, err := postgres.New(pgHost, pgPort, pgUser, pgPassword, pgDBName)
	if err != nil {
		panic(err)
	}

	s3storage, err := visualisations3.New(context.Background(), s3cfg)
	if err != nil {
		panic(err)
	}

	familyTree, err := familytreeclient.New(log, familyTreeAddr, familyTreeTimeout, familyTreeRetries)
	if err != nil {
		panic(err)
	}

	photos, err := photosclient.New(log, photosAddr, photosTimeout, photosRetries)
	if err != nil {
		panic(err)
	}

	events, err := eventsclient.New(log, eventsAddr, eventsTimeout, eventsRetries)
	if err != nil {
		panic(err)
	}

	visualisationService := visualisationsvc.New(log, metaStorage, s3storage, familyTree, photos, events)
	grpcServer := grpcapp.New(log, visualisationService, grpcPort)

	return &App{
		GRPCServer: grpcServer,
		service:    visualisationService,
		postgres:   metaStorage,
		familyTree: familyTree,
		photos:     photos,
		events:     events,
	}
}

func (a *App) Stop() {
	a.GRPCServer.Stop()
	a.service.Close()
	a.postgres.Close()
	_ = a.familyTree.Close()
	_ = a.photos.Close()
	_ = a.events.Close()
}
