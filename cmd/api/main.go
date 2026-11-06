package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-memolab-service/internal/api"
	"go-memolab-service/internal/core"
	"go-memolab-service/internal/framework/primary/rest"
	"go-memolab-service/internal/framework/secondary/repository"
	"go-memolab-service/internal/framework/secondary/storage"
	"go-memolab-service/internal/infrastructure"
	appmiddleware "go-memolab-service/internal/middleware"
	"go-memolab-service/internal/pkg/health"
	"go-memolab-service/internal/pkg/middleware"
	"go-memolab-service/internal/pkg/ratelimiter"
	"go-memolab-service/internal/pkg/security"
	"go-memolab-service/internal/spesial"
	"go-memolab-service/internal/value_object/v1/helper"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/spf13/viper"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	infrastructure.NewViper()

	if err := infrastructure.ValidateConfig(); err != nil {
		slog.Error("configuration validation failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	helper.InitLog(ctx)

	var sentryInitialized bool
	if err := infrastructure.NewSentry(sentry.ClientOptions{
		Dsn:              viper.GetString("SENTRY_DSN"),
		EnableTracing:    viper.GetBool("SENTRY_ENABLE_TRACKING"),
		TracesSampleRate: viper.GetFloat64("SENTRY_TRACES_SAMPLE_RATE"),
	}); err != nil {
		slog.Warn("failed to initialize Sentry", slog.String("error", err.Error()))
	} else {
		sentryInitialized = true
	}

	db, err := infrastructure.NewPostgres(viper.GetString("POSTGRES_CONNECTION"))
	if err != nil {
		slog.Error("failed to connect to PostgreSQL", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	redis, err := infrastructure.NewRedis(ctx, viper.GetString("REDIS_CONNECTION"))
	if err != nil {
		slog.Error("failed to connect to Redis", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer redis.Close()

	minioClient, err := infrastructure.NewMinio()
	if err != nil {
		slog.Error("failed to connect to MinIO", slog.String("error", err.Error()))
		os.Exit(1)
	}
	minioService := storage.NewMinio(minioClient)

	storageService := core.NewStorage(
		core.StorageWithStorage(minioService),
	)

	fileRepo := repository.NewFile(db.GetClientGorm())
	fileProject := core.NewFile(
		core.FileWithFileRepo(fileRepo),
		core.FileWithStorage(storageService),
	)
	fileApi := api.NewFile(fileProject)

	userRepo := repository.NewUser(db, redis.GetClient(), storageService)
	userService := core.NewUser(userRepo)
	userApi := api.NewUser(userService)

	deviceLogRepo := repository.NewdeviceLog(db)
	deviceLogService := core.NewDeviceLog(deviceLogRepo)

	authService := core.NewAuth(
		core.AuthWithRepo(userRepo),
		core.AuthWithRedis(redis.GetClient()),
		core.AuthWithDeviceLog(deviceLogService),
	)
	authApi := api.NewAuth(authService)

	sentryHandler := sentryhttp.New(sentryhttp.Options{
		Repanic:         true,
		Timeout:         2 * 1000,
		WaitForDelivery: false,
	})

	healthCheckRoute := health.NewHealthRoute()
	if err := healthCheckRoute.Register(
		db,
		redis,
	); err != nil {
		slog.Error("failed to register health check route", slog.Any("error", err))
		os.Exit(1)
	}

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(security.CORS)
	r.Use(security.SecurityHeaders)
	r.Use(middleware.RequestID)
	r.Use(middleware.BodyLimit(10 << 20))
	r.Use(middleware.ContentType("application/json"))
	r.Use(middleware.Compression)
	r.Use(sentryHandler.Handle)
	r.Use(chimiddleware.RealIP)
	r.Use(ratelimiter.Middleware(10000, 15*time.Minute))
	r.Use(spesial.WithRequestInfo)
	r.Use(chimiddleware.Heartbeat("/ping"))
	r.Use(healthCheckRoute.Middleware("/health"))
	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	})

	// Swagger UI (only in non-production environments)
	if viper.GetString("APP_ENV") != "production" {
		r.Get("/swagger/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "docs/openapi.yaml")
		})
		r.Get("/swagger/index.html", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			html := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Go Memolab Service - API Documentation</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
  <script>
    const ui = SwaggerUIBundle({
      url: "/swagger/openapi.yaml",
      dom_id: "#swagger-ui",
      deepLinking: true,
      presets: [
        SwaggerUIBundle.presets.apis,
        SwaggerUIStandalonePreset
      ],
      layout: "StandaloneLayout"
    })
    window.ui = ui
  </script>
</body>
</html>`
			_, _ = w.Write([]byte(html))
		})
	}

	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Mount("/auth", rest.NewAuth(authApi, userRepo, storageService, redis.GetClient()).Start(ctx))

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(appmiddleware.AuthMiddleware(userRepo))

			r.Mount("/users", rest.NewUser(userApi).Start(ctx))
			r.Mount("/files", rest.NewFile(fileApi).Start(ctx))
		})
	})

	httpPort := fmt.Sprintf(":%v", viper.GetString("HTTP_PORT"))
	httpTimeout := time.Duration(30) * time.Second

	httpServer := &http.Server{
		Addr:         httpPort,
		Handler:      http.TimeoutHandler(r, httpTimeout, ""),
		ReadTimeout:  httpTimeout,
		WriteTimeout: httpTimeout + 2,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("Server listening on", slog.Any("port", httpPort))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Error starting server", slog.Any("error", err))
		}
	}()

	<-ctx.Done()
	slog.Info("Received shutdown signal, shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("unexpected error during server shutdown", slog.Any("bind_addr", err))
	}

	if sentryInitialized {
		sentry.Flush(2 * time.Second)
	}

	select {
	case <-shutdownCtx.Done():
		slog.Error("Timeout exceeded: forced exit while waiting for workers.")
	default:
		slog.Info("Shutdown complete. Exiting.")
	}
}
