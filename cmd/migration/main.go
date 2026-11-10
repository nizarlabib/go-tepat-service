package main

import (
	"log/slog"
	"os"

	"go-starter-service/internal/entity"
	infrastructure2 "go-starter-service/internal/infrastructure"

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
		// entity.File{},
		// entity.User{},
		// entity.Role{},
		// entity.DeviceLog{},
		// entity.Permission{},
		entity.Category{},
		entity.Product{},
	); err != nil {
		slog.Error("migration failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Partial unique indexes for slug so the same slug can be reused after
	// a record is soft-deleted (deleted_at IS NOT NULL).
	// partialIndexes := []string{
	// 	`CREATE UNIQUE INDEX IF NOT EXISTS idx_products_slug_not_deleted ON products(slug) WHERE deleted_at IS NULL`,
	// 	`CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_slug_not_deleted ON categories(slug) WHERE deleted_at IS NULL`,
	// }
	// for _, idxSQL := range partialIndexes {
	// 	if err := db.GetClientGorm().Exec(idxSQL).Error; err != nil {
	// 		slog.Error("failed to create partial unique index", slog.String("error", err.Error()))
	// 		os.Exit(1)
	// 	}
	// }

	slog.Info("migration success")
}
