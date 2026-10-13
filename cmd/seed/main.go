package main

import (
	"log/slog"
	"os"

	"go-memolab-service/internal/infrastructure"

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

	db := infrastructure.NewPostgres(viper.GetString("POSTGRES_CONNECTION")).GetClientGorm()

	permissions := seedPermissions(db)

	seedRoles(permissions, db)

	logger.Info("Seeding process completed")
}
