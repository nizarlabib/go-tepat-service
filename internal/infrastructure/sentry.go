package infrastructure

import (
	"log"
	"time"

	"github.com/getsentry/sentry-go"
)

func NewSentry(options sentry.ClientOptions) {
	err := sentry.Init(options)
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	// Tambahkan breadcrumb global yang berlaku sepanjang waktu server berjalan
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category: "init",
		Message:  "Server started and Sentry initialized",
		Level:    sentry.LevelInfo,
	})

	defer sentry.Flush(2 * time.Second)
}
