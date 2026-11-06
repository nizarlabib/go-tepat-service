package main

import (
	"log/slog"
	"os"

	"go-memolab-service/internal/entity"
	infrastructure2 "go-memolab-service/internal/infrastructure"

	"github.com/spf13/viper"
)

func main() {
	infrastructure2.NewViper()

	if err := infrastructure2.ValidateConfig(); err != nil {
		slog.Error("configuration validation failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	db, err := infrastructure2.NewPostgres(viper.GetString("POSTGRES_CONNECTION"))
	if err != nil {
		slog.Error("failed to connect to PostgreSQL", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	enumTypes := []string{}

	for _, enumSQL := range enumTypes {
		if err := db.GetClientGorm().Exec(enumSQL).Error; err != nil {
			slog.Error("failed to create enum type", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}

	if err := db.GetClientGorm().AutoMigrate(
		entity.File{},
		entity.User{},
		entity.Role{},
		entity.DeviceLog{},
		entity.Permission{},
	); err != nil {
		slog.Error("migration failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("migration success")
}
