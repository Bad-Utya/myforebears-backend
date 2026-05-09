package app

import (
	"context"
	"log/slog"
	"time"

	grpcapp "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/app/grpc"
	familytreeclient "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/clients/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/config"
	visualisationsvc "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/services/visualisation"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/storage/postgres"
	visualisations3 "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/storage/s3"
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
	s3cfg config.S3Config,
	familyTreeAddr string,
	familyTreeTimeout time.Duration,
	familyTreeRetries int,
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

	visualisationService := visualisationsvc.New(log, metaStorage, s3storage, familyTree)
	grpcServer := grpcapp.New(log, visualisationService, grpcPort)

	return &App{
		GRPCServer: grpcServer,
		postgres:   metaStorage,
		familyTree: familyTree,
	}
}

func (a *App) Stop() {
	a.GRPCServer.Stop()
	a.postgres.Close()
	_ = a.familyTree.Close()
}
