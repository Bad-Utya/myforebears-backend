package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/Bad-Utya/myforebears-backend/services/auth/internal/app/grpc"
	"github.com/Bad-Utya/myforebears-backend/services/auth/internal/services/auth"
	"github.com/Bad-Utya/myforebears-backend/services/auth/internal/services/rabbitmq"
	"github.com/Bad-Utya/myforebears-backend/services/auth/internal/storage/postgres"
	redisstorage "github.com/Bad-Utya/myforebears-backend/services/auth/internal/storage/redis"
)

type App struct {
	GRPCServer *grpcapp.App
	publisher  *rabbitmq.Publisher
}

func New(
	log *slog.Logger,
	grpcPort int,
	pgHost string,
	pgPort int,
	pgUser string,
	pgPassword string,
	pgDBName string,
	redisAddr string,
	redisPassword string,
	redisDB int,
	jwtSecret string,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
	linkForResetPassword string,
	linkTTL time.Duration,
	rabbitURL string,
	rabbitExchange string,
	rabbitRoutingKey string,
) *App {
	// Fail fast on startup if storage connections cannot be established.
	userStorage, err := postgres.New(pgHost, pgPort, pgUser, pgPassword, pgDBName)
	if err != nil {
		panic(err)
	}

	verificationStorage, err := redisstorage.New(redisAddr, redisPassword, redisDB)
	if err != nil {
		panic(err)
	}

	// Publisher sends email tasks to RabbitMQ.
	pub, err := rabbitmq.New(log, rabbitURL, rabbitExchange, rabbitRoutingKey)
	if err != nil {
		panic(err)
	}

	// Auth service uses Postgres for users, Redis for verification codes,
	// and RabbitMQ to dispatch verification emails.
	authService := auth.New(log, userStorage, verificationStorage, jwtSecret, accessTokenTTL, refreshTokenTTL, linkForResetPassword, linkTTL, verificationStorage, pub)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
		publisher:  pub,
	}
}

// Stop shuts down the gRPC server and releases the RabbitMQ connection.
func (a *App) Stop() {
	a.GRPCServer.Stop()
	a.publisher.Close()
}
