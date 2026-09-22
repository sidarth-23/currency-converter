package rates

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RateCacheEntry is a persisted exchange rate for one base-target pair.
type RateCacheEntry struct {
	Base      string `gorm:"primaryKey"`
	Target    string `gorm:"primaryKey"`
	Rate      float64
	ExpiresAt time.Time `gorm:"index"`
}

// CachedRate is an exchange rate and the time it becomes stale.
type CachedRate struct {
	Rate      float64
	ExpiresAt time.Time
}

// Cache fetches rates and retains successful target-level results for a fixed TTL.
type Cache struct {
	fetcher RateFetcher
	ttl     time.Duration
	now     func() time.Time
	db      *gorm.DB

	mu       sync.Mutex
	inFlight map[string]*cacheCall
}

type cacheCall struct {
	done  chan struct{}
	rates map[string]CachedRate
	err   error
}

func NewCache(fetcher RateFetcher, databasePath string) (*Cache, error) {
	return NewCacheWithTTL(fetcher, databasePath, time.Hour)
}

func NewCacheWithTTL(fetcher RateFetcher, databasePath string, ttl time.Duration) (*Cache, error) {
	if ttl <= 0 {
		ttl = time.Hour
	}
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		return nil, fmt.Errorf("create rate cache directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open rate cache database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("access rate cache database: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&RateCacheEntry{}); err != nil {
		return nil, fmt.Errorf("migrate rate cache database: %w", err)
	}

	return &Cache{
		fetcher:  fetcher,
		ttl:      ttl,
		now:      time.Now,
		db:       db,
		inFlight: make(map[string]*cacheCall),
	}, nil
}

// Get returns a copy of the cached or newly fetched rates.
func (c *Cache) Get(ctx context.Context, base string, targets []string) (map[string]CachedRate, error) {
	if c == nil || c.fetcher == nil {
		return nil, fmt.Errorf("rate cache fetcher is nil")
	}
	if c.db == nil {
		return nil, fmt.Errorf("rate cache database is nil")
	}
	key := cacheKey(base, targets)

	c.mu.Lock()
	if call, ok := c.inFlight[key]; ok {
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-call.done:
			return cloneCachedRates(call.rates), call.err
		}
	}
	call := &cacheCall{done: make(chan struct{})}
	c.inFlight[key] = call
	c.mu.Unlock()

	cachedRates, err := c.get(ctx, base, targets)

	c.mu.Lock()
	call.rates = cloneCachedRates(cachedRates)
	call.err = err
	delete(c.inFlight, key)
	close(call.done)
	c.mu.Unlock()

	return cloneCachedRates(cachedRates), err
}

func (c *Cache) get(ctx context.Context, base string, targets []string) (map[string]CachedRate, error) {
	now := c.now()
	var entries []RateCacheEntry
	if err := c.db.WithContext(ctx).
		Where("base = ? AND target IN ?", base, targets).
		Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("read rate cache: %w", err)
	}

	rates := make(map[string]CachedRate, len(targets))
	for _, entry := range entries {
		if entry.ExpiresAt.After(now) {
			rates[entry.Target] = CachedRate{Rate: entry.Rate, ExpiresAt: entry.ExpiresAt}
		}
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

	fetched, err := c.fetcher.Fetch(ctx, base, missingTargets)
	if err != nil {
		return nil, err
	}
	if err := validateFetchedRates(fetched, missingTargets); err != nil {
		return nil, err
	}

	expiresAt := now.Add(c.ttl).UTC()
	newEntries := make([]RateCacheEntry, 0, len(missingTargets))
	for _, target := range missingTargets {
		rate := fetched[target]
		rates[target] = CachedRate{Rate: rate, ExpiresAt: expiresAt}
		newEntries = append(newEntries, RateCacheEntry{
			Base:      base,
			Target:    target,
			Rate:      rate,
			ExpiresAt: expiresAt,
		})
	}
	if err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "base"}, {Name: "target"}},
			DoUpdates: clause.AssignmentColumns([]string{"rate", "expires_at"}),
		}).Create(&newEntries).Error
	}); err != nil {
		return nil, fmt.Errorf("write rate cache: %w", err)
	}
	return rates, nil
}

func validateFetchedRates(fetched map[string]float64, targets []string) error {
	for _, target := range targets {
		if _, ok := fetched[target]; !ok {
			return fmt.Errorf("rate fetcher returned no rate for %q", target)
		}
	}
	return nil
}

func cloneCachedRates(rates map[string]CachedRate) map[string]CachedRate {
	if rates == nil {
		return nil
	}
	copyRates := make(map[string]CachedRate, len(rates))
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
