package currency

import (
	"context"
	"time"
)

// RateProvider retrieves rates from an upstream provider.
type RateProvider interface {
	Fetch(ctx context.Context, base string, targets []string) (map[string]float64, error)
}

// RateStore retains cached exchange rates.
type RateStore interface {
	FindFresh(ctx context.Context, base string, targets []string, now time.Time) (map[string]CachedRate, error)
	Upsert(ctx context.Context, base string, rates map[string]CachedRate) error
}
