package handling

import (
	"fmt"
	"log"
	"strings"
	"time"
)

const maxUnauthorizedRetries = 10

func Retry(attempts int, delay time.Duration, fn func() error) error {
	unauthCount := 0
	for i := 0; i < attempts; i++ {
		err := fn()
		if err == nil {
			return nil
		}

		if strings.Contains(err.Error(), "User unauthorized") {
			unauthCount++
			log.Printf("[unauthorized retry %d] error: %v", unauthCount, err)

			if unauthCount >= maxUnauthorizedRetries {
				return fmt.Errorf("too many unauthorized retries (%d)", unauthCount)
			}

			time.Sleep(delay)
			i-- // tetap ulangi tanpa habis attempt utama
			continue
		}

		log.Printf("[retry %d/%d] error: %v", i+1, attempts, err)
		time.Sleep(delay)
	}

	return fmt.Errorf("failed after %d retries", attempts)
}
