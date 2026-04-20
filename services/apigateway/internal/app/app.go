package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	authclient "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/clients/auth"
	eventsclient "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/clients/events"
	familytreeclient "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/clients/familytree"
	redisclient "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/clients/redis"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/config"
	authhandler "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/handlers/auth"
	eventshandler "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/handlers/events"
	familytreehandler "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/handlers/familytree"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/middleware"
)

type App struct {
	log              *slog.Logger
	httpServer       *http.Server
	authClient       *authclient.Client
	familyTreeClient *familytreeclient.Client
	eventsClient     *eventsclient.Client
	redisClient      *redisclient.Client
}

func New(log *slog.Logger, cfg *config.Config) *App {
	ctx := context.Background()

	// Connect to auth gRPC service.
	authGRPC, err := authclient.New(
		ctx,
		log,
		cfg.Clients.Auth.Address,
		cfg.Clients.Auth.Timeout,
		cfg.Clients.Auth.RetriesCount,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to auth service: %v", err))
	}

	familyTreeGRPC, err := familytreeclient.New(
		ctx,
		log,
		cfg.Clients.FamilyTree.Address,
		cfg.Clients.FamilyTree.Timeout,
		cfg.Clients.FamilyTree.RetriesCount,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to familytree service: %v", err))
	}

	eventsGRPC, err := eventsclient.New(
		ctx,
		log,
		cfg.Clients.Events.Address,
		cfg.Clients.Events.Timeout,
		cfg.Clients.Events.RetriesCount,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to events service: %v", err))
	}

	// Build HTTP router.
	router := chi.NewRouter()

	// Global middleware.
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-Id"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	router.Use(chimw.RequestID)
	router.Use(chimw.RealIP)
	router.Use(middleware.Logging(log))
	router.Use(chimw.Recoverer)

	// Connect to Redis for token blacklist checks.
	redisClient, err := redisclient.New(
		cfg.Clients.TokenStorage.Address,
		cfg.Clients.TokenStorage.Password,
		cfg.Clients.TokenStorage.Database,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to Redis: %v", err))
	}

	tokenChecker := middleware.NewTokenChecker(redisClient, cfg.JWTSecret, log)

	// Auth routes.
	authHandler := authhandler.New(log, authGRPC)
	familyTreeHandler := familytreehandler.New(log, familyTreeGRPC)
	eventsHandler := eventshandler.New(log, eventsGRPC)

	router.Route("/api/auth", func(r chi.Router) {
		// Public routes.
		r.Post("/send-code", authHandler.SendCode)
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/send-link-for-reset-password", authHandler.SendLinkForResetPassword)
		r.Post("/reset-password-with-link", authHandler.ResetPasswordWithLink)
		r.Post("/refresh", authHandler.RefreshTokens)

		// Protected routes.
		r.Group(func(r chi.Router) {
			r.Use(tokenChecker.Middleware)
			r.Post("/reset-password-with-token", authHandler.ResetPasswordWithToken)
			r.Post("/logout", authHandler.Logout)
			r.Post("/logout-all", authHandler.LogoutFromAllDevices)
		})
	})

	router.Route("/api/familytree", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(tokenChecker.Middleware)

			r.Post("/trees", familyTreeHandler.CreateTree)
			r.Get("/trees", familyTreeHandler.ListTrees)
			r.Get("/trees/{tree_id}", familyTreeHandler.GetTree)

			r.Post("/trees/{tree_id}/parents", familyTreeHandler.AddParent)
			r.Post("/trees/{tree_id}/children", familyTreeHandler.AddChild)
			r.Post("/trees/{tree_id}/partners", familyTreeHandler.AddPartner)

			r.Patch("/trees/{tree_id}/persons/{person_id}", familyTreeHandler.UpdatePersonName)
			r.Delete("/trees/{tree_id}/persons/{person_id}", familyTreeHandler.DeletePerson)
		})
	})

	router.Route("/api/event-types", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(tokenChecker.Middleware)
			r.Post("/", eventsHandler.CreateEventType)
			r.Delete("/{event_type_id}", eventsHandler.DeleteEventType)
		})
	})

	router.Route("/api/events", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(tokenChecker.Middleware)
			r.Post("/", eventsHandler.CreateEvent)
			r.Put("/{event_id}", eventsHandler.UpdateEvent)
			r.Delete("/{event_id}", eventsHandler.DeleteEvent)
		})
	})

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:      router,
		ReadTimeout:  cfg.HTTP.Timeout,
		WriteTimeout: cfg.HTTP.Timeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	return &App{
		log:              log,
		httpServer:       httpServer,
		authClient:       authGRPC,
		familyTreeClient: familyTreeGRPC,
		eventsClient:     eventsGRPC,
		redisClient:      redisClient,
	}
}

func (a *App) Run() error {
	const op = "app.Run"

	a.log.Info("http server is running", slog.String("addr", a.httpServer.Addr))

	if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Stop(ctx context.Context) {
	const op = "app.Stop"

	a.log.Info("stopping http server")

	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.log.Error("failed to shutdown http server", slog.String("op", op), slog.String("error", err.Error()))
	}

	if err := a.authClient.Close(); err != nil {
		a.log.Error("failed to close auth grpc connection", slog.String("op", op), slog.String("error", err.Error()))
	}

	if err := a.familyTreeClient.Close(); err != nil {
		a.log.Error("failed to close familytree grpc connection", slog.String("op", op), slog.String("error", err.Error()))
	}

	if err := a.eventsClient.Close(); err != nil {
		a.log.Error("failed to close events grpc connection", slog.String("op", op), slog.String("error", err.Error()))
	}

	if err := a.redisClient.Close(); err != nil {
		a.log.Error("failed to close redis connection", slog.String("op", op), slog.String("error", err.Error()))
	}
}
