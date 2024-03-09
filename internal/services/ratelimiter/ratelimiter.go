package ratelimiter

import "sync/atomic"

type RateLimiter struct {
	serverStopped atomic.Bool
	bucket        chan struct{}
}

func New(bucketSize uint32) *RateLimiter {
	rl := new(RateLimiter)
	if bucketSize != 0 {
		rl.bucket = make(chan struct{}, bucketSize)
	}

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

func (rl *RateLimiter) Stop() {
	rl.serverStopped.Store(true)
}

func (rl *RateLimiter) IsEmpty() bool {
	return rl.bucket == nil || len(rl.bucket) == 0
}
