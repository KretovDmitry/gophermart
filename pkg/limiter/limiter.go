package limiter

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

type DynamicRateLimiter struct {
	limiter  *rate.Limiter
	updates  chan rateParams
	interval time.Duration
	burst    int
}

type rateParams struct {
	interval time.Duration
	burst    int
}

func NewDynamicRateLimiter(interval time.Duration, burst int) *DynamicRateLimiter {
	limiter := rate.NewLimiter(rate.Every(interval), burst)
	updates := make(chan rateParams)
	go func() {
		for params := range updates {
			limiter.SetLimit(rate.Every(params.interval))
			limiter.SetBurst(params.burst)
		}
	}()
	return &DynamicRateLimiter{
		limiter:  limiter,
		interval: interval,
		burst:    burst,
		updates:  updates,
	}
}

func (drl *DynamicRateLimiter) Wait(ctx context.Context) error {
	return drl.limiter.Wait(ctx)
}

func (drl *DynamicRateLimiter) Allow() bool {
	return drl.limiter.Allow()
}

func (drl *DynamicRateLimiter) Update(interval time.Duration, burst int) {
	drl.updates <- rateParams{interval: interval, burst: burst}
}
