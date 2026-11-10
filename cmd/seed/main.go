package main

import (
	"log/slog"
	"os"

	"go-starter-service/internal/infrastructure"

	"github.com/spf13/viper"
)

var logger = slog.
	New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	)).
	With("module", "cmd/seed")

func main() {
	logger.Info("Starting seeding process")

	infrastructure.NewViper()

	if err := infrastructure.ValidateConfig(); err != nil {
		logger.Error("configuration validation failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	db, err := infrastructure.NewPostgres(viper.GetString("POSTGRES_CONNECTION"))
	if err != nil {
		logger.Error("failed to connect to PostgreSQL", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	permissions := seedPermissions(db.GetClientGorm())

	seedRoles(permissions, db.GetClientGorm())

	seedCategories(db.GetClientGorm())

	logger.Info("Seeding process completed")
}
