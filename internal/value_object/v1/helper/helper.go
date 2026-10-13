package helper

import (
	"context"
	"log/slog"
	"os"
	"regexp"
)

func InitLog(ctx context.Context) {
	loggerOpts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &loggerOpts))

	slog.SetDefault(logger)
}

func StrToOnlyNum(str string) string {
	re := regexp.MustCompile(`[^0-9]`)
	return re.ReplaceAllString(str, "")
}
