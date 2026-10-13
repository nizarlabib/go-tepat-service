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
	"go-memolab-service/internal/pkg/ratelimiter"
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

	helper.InitLog(ctx)

	infrastructure.NewSentry(sentry.ClientOptions{
		Dsn:              viper.GetString("SENTRY_DSN"),
		EnableTracing:    viper.GetBool("SENTRY_ENABLE_TRACKING"),
		TracesSampleRate: viper.GetFloat64("SENTRY_TRACES_SAMPLE_RATE"),
	})
	defer sentry.Flush(2 * time.Second)

	db := infrastructure.NewPostgres(viper.GetString("POSTGRES_CONNECTION"))
	defer db.Close()

	redis := infrastructure.NewRedis(ctx, viper.GetString("REDIS_CONNECTION"))
	defer redis.Close()

	minioClient := infrastructure.NewMinio()
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
		// redis,
	); err != nil {
		slog.Error("failed to register health check route", slog.Any("error", err))
		os.Exit(1)
	}

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	// Important: Chi has a middleware stack and thus it is important to put the
	// Sentry handler on the appropriate place. If using middleware.Recoverer,
	// the Sentry middleware must come afterwards (and configure it with
	// Repanic: true)
	r.Use(sentryHandler.Handle)
	r.Use(chimiddleware.RealIP)
	r.Use(ratelimiter.Middleware(10000, 15*time.Minute))
	r.Use(spesial.WithRequestInfo)
	r.Use(chimiddleware.Heartbeat("/ping"))
	r.Use(healthCheckRoute.Middleware("/health"))

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
	httpTimeout := time.Duration(150) * time.Second

	httpServer := &http.Server{
		Addr:         httpPort,
		Handler:      http.TimeoutHandler(r, httpTimeout, ""),
		ReadTimeout:  httpTimeout,
		WriteTimeout: httpTimeout + 2,
	}

	go func() {
		slog.Info("Server listening on", slog.Any("port", httpPort))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Error starting server", slog.Any("error", err))
		}
	}()

	<-ctx.Done()
	slog.Info("Received shutdown signal, shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("unexpected error during server shutdown", slog.Any("bind_addr", err))
	}

	select {
	case <-shutdownCtx.Done():
		slog.Error("Timeout exceeded: forced exit while waiting for workers.")
	default:
		slog.Info("Shutdown complete. Exiting.")
	}
}
