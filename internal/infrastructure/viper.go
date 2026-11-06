package infrastructure

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func NewViper() {
	// Load the .env file using godotenv
	if err := godotenv.Load(); err != nil {
		slog.WarnContext(context.Background(), "No .env file found")
	}

	viper.SetConfigName(".env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../../")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		slog.WarnContext(context.Background(), "error while reading .env file", slog.String("error", err.Error()))
	}

	viper.SetDefault("APP_ENV", "local")
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "local"
	}
	viper.SetDefault("APP_VERSION", version)
	viper.SetDefault("APP_SERVICE_NAME", "starter-service")

	viper.SetDefault("AUTH_SERVICE_URL", "http://127.0.0.1:9000/api/v1")
}

var requiredConfigKeys = []string{
	"POSTGRES_CONNECTION",
	"REDIS_CONNECTION",
	"MINIO_ENDPOINT",
	"MINIO_ACCESS_KEY",
	"MINIO_SECRET_KEY",
	"MINIO_BUCKET",
	"JWT_KEY",
	"AES_KEY",
	"HTTP_PORT",
}

func ValidateConfig() error {
	var missing []string
	for _, key := range requiredConfigKeys {
		value := strings.TrimSpace(viper.GetString(key))
		if value == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return errors.New("missing required configuration: " + strings.Join(missing, ", "))
	}
	return nil
}
