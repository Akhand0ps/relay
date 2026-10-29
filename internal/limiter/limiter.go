package limiter

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
)

// EndpointLimiter maintains a token bucket rate limiter for each destination endpoint.
// It ensures high-volume traffic to one endpoint does not overwhelm the destination server
// or cause thundering herds.
type EndpointLimiter struct {
	mu           sync.RWMutex
	limiters     map[string]*rate.Limiter
	defaultRate  rate.Limit // Refill rate: tokens added per second
	defaultBurst int        // Bucket capacity: max tokens held at once
}

// NewEndpointLimiter initializes the per-destination rate limiter registry.
// rps: tokens refilled per second (e.g. 10.0)
// burst: maximum burst capacity (e.g. 20)
func NewEndpointLimiter(rps float64, burst int) *EndpointLimiter {
	return &EndpointLimiter{
		limiters:     make(map[string]*rate.Limiter),
		defaultRate:  rate.Limit(rps),
		defaultBurst: burst,
	}
}

// GetLimiter safely retrieves or creates a token bucket for a specific endpoint ID.
// Uses double-checked locking to ensure high read concurrency across worker goroutines.
func (el *EndpointLimiter) GetLimiter(endpointID string) *rate.Limiter {
	el.mu.RLock()
	l, exists := el.limiters[endpointID]
	el.mu.RUnlock()
	if exists {
		return l
	}

	el.mu.Lock()
	defer el.mu.Unlock()

	// Double-check in case another goroutine created it while acquiring the write lock
	if l, exists = el.limiters[endpointID]; exists {
		return l
	}

	l = rate.NewLimiter(el.defaultRate, el.defaultBurst)
	el.limiters[endpointID] = l
	return l
}

// Wait blocks until 1 token is available in the endpoint bucket, or until the context is cancelled.
// It uses efficient goroutine sleep without busy-spinning.
func (el *EndpointLimiter) Wait(ctx context.Context, endpointID string) error {
	return el.GetLimiter(endpointID).Wait(ctx)
}

// Allow reports whether a token may be consumed immediately without waiting.
func (el *EndpointLimiter) Allow(endpointID string) bool {
	return el.GetLimiter(endpointID).Allow()
}
