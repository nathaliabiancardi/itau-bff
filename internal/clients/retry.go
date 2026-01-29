package clients

import (
	"context"
	"time"
)

func retry(ctx context.Context, attempts int, fn func() error) error {
	backoff := 100 * time.Millisecond

	for i := 0; i < attempts; i++ {
		err := fn()
		if err == nil {
			return nil
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		time.Sleep(backoff)
		backoff *= 2
	}

	return ctx.Err()
}
