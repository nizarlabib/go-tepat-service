package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
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
	if p.client != nil {
		if err := p.client.Close(); err != nil {
			slog.Error("error closing PostgreSQL connection", slog.String("error", err.Error()))
		}
	}
}

func NewPostgres(connection string) (*postgresClient, error) {
	slog.Info("Connecting to PostgreSQL...")

	conn, err := sql.Open("postgres", connection)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	if err = conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
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

	gormDB, err := gorm.Open(gormPostgres.Open(connection), gormConfig)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize GORM for PostgreSQL: %w", err)
	}

	slog.Info("Connection to PostgreSQL established successfully...")

	db = gormDB

	return &postgresClient{client: conn}, nil
}

func NewDBTrx(ctx context.Context) *gorm.DB {
	if db == nil {
		return nil
	}
	return db.WithContext(ctx).Begin()
}
