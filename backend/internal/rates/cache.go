package rates

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Cache fetches rates and retains complete successful results for a fixed TTL.
type Cache struct {
	fetcher RateFetcher
	ttl     time.Duration
	now     func() time.Time

	mu       sync.Mutex
	entries  map[string]cacheEntry
	inFlight map[string]*cacheCall
}

type cacheEntry struct {
	rates     map[string]float64
	expiresAt time.Time
}

type cacheCall struct {
	done  chan struct{}
	rates map[string]float64
	err   error
}

func NewCache(fetcher RateFetcher) *Cache {
	return NewCacheWithTTL(fetcher, time.Hour)
}

func NewCacheWithTTL(fetcher RateFetcher, ttl time.Duration) *Cache {
	if ttl <= 0 {
		ttl = time.Hour
	}
	return &Cache{
		fetcher:  fetcher,
		ttl:      ttl,
		now:      time.Now,
		entries:  make(map[string]cacheEntry),
		inFlight: make(map[string]*cacheCall),
	}
}

// Get returns a copy of the cached or newly fetched rates.
func (c *Cache) Get(ctx context.Context, base string, targets []string) (map[string]float64, error) {
	if c == nil || c.fetcher == nil {
		return nil, fmt.Errorf("rate cache fetcher is nil")
	}
	key := cacheKey(base, targets)

	c.mu.Lock()
	if entry, ok := c.entries[key]; ok && c.now().Before(entry.expiresAt) {
		rates := cloneRates(entry.rates)
		c.mu.Unlock()
		return rates, nil
	}
	if call, ok := c.inFlight[key]; ok {
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-call.done:
			return cloneRates(call.rates), call.err
		}
	}
	call := &cacheCall{done: make(chan struct{})}
	c.inFlight[key] = call
	c.mu.Unlock()

	rates, fetchErr := c.fetcher.Fetch(ctx, base, targets)
	if fetchErr == nil {
		rates = cloneRates(rates)
	}

	c.mu.Lock()
	call.rates = cloneRates(rates)
	call.err = fetchErr
	if fetchErr == nil {
		c.entries[key] = cacheEntry{rates: cloneRates(rates), expiresAt: c.now().Add(c.ttl)}
	}
	delete(c.inFlight, key)
	close(call.done)
	c.mu.Unlock()
	return cloneRates(rates), fetchErr
}

func cloneRates(rates map[string]float64) map[string]float64 {
	if rates == nil {
		return nil
	}
	copyRates := make(map[string]float64, len(rates))
	for key, value := range rates {
		copyRates[key] = value
	}
	return copyRates
}

func cacheKey(base string, targets []string) string {
	sortedTargets := append([]string(nil), targets...)
	sort.Strings(sortedTargets)
	return base + "|" + strings.Join(sortedTargets, ",")
}
