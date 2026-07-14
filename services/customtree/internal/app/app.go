package app

import (
	"context"
	"fmt"
	"github.com/Bad-Utya/myforebears-backend/services/customtree/internal/config"
	"github.com/Bad-Utya/myforebears-backend/services/customtree/internal/handler"
	"github.com/Bad-Utya/myforebears-backend/services/customtree/internal/service"
	"github.com/Bad-Utya/myforebears-backend/services/customtree/internal/storage/postgres"
	objects "github.com/Bad-Utya/myforebears-backend/services/customtree/internal/storage/s3"
	"google.golang.org/grpc"
	"net"
)

type App struct {
	server *grpc.Server
	port   int
	svc    *service.Service
}

func New(c *config.Config) *App {
	db, err := postgres.New(c.Postgres.Host, c.Postgres.Port, c.Postgres.Username, c.Postgres.Password, c.Postgres.Database)
	if err != nil {
		panic(err)
	}
	s3, err := objects.New(context.Background(), c.S3)
	if err != nil {
		panic(err)
	}
	svc := service.New(db, s3)
	g := grpc.NewServer()
	handler.Register(g, svc)
	return &App{g, c.GRPC.Port, svc}
}
func (a *App) MustRun() {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		panic(err)
	}
	if err = a.server.Serve(l); err != nil {
		panic(err)
	}
}
func (a *App) Stop() { a.server.GracefulStop(); a.svc.Close() }
