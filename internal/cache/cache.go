// Package cache provides a concurrency-safe in-memory cache with per-entry
// expiry, which also collapses concurrent lookups of the same key into a single
// fetch. It exists to keep repeated requests from reaching the upstream APIs,
// whose quotas and usage rules the application has to respect.
package cache

import (
	"context"
	"sync"
	"time"
)

// Upper bound on the number of entries kept, so that a long-running process
// cannot grow the cache without limit.
const defaultMaxEntries = 1024

type entry[V any] struct {
	// Closed once value and err have been written, which also publishes them to
	// goroutines waiting on the same key.
	ready chan struct{}

	value   V
	err     error
	expires time.Time
}

type Cache[V any] struct {
	ttl        time.Duration
	maxEntries int
	// Overridable so that expiry can be tested without sleeping.
	now func() time.Time

	mu      sync.Mutex
	entries map[string]*entry[V]
}

func New[V any](ttl time.Duration) *Cache[V] {
	return &Cache[V]{
		ttl:        ttl,
		maxEntries: defaultMaxEntries,
		now:        time.Now,
		entries:    make(map[string]*entry[V]),
	}
}

// Get returns the cached value for key, calling fetch to produce it when the
// key is missing or expired. Concurrent calls for the same key share a single
// fetch: the first caller runs it and the others wait for its result. Failed
// fetches are not cached.
//
// The shared fetch runs with the first caller's context, so cancelling that
// caller also fails the callers waiting on it. They are free to retry.
func (c *Cache[V]) Get(ctx context.Context, key string, fetch func(context.Context) (V, error)) (V, error) {
	c.mu.Lock()

	if existing, found := c.entries[key]; found {
		select {
		case <-existing.ready:
			if existing.err == nil && c.now().Before(existing.expires) {
				value := existing.value
				c.mu.Unlock()
				return value, nil
			}

			// Expired, so drop it and fetch a fresh value below.
			delete(c.entries, key)
		default:
			// Another caller is already fetching this key, so wait for it
			// instead of issuing a second identical request.
			c.mu.Unlock()

			select {
			case <-existing.ready:
				return existing.value, existing.err
			case <-ctx.Done():
				var zero V
				return zero, ctx.Err()
			}
		}
	}

	if len(c.entries) >= c.maxEntries {
		c.evictLocked()
	}

	pending := &entry[V]{ready: make(chan struct{})}
	c.entries[key] = pending
	c.mu.Unlock()

	pending.value, pending.err = fetch(ctx)

	c.mu.Lock()
	pending.expires = c.now().Add(c.ttl)
	if pending.err != nil && c.entries[key] == pending {
		delete(c.entries, key)
	}
	c.mu.Unlock()

	close(pending.ready)

	return pending.value, pending.err
}

// Removes expired entries and, if the cache is still full, arbitrary completed
// ones. Entries that are still being fetched are always kept, because a waiter
// may already be holding on to them. Callers must hold c.mu.
func (c *Cache[V]) evictLocked() {
	now := c.now()

	for key, e := range c.entries {
		select {
		case <-e.ready:
			if !now.Before(e.expires) {
				delete(c.entries, key)
			}
		default:
		}
	}

	for key, e := range c.entries {
		if len(c.entries) < c.maxEntries {
			return
		}

		select {
		case <-e.ready:
			delete(c.entries, key)
		default:
		}
	}
}
