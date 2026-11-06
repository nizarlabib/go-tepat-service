package helper

import (
	"context"
	"log/slog"
	"os"
	"regexp"

	"github.com/spf13/viper"
)

func InitLog(ctx context.Context) {
	var level slog.Level
	switch viper.GetString("APP_ENV") {
	case "production":
		level = slog.LevelInfo
	case "staging":
		level = slog.LevelWarn
	default:
		level = slog.LevelDebug
	}
	loggerOpts := slog.HandlerOptions{
		Level: level,
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &loggerOpts))

	slog.SetDefault(logger)
}

func StrToOnlyNum(str string) string {
	re := regexp.MustCompile(`[^0-9]`)
	return re.ReplaceAllString(str, "")
}
