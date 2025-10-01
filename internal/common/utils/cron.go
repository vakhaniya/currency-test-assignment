package utils

import (
	"context"
	"time"

	"currency-rate-app/internal/common/tracing"
)

func CreateCronJob(ctx context.Context, interval time.Duration, job func(ctx context.Context)) context.CancelFunc {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				func() {
					defer HandleRecover()
					job(tracing.WithTraceID(ctx, tracing.NewTraceID()))
				}()
			}
		}
	}()

	return cancel
}
