package infrastructure

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func NewViper() {
	// Load the .env file using godotenv
	fmt.Println("Loading .env file", os.Getenv("ENV"))
	if err := godotenv.Load(); err != nil {
		slog.WarnContext(context.Background(), "No .env file found")
	}

	viper.SetConfigName(".env")   // Nama file tanpa ekstensi
	viper.AddConfigPath(".")      // Cari di direktori kerja saat ini
	viper.AddConfigPath("../../") // Tambahkan parent directory jika diperlukan
	viper.SetConfigType("env")    // Format file .env
	// Enable automatic environment variable lookup
	viper.AutomaticEnv()

	// Membaca file konfigurasi
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("error while reading .env file: %v\n", err)
	}

	viper.SetDefault("APP_ENV", "local")
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "local"
	}
	viper.SetDefault("APP_VERSION", version)
	viper.SetDefault("APP_SERVICE_NAME", "memolab-service")

	viper.SetDefault("AES_KEY", "super-long-secret-key-12345678901")
	viper.SetDefault("JWT_KEY", "super-long-secret-key")

	viper.SetDefault("AUTH_SERVICE_URL", "http://127.0.0.1:9000/api/v1")
}
