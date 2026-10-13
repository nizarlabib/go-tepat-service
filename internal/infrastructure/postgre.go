package infrastructure

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

type PostgresConfig struct {
	Connection string
}

type Postgres interface {
	GetClient() *sql.DB
	GetClientGorm() *gorm.DB
	Close()
}

type postgresClient struct {
	client *sql.DB
}

func (p postgresClient) GetClient() *sql.DB {
	return p.client
}

func (p postgresClient) GetClientGorm() *gorm.DB {
	return db
}

func (p postgresClient) Close() {
	slog.Info("Closing PostgreSQL connection...")
	p.client.Close()
}

func NewPostgres(connection string) Postgres {
	slog.Info("Connecting to PostgreSQL...")

	// Initialize sql.DB client
	conn, err := sql.Open("postgres", connection)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	// Test the connection
	if err = conn.Ping(); err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	isLocal := viper.Get("APP_ENV") == "local"
	logLevel := logger.Silent
	if isLocal {
		logLevel = logger.Info
	}
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel:                  logLevel,
			SlowThreshold:             100 * time.Millisecond,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      !isLocal,
			Colorful:                  isLocal,
		},
	)
	gormConfig := &gorm.Config{
		Logger: gormLogger,
	}

	// Initialize GORM client
	gormDB, err := gorm.Open(gormPostgres.Open(connection), gormConfig)
	if err != nil {
		log.Fatal("Failed to initialize GORM for PostgreSQL:", err)
	}

	log.Println("Connection to PostgreSQL established successfully...")

	db = gormDB

	// Return the client
	return postgresClient{client: conn}
}

func NewDBTrx(ctx context.Context) *gorm.DB {
	return db.WithContext(ctx).Begin()
}
