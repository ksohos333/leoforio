// Package cache provides a TTL cache that coalesces concurrent loads of the
// same key into a single upstream call.
//
// Coalescing is what keeps upstream traffic proportional to distinct keys
// rather than to the number of clients (see ADR 0004).
package cache

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// Getter loads a value for a key.
type Getter[T any] func(ctx context.Context) (T, error)

type entry[T any] struct {
	value     T
	fetchedAt time.Time
}

// Cache is a concurrency-safe TTL cache for a single value type.
type Cache[T any] struct {
	ttl     time.Duration
	group   singleflight.Group
	mu      sync.Mutex
	entries map[string]entry[T]
}

// New returns a cache whose entries expire after ttl.
func New[T any](ttl time.Duration) *Cache[T] {
	return &Cache[T]{
		ttl:     ttl,
		entries: make(map[string]entry[T]),
	}
}

// Get returns the cached value for key, loading it with get when needed.
//
// The stale flag is true when the upstream load failed but a previously
// fetched value was still available, so callers can serve last-known data
// and mark it as such instead of reporting an error.
func (c *Cache[T]) Get(ctx context.Context, key string, get Getter[T]) (value T, stale bool, err error) {
	if v, ok := c.fresh(key); ok {
		return v, false, nil
	}

	// The load is detached from the caller's cancellation: a client that
	// disconnects mid-flight must not cancel a load that other clients are
	// waiting on.
	loadCtx := context.WithoutCancel(ctx)

	result, err, _ := c.group.Do(key, func() (any, error) {
		if v, ok := c.fresh(key); ok {
			return v, nil
		}
		loaded, loadErr := get(loadCtx)
		if loadErr != nil {
			return loaded, loadErr
		}
		c.store(key, loaded)
		return loaded, nil
	})
	if err != nil {
		if v, ok := c.any(key); ok {
			return v, true, err
		}
		var zero T
		return zero, false, err
	}

	return result.(T), false, nil
}

func (c *Cache[T]) fresh(key string) (T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok {
		var zero T
		return zero, false
	}
	if time.Since(e.fetchedAt) > c.ttl {
		var zero T
		return zero, false
	}
	return e.value, true
}

func (c *Cache[T]) any(key string) (T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok {
		var zero T
		return zero, false
	}
	return e.value, true
}

func (c *Cache[T]) store(key string, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = entry[T]{value: value, fetchedAt: time.Now()}
}
