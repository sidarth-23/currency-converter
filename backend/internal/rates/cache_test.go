package rates

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type fakeFetcher struct {
	mu       sync.Mutex
	calls    int
	requests [][]string
	block    <-chan struct{}
}

func (f *fakeFetcher) Fetch(ctx context.Context, _ string, targets []string) (map[string]float64, error) {
	f.mu.Lock()
	f.calls++
	f.requests = append(f.requests, append([]string(nil), targets...))
	f.mu.Unlock()
	if f.block != nil {
		select {
		case <-f.block:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	rates := make(map[string]float64, len(targets))
	for _, target := range targets {
		switch target {
		case "EUR":
			rates[target] = 0.85
		case "SGD":
			rates[target] = 1.35
		}
	}
	return rates, nil
}

func (f *fakeFetcher) Calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func (f *fakeFetcher) Requests() [][]string {
	f.mu.Lock()
	defer f.mu.Unlock()
	requests := make([][]string, len(f.requests))
	for i, request := range f.requests {
		requests[i] = append([]string(nil), request...)
	}
	return requests
}

func newTestCache(t *testing.T, fetcher RateFetcher, ttl time.Duration) *Cache {
	t.Helper()
	cache, err := NewCacheWithTTL(fetcher, filepath.Join(t.TempDir(), "rates.db"), ttl)
	if err != nil {
		t.Fatal(err)
	}
	return cache
}

func TestCacheHitAndCopySafety(t *testing.T) {
	fetcher := &fakeFetcher{}
	cache := newTestCache(t, fetcher, time.Hour)
	first, err := cache.Get(context.Background(), "USD", []string{"SGD", "EUR"})
	if err != nil {
		t.Fatal(err)
	}
	first["EUR"] = CachedRate{Rate: 123}

	second, err := cache.Get(context.Background(), "USD", []string{"EUR", "SGD"})
	if err != nil {
		t.Fatal(err)
	}
	if second["EUR"].Rate != 0.85 {
		t.Fatalf("cached result was mutated: %#v", second)
	}
	if fetcher.Calls() != 1 {
		t.Fatalf("expected one fetch, got %d", fetcher.Calls())
	}
	requests := fetcher.Requests()
	if len(requests) != 1 || len(requests[0]) != 2 || requests[0][0] != "SGD" || requests[0][1] != "EUR" {
		t.Fatalf("fetcher received normalized request: %#v", requests)
	}
}

func TestCacheExpiryRefetchesAfterOneHourFallback(t *testing.T) {
	fetcher := &fakeFetcher{}
	cache := newTestCache(t, fetcher, 0)
	now := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	cache.now = func() time.Time { return now }

	if _, err := cache.Get(context.Background(), "USD", []string{"EUR"}); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Hour)
	if _, err := cache.Get(context.Background(), "USD", []string{"EUR"}); err != nil {
		t.Fatal(err)
	}
	if fetcher.Calls() != 2 {
		t.Fatalf("expected two fetches after expiry, got %d", fetcher.Calls())
	}
}

func TestCacheFetchesOnlyMissingTargets(t *testing.T) {
	fetcher := &fakeFetcher{}
	cache := newTestCache(t, fetcher, time.Hour)
	now := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	cache.now = func() time.Time { return now }

	first, err := cache.Get(context.Background(), "USD", []string{"EUR"})
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(5 * time.Minute)
	second, err := cache.Get(context.Background(), "USD", []string{"EUR", "SGD"})
	if err != nil {
		t.Fatal(err)
	}

	requests := fetcher.Requests()
	if len(requests) != 2 || len(requests[1]) != 1 || requests[1][0] != "SGD" {
		t.Fatalf("expected only missing SGD to be fetched, got %#v", requests)
	}
	if second["EUR"] != first["EUR"] {
		t.Fatalf("expected cached EUR to be retained, got %#v after %#v", second["EUR"], first["EUR"])
	}
	if second["SGD"].Rate != 1.35 {
		t.Fatalf("missing SGD result: %#v", second)
	}
	if !second["EUR"].ExpiresAt.Before(second["SGD"].ExpiresAt) {
		t.Fatalf("expected target expiries to be independent: %#v", second)
	}
}

func TestCacheCoalescesConcurrentMisses(t *testing.T) {
	block := make(chan struct{})
	fetcher := &fakeFetcher{block: block}
	cache, err := NewCache(fetcher, filepath.Join(t.TempDir(), "rates.db"))
	if err != nil {
		t.Fatal(err)
	}
	const callers = 8
	results := make(chan map[string]CachedRate, callers)
	errs := make(chan error, callers)
	var group sync.WaitGroup
	for range callers {
		group.Go(func() {
			result, err := cache.Get(context.Background(), "USD", []string{"EUR", "SGD"})
			results <- result
			errs <- err
		})
	}
	for fetcher.Calls() == 0 {
		time.Sleep(time.Millisecond)
	}
	close(block)
	group.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if fetcher.Calls() != 1 {
		t.Fatalf("expected one coalesced fetch, got %d", fetcher.Calls())
	}
	for result := range results {
		if result["EUR"].Rate != 0.85 || result["SGD"].Rate != 1.35 {
			t.Fatalf("unexpected coalesced result: %#v", result)
		}
	}
}
