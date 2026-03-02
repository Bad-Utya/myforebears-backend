package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/Bad-Utya/myforebears-backend/services/auth/internal/app/grpc"
	"github.com/Bad-Utya/myforebears-backend/services/auth/internal/services/auth"
	"github.com/Bad-Utya/myforebears-backend/services/auth/internal/storage/postgres"
	redisstorage "github.com/Bad-Utya/myforebears-backend/services/auth/internal/storage/redis"
)

type App struct {
	GRPCServer *grpcapp.App
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

	// Auth service uses Postgres for users and Redis for verification codes.
	authService := auth.New(log, userStorage, verificationStorage, jwtSecret, accessTokenTTL, refreshTokenTTL, linkForResetPassword, linkTTL, verificationStorage)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
