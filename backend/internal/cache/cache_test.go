package cache

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCoalescesConcurrentLoads(t *testing.T) {
	c := New[string](time.Minute)

	var calls atomic.Int64
	start := make(chan struct{})
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			value, stale, err := c.Get(context.Background(), "k", func(context.Context) (string, error) {
				calls.Add(1)
				time.Sleep(20 * time.Millisecond)
				return "v", nil
			})
			if err != nil || stale || value != "v" {
				t.Errorf("unexpected result: value=%q stale=%v err=%v", value, stale, err)
			}
		}()
	}

	close(start)
	wg.Wait()

	if got := calls.Load(); got != 1 {
		t.Errorf("expected 1 upstream call for 50 concurrent clients, got %d", got)
	}
}

func TestServesCachedValueWithinTTL(t *testing.T) {
	c := New[string](time.Minute)

	var calls atomic.Int64
	get := func(context.Context) (string, error) {
		calls.Add(1)
		return "v", nil
	}

	if _, _, err := c.Get(context.Background(), "k", get); err != nil {
		t.Fatalf("first get failed: %v", err)
	}
	if _, _, err := c.Get(context.Background(), "k", get); err != nil {
		t.Fatalf("second get failed: %v", err)
	}

	if got := calls.Load(); got != 1 {
		t.Errorf("expected 1 upstream call within TTL, got %d", got)
	}
}

func TestReloadsAfterTTL(t *testing.T) {
	c := New[string](20 * time.Millisecond)

	var calls atomic.Int64
	get := func(context.Context) (string, error) {
		calls.Add(1)
		return "v", nil
	}

	if _, _, err := c.Get(context.Background(), "k", get); err != nil {
		t.Fatalf("first get failed: %v", err)
	}
	time.Sleep(40 * time.Millisecond)
	if _, _, err := c.Get(context.Background(), "k", get); err != nil {
		t.Fatalf("second get failed: %v", err)
	}

	if got := calls.Load(); got != 2 {
		t.Errorf("expected 2 upstream calls after TTL expiry, got %d", got)
	}
}

func TestReturnsStaleValueOnError(t *testing.T) {
	c := New[string](10 * time.Millisecond)

	if _, _, err := c.Get(context.Background(), "k", func(context.Context) (string, error) {
		return "old", nil
	}); err != nil {
		t.Fatalf("first get failed: %v", err)
	}

	time.Sleep(30 * time.Millisecond)

	value, stale, err := c.Get(context.Background(), "k", func(context.Context) (string, error) {
		return "", errors.New("upstream down")
	})
	if err == nil {
		t.Fatal("expected an error from a failing upstream")
	}
	if !stale {
		t.Error("expected the value to be reported as stale")
	}
	if value != "old" {
		t.Errorf("expected the last known value %q, got %q", "old", value)
	}
}

func TestDoesNotCacheErrors(t *testing.T) {
	c := New[string](time.Minute)

	var calls atomic.Int64
	failing := func(context.Context) (string, error) {
		calls.Add(1)
		return "", errors.New("upstream down")
	}

	for i := 0; i < 3; i++ {
		if _, _, err := c.Get(context.Background(), "k", failing); err == nil {
			t.Fatal("expected an error from a failing upstream")
		}
	}

	if got := calls.Load(); got != 3 {
		t.Errorf("expected 3 upstream calls (errors are not cached), got %d", got)
	}
}
