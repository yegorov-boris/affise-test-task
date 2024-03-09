package ratelimiter

import (
	"context"
	"sync/atomic"
)

type RateLimiter struct {
	serverStopped atomic.Bool
	bucket        chan struct{}
}

func New(ctx context.Context, bucketSize uint32) *RateLimiter {
	rl := new(RateLimiter)
	if bucketSize != 0 {
		rl.bucket = make(chan struct{}, bucketSize)
	}

	go func() {
		<-ctx.Done()
		rl.serverStopped.Store(true)
	}()

	return rl
}

func (rl *RateLimiter) Try() bool {
	if rl.serverStopped.Load() {
		return false
	}

	select {
	case rl.bucket <- struct{}{}:
		return true
	default:
		return false
	}
}

func (rl *RateLimiter) IsEmpty() bool {
	return rl.bucket == nil || len(rl.bucket) == 0
}
