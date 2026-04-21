package app

import (
	"context"
	"log/slog"
	"time"

	grpcapp "github.com/Bad-Utya/myforebears-backend/services/familytree/internal/app/grpc"
	eventsclient "github.com/Bad-Utya/myforebears-backend/services/familytree/internal/clients/events"
	familytreesvc "github.com/Bad-Utya/myforebears-backend/services/familytree/internal/services/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/storage/neo4j"
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/storage/postgres"
)

type App struct {
	GRPCServer *grpcapp.App
	postgres   *postgres.Storage
	neo4j      *neo4j.Storage
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
	neo4jURI string,
	neo4jUser string,
	neo4jPassword string,
	eventsAddr string,
	eventsTimeout time.Duration,
	eventsRetries int,
) *App {
	personStorage, err := postgres.New(pgHost, pgPort, pgUser, pgPassword, pgDBName)
	if err != nil {
		panic(err)
	}

	relStorage, err := neo4j.New(neo4jURI, neo4jUser, neo4jPassword)
	if err != nil {
		panic(err)
	}

	events, err := eventsclient.New(log, eventsAddr, eventsTimeout, eventsRetries)
	if err != nil {
		panic(err)
	}

	personService := familytreesvc.New(log, personStorage, relStorage, events)
	grpcServer := grpcapp.New(log, personService, grpcPort)

	return &App{
		GRPCServer: grpcServer,
		postgres:   personStorage,
		neo4j:      relStorage,
		events:     events,
	}
}

func (a *App) Stop() {
	a.GRPCServer.Stop()
	a.postgres.Close()
	_ = a.neo4j.Close(context.Background())
	_ = a.events.Close()
}
