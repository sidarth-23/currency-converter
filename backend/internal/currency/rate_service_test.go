package currency

import (
	"context"
	"sync"
	"testing"
	"time"
)

type fakeProvider struct {
	mu       sync.Mutex
	calls    int
	requests [][]string
	block    <-chan struct{}
	started  chan struct{}
}

func (p *fakeProvider) Fetch(ctx context.Context, _ string, targets []string) (map[string]float64, error) {
	p.mu.Lock()
	p.calls++
	p.requests = append(p.requests, append([]string(nil), targets...))
	p.mu.Unlock()
	if p.started != nil {
		p.started <- struct{}{}
	}
	if p.block != nil {
		select {
		case <-p.block:
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

func (p *fakeProvider) Calls() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls
}

func (p *fakeProvider) Requests() [][]string {
	p.mu.Lock()
	defer p.mu.Unlock()
	requests := make([][]string, len(p.requests))
	for i, request := range p.requests {
		requests[i] = append([]string(nil), request...)
	}
	return requests
}

type memoryStore struct {
	mu    sync.Mutex
	rates map[string]map[string]CachedRate
}

func (s *memoryStore) FindFresh(_ context.Context, base string, targets []string, now time.Time) (map[string]CachedRate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fresh := make(map[string]CachedRate, len(targets))
	for _, target := range targets {
		if rate, ok := s.rates[base][target]; ok && rate.ExpiresAt.After(now) {
			fresh[target] = rate
		}
	}
	return fresh, nil
}

func (s *memoryStore) Upsert(_ context.Context, base string, rates map[string]CachedRate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rates == nil {
		s.rates = make(map[string]map[string]CachedRate)
	}
	if s.rates[base] == nil {
		s.rates[base] = make(map[string]CachedRate)
	}
	for target, rate := range rates {
		s.rates[base][target] = rate
	}
	return nil
}

func newTestRateService(provider RateProvider, ttl time.Duration) *RateService {
	return NewRateService(provider, &memoryStore{}, ttl)
}

func TestRateServiceCacheHitAndCopySafety(t *testing.T) {
	provider := &fakeProvider{}
	service := newTestRateService(provider, time.Hour)
	first, err := service.Get(context.Background(), "USD", []string{"SGD", "EUR"})
	if err != nil {
		t.Fatal(err)
	}
	first["EUR"] = CachedRate{Rate: 123}

	second, err := service.Get(context.Background(), "USD", []string{"EUR", "SGD"})
	if err != nil {
		t.Fatal(err)
	}
	if second["EUR"].Rate != 0.85 {
		t.Fatalf("cached result was mutated: %#v", second)
	}
	if provider.Calls() != 1 {
		t.Fatalf("expected one fetch, got %d", provider.Calls())
	}
	requests := provider.Requests()
	if len(requests) != 1 || len(requests[0]) != 2 || requests[0][0] != "SGD" || requests[0][1] != "EUR" {
		t.Fatalf("provider received normalized request: %#v", requests)
	}
}

func TestRateServiceExpiryRefetchesAfterOneHourFallback(t *testing.T) {
	provider := &fakeProvider{}
	service := newTestRateService(provider, 0)
	now := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	service.now = func() time.Time { return now }

	if _, err := service.Get(context.Background(), "USD", []string{"EUR"}); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Hour)
	if _, err := service.Get(context.Background(), "USD", []string{"EUR"}); err != nil {
		t.Fatal(err)
	}
	if provider.Calls() != 2 {
		t.Fatalf("expected two fetches after expiry, got %d", provider.Calls())
	}
}

func TestRateServiceFetchesOnlyMissingTargets(t *testing.T) {
	provider := &fakeProvider{}
	service := newTestRateService(provider, time.Hour)
	now := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	service.now = func() time.Time { return now }

	first, err := service.Get(context.Background(), "USD", []string{"EUR"})
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(5 * time.Minute)
	second, err := service.Get(context.Background(), "USD", []string{"EUR", "SGD"})
	if err != nil {
		t.Fatal(err)
	}

	requests := provider.Requests()
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

func TestRateServiceCoalescesConcurrentMisses(t *testing.T) {
	block := make(chan struct{})
	provider := &fakeProvider{block: block, started: make(chan struct{}, 1)}
	service := newTestRateService(provider, time.Hour)
	const callers = 8
	results := make(chan map[string]CachedRate, callers)
	errs := make(chan error, callers)
	var group sync.WaitGroup
	for range callers {
		group.Go(func() {
			result, err := service.Get(context.Background(), "USD", []string{"EUR", "SGD"})
			results <- result
			errs <- err
		})
	}
	<-provider.started
	close(block)
	group.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if provider.Calls() != 1 {
		t.Fatalf("expected one coalesced fetch, got %d", provider.Calls())
	}
	for result := range results {
		if result["EUR"].Rate != 0.85 || result["SGD"].Rate != 1.35 {
			t.Fatalf("unexpected coalesced result: %#v", result)
		}
	}
}

func TestRateServiceCanceledWaiterReturnsContextError(t *testing.T) {
	block := make(chan struct{})
	started := make(chan struct{}, 1)
	provider := &fakeProvider{block: block, started: started}
	service := newTestRateService(provider, time.Hour)
	leaderDone := make(chan error, 1)
	go func() {
		_, err := service.Get(context.Background(), "USD", []string{"EUR"})
		leaderDone <- err
	}()
	<-started

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := service.Get(ctx, "USD", []string{"EUR"}); err != context.Canceled {
		t.Fatalf("expected canceled waiter error, got %v", err)
	}
	close(block)
	if err := <-leaderDone; err != nil {
		t.Fatal(err)
	}
	if provider.Calls() != 1 {
		t.Fatalf("expected one fetch, got %d", provider.Calls())
	}
}
