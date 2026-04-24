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
	photosclient "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/clients/photos"
	redisclient "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/clients/redis"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/config"
	authhandler "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/handlers/auth"
	eventshandler "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/handlers/events"
	familytreehandler "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/handlers/familytree"
	photoshandler "github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/handlers/photos"
	"github.com/Bad-Utya/myforebears-backend/services/apigateway/internal/middleware"
)

type App struct {
	log              *slog.Logger
	httpServer       *http.Server
	authClient       *authclient.Client
	familyTreeClient *familytreeclient.Client
	eventsClient     *eventsclient.Client
	photosClient     *photosclient.Client
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

	photosGRPC, err := photosclient.New(
		ctx,
		log,
		cfg.Clients.Photos.Address,
		cfg.Clients.Photos.Timeout,
		cfg.Clients.Photos.RetriesCount,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to photos service: %v", err))
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
	photosHandler := photoshandler.New(log, photosGRPC)

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

			r.Post("/", familyTreeHandler.CreateTree)
			r.Get("/", familyTreeHandler.ListTrees)
			r.Get("/{tree_id}", familyTreeHandler.GetTree)
			r.Patch("/{tree_id}", familyTreeHandler.UpdateTreeSettings)
			r.Get("/{tree_id}/content", familyTreeHandler.GetTreeContent)
			r.Post("/{tree_id}/access-emails", familyTreeHandler.AddTreeAccessEmail)
			r.Get("/{tree_id}/access-emails", familyTreeHandler.ListTreeAccessEmails)
			r.Delete("/{tree_id}/access-emails", familyTreeHandler.DeleteTreeAccessEmail)

			r.Post("/{tree_id}/parents", familyTreeHandler.AddParent)
			r.Post("/{tree_id}/children", familyTreeHandler.AddChild)
			r.Post("/{tree_id}/partners", familyTreeHandler.AddPartner)
			r.Get("/{tree_id}/persons", familyTreeHandler.ListPersons)
			r.Get("/{tree_id}/persons/{person_id}", familyTreeHandler.GetPerson)

			r.Patch("/{tree_id}/persons/{person_id}", familyTreeHandler.UpdatePersonName)
			r.Delete("/{tree_id}/persons/{person_id}", familyTreeHandler.DeletePerson)
		})
	})

	router.Route("/api/event-types", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(tokenChecker.Middleware)
			r.Get("/", eventsHandler.ListEventTypes)
			r.Get("/{event_type_id}", eventsHandler.GetEventType)
			r.Post("/", eventsHandler.CreateEventType)
			r.Delete("/{event_type_id}", eventsHandler.DeleteEventType)
		})
	})

	router.Route("/api/events", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(tokenChecker.Middleware)
			r.Get("/", eventsHandler.ListEventsByTree)
			r.Get("/{event_id}", eventsHandler.GetEvent)
			r.Post("/", eventsHandler.CreateEvent)
			r.Put("/{event_id}", eventsHandler.UpdateEvent)
			r.Delete("/{event_id}", eventsHandler.DeleteEvent)
		})
	})

	router.Route("/api/photos", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(tokenChecker.Middleware)

			r.Post("/user/avatar", photosHandler.UploadUserAvatar)
			r.Get("/user/avatar", photosHandler.GetUserAvatar)

			r.Post("/persons/{person_id}/avatar", photosHandler.UploadPersonAvatar)
			r.Get("/persons/{person_id}/avatar", photosHandler.GetPersonAvatar)
			r.Post("/persons/{person_id}", photosHandler.UploadPersonPhoto)
			r.Get("/persons/{person_id}", photosHandler.ListPersonPhotos)

			r.Post("/events/{event_id}", photosHandler.UploadEventPhoto)
			r.Get("/events/{event_id}", photosHandler.ListEventPhotos)

			r.Get("/{photo_id}", photosHandler.GetPhotoByID)
			r.Delete("/{photo_id}", photosHandler.DeletePhotoByID)
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
		photosClient:     photosGRPC,
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

	if err := a.photosClient.Close(); err != nil {
		a.log.Error("failed to close photos grpc connection", slog.String("op", op), slog.String("error", err.Error()))
	}

	if err := a.redisClient.Close(); err != nil {
		a.log.Error("failed to close redis connection", slog.String("op", op), slog.String("error", err.Error()))
	}
}
