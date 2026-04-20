package app

import (
	"context"
	"log/slog"

	grpcapp "github.com/Bad-Utya/myforebears-backend/services/familytree/internal/app/grpc"
	familytreesvc "github.com/Bad-Utya/myforebears-backend/services/familytree/internal/services/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/storage/neo4j"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/storage/postgres"
)

type App struct {
	GRPCServer *grpcapp.App
	postgres   *postgres.Storage
	neo4j      *neo4j.Storage
}

func New(
	log *slog.Logger,
	grpcPort int,
	pgHost string,
	pgPort int,
	pgUser string,
	pgPassword string,
	pgDBName string,
	neo4jURI string,
	neo4jUser string,
	neo4jPassword string,
) *App {
	personStorage, err := postgres.New(pgHost, pgPort, pgUser, pgPassword, pgDBName)
	if err != nil {
		panic(err)
	}

	relStorage, err := neo4j.New(neo4jURI, neo4jUser, neo4jPassword)
	if err != nil {
		panic(err)
	}

	personService := familytreesvc.New(log, personStorage, relStorage)
	grpcServer := grpcapp.New(log, personService, grpcPort)

	return &App{
		GRPCServer: grpcServer,
		postgres:   personStorage,
		neo4j:      relStorage,
	}
}

func (a *App) Stop() {
	a.GRPCServer.Stop()
	a.postgres.Close()
	_ = a.neo4j.Close(context.Background())
}
