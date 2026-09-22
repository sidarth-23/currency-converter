package rates

import (
	"context"
	"sync"
	"testing"
	"time"
)

type fakeFetcher struct {
	mu    sync.Mutex
	calls int
	block <-chan struct{}
}

func (f *fakeFetcher) Fetch(ctx context.Context, base string, targets []string) (map[string]float64, error) {
	f.mu.Lock()
	f.calls++
	f.mu.Unlock()
	if f.block != nil {
		select {
		case <-f.block:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return map[string]float64{"EUR": 0.85, "SGD": 1.35}, nil
}

func (f *fakeFetcher) Calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func TestCacheHitAndCopySafety(t *testing.T) {
	fetcher := &fakeFetcher{}
	cache := NewCacheWithTTL(fetcher, time.Hour)
	first, err := cache.Get(context.Background(), "usd", []string{"SGD", "EUR"})
	if err != nil {
		t.Fatal(err)
	}
	first["EUR"] = 123
	second, err := cache.Get(context.Background(), "USD", []string{"EUR", "SGD"})
	if err != nil {
		t.Fatal(err)
	}
	if second["EUR"] != 0.85 {
		t.Fatalf("cached result was mutated: %#v", second)
	}
	if fetcher.Calls() != 1 {
		t.Fatalf("expected one fetch, got %d", fetcher.Calls())
	}
}

func TestCacheExpiryRefetches(t *testing.T) {
	fetcher := &fakeFetcher{}
	cache := NewCacheWithTTL(fetcher, time.Hour)
	now := time.Now()
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

func TestCacheCoalescesConcurrentMisses(t *testing.T) {
	block := make(chan struct{})
	fetcher := &fakeFetcher{block: block}
	cache := NewCache(fetcher)
	const callers = 8
	results := make(chan map[string]float64, callers)
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
}
