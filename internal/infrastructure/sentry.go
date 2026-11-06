package infrastructure

import (
	"fmt"

	"github.com/getsentry/sentry-go"
)

func NewSentry(options sentry.ClientOptions) error {
	err := sentry.Init(options)
	if err != nil {
		return fmt.Errorf("sentry.Init: %w", err)
	}

	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category: "init",
		Message:  "Server started and Sentry initialized",
		Level:    sentry.LevelInfo,
	})

	return nil
}
