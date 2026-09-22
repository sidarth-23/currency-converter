package currency

import (
	"maps"
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// RateService fetches rates and retains successful target-level results for a fixed TTL.
type RateService struct {
	provider RateProvider
	store    RateStore
	ttl      time.Duration
	now      func() time.Time

	mu       sync.Mutex
	inFlight map[string]*rateCall
}

type rateCall struct {
	done  chan struct{}
	rates map[string]CachedRate
	err   error
}

// NewRateService creates a rate service with the supplied dependencies.
func NewRateService(provider RateProvider, store RateStore, ttl time.Duration) *RateService {
	if ttl <= 0 {
		ttl = time.Hour
	}
	return &RateService{
		provider: provider,
		store:    store,
		ttl:      ttl,
		now:      time.Now,
		inFlight: make(map[string]*rateCall),
	}
}

// Get returns a copy of the cached or newly fetched rates.
func (s *RateService) Get(ctx context.Context, base string, targets []string) (map[string]CachedRate, error) {
	if s == nil || s.provider == nil {
		return nil, fmt.Errorf("rate provider is nil")
	}
	if s.store == nil {
		return nil, fmt.Errorf("rate store is nil")
	}
	key := rateKey(base, targets)

	s.mu.Lock()
	if call, ok := s.inFlight[key]; ok {
		s.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-call.done:
			return cloneCachedRates(call.rates), call.err
		}
	}
	call := &rateCall{done: make(chan struct{})}
	s.inFlight[key] = call
	s.mu.Unlock()

	cachedRates, err := s.get(ctx, base, targets)

	s.mu.Lock()
	call.rates = cloneCachedRates(cachedRates)
	call.err = err
	delete(s.inFlight, key)
	close(call.done)
	s.mu.Unlock()

	return cloneCachedRates(cachedRates), err
}

// get reads fresh cached rates, fetches and validates missing targets, then persists the new entries.
func (s *RateService) get(ctx context.Context, base string, targets []string) (map[string]CachedRate, error) {
	now := s.now()
	rates, err := s.store.FindFresh(ctx, base, targets, now)
	if err != nil {
		return nil, err
	}

	missingTargets := make([]string, 0, len(targets))
	for _, target := range targets {
		if _, ok := rates[target]; !ok {
			missingTargets = append(missingTargets, target)
		}
	}
	if len(missingTargets) == 0 {
		return rates, nil
	}

	fetched, err := s.provider.Fetch(ctx, base, missingTargets)
	if err != nil {
		return nil, err
	}
	if err := validateFetchedRates(fetched, missingTargets); err != nil {
		return nil, err
	}

	expiresAt := now.Add(s.ttl).UTC()
	newRates := make(map[string]CachedRate, len(missingTargets))
	for _, target := range missingTargets {
		rate := CachedRate{Rate: fetched[target], ExpiresAt: expiresAt}
		rates[target] = rate
		newRates[target] = rate
	}
	if err := s.store.Upsert(ctx, base, newRates); err != nil {
		return nil, err
	}
	return rates, nil
}

// validateFetchedRates rejects provider results that omit any requested target.
func validateFetchedRates(fetched map[string]float64, targets []string) error {
	for _, target := range targets {
		if _, ok := fetched[target]; !ok {
			return fmt.Errorf("rate fetcher returned no rate for %q", target)
		}
	}
	return nil
}

// cloneCachedRates returns an independent copy so callers cannot mutate shared rate maps.
func cloneCachedRates(rates map[string]CachedRate) map[string]CachedRate {
	if rates == nil {
		return nil
	}
	copyRates := make(map[string]CachedRate, len(rates))
	maps.Copy(copyRates, rates)
	return copyRates
}

// rateKey builds an order-independent key for coalescing requests with the same base and targets.
func rateKey(base string, targets []string) string {
	sortedTargets := append([]string(nil), targets...)
	sort.Strings(sortedTargets)
	return base + "|" + strings.Join(sortedTargets, ",")
}
