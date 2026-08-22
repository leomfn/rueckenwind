package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetCachesUntilExpiry(t *testing.T) {
	c := New[int](time.Minute)

	current := time.Unix(0, 0)
	c.now = func() time.Time { return current }

	var fetches atomic.Int64
	fetch := func(context.Context) (int, error) {
		fetches.Add(1)
		return 42, nil
	}

	for range 5 {
		value, err := c.Get(context.Background(), "key", fetch)
		if err != nil || value != 42 {
			t.Fatalf("expected 42, but got %v (%v)", value, err)
		}
	}

	if fetches.Load() != 1 {
		t.Errorf("expected 1 fetch while the entry is fresh, but got %d", fetches.Load())
	}

	// Move past the ttl.
	current = current.Add(2 * time.Minute)

	if _, err := c.Get(context.Background(), "key", fetch); err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	if fetches.Load() != 2 {
		t.Errorf("expected a refetch after expiry, but got %d fetches", fetches.Load())
	}
}

func TestGetSeparatesKeys(t *testing.T) {
	c := New[string](time.Minute)

	for _, key := range []string{"a", "b", "a"} {
		value, err := c.Get(context.Background(), key, func(context.Context) (string, error) {
			return "value-" + key, nil
		})

		if err != nil || value != "value-"+key {
			t.Fatalf("expected value-%s, but got %v (%v)", key, value, err)
		}
	}
}

// Concurrent lookups of the same key must result in a single upstream call.
func TestGetCollapsesConcurrentFetches(t *testing.T) {
	c := New[int](time.Minute)

	var fetches atomic.Int64
	release := make(chan struct{})

	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			value, err := c.Get(context.Background(), "key", func(context.Context) (int, error) {
				fetches.Add(1)
				<-release
				return 7, nil
			})

			if err != nil || value != 7 {
				t.Errorf("expected 7, but got %v (%v)", value, err)
			}
		})
	}

	// Give the goroutines a chance to queue up behind the first fetch.
	time.Sleep(50 * time.Millisecond)
	close(release)
	wg.Wait()

	if fetches.Load() != 1 {
		t.Errorf("expected 1 shared fetch, but got %d", fetches.Load())
	}
}

func TestGetDoesNotCacheFailures(t *testing.T) {
	c := New[int](time.Minute)

	var fetches atomic.Int64
	fetch := func(context.Context) (int, error) {
		fetches.Add(1)
		return 0, errors.New("upstream is down")
	}

	for range 3 {
		if _, err := c.Get(context.Background(), "key", fetch); err == nil {
			t.Fatal("expected an error, but got none")
		}
	}

	if fetches.Load() != 3 {
		t.Errorf("expected every failed lookup to retry, but got %d fetches", fetches.Load())
	}
}

func TestGetHonoursWaiterCancellation(t *testing.T) {
	c := New[int](time.Minute)

	started := make(chan struct{})
	release := make(chan struct{})
	defer close(release)

	go func() {
		c.Get(context.Background(), "key", func(context.Context) (int, error) {
			close(started)
			<-release
			return 1, nil
		})
	}()

	<-started

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := c.Get(ctx, "key", func(context.Context) (int, error) {
		t.Error("waiting caller must not start its own fetch")
		return 0, nil
	}); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, but got %v", err)
	}
}

// The cache must stay bounded even when every key is distinct.
func TestGetEvictsWhenFull(t *testing.T) {
	c := New[int](time.Minute)
	c.maxEntries = 16

	current := time.Unix(0, 0)
	c.now = func() time.Time { return current }

	for i := range 200 {
		if _, err := c.Get(context.Background(), fmt.Sprintf("key-%d", i),
			func(context.Context) (int, error) { return i, nil }); err != nil {
			t.Fatalf("expected no error, but got %v", err)
		}

		// Age entries so that the expiry sweep has something to collect.
		current = current.Add(10 * time.Second)
	}

	c.mu.Lock()
	size := len(c.entries)
	c.mu.Unlock()

	if size > c.maxEntries {
		t.Errorf("expected at most %d entries, but got %d", c.maxEntries, size)
	}
}
