package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/example/currency-watcher/backend/internal/currency"
)

func newTestRateStore(t *testing.T) *RateStore {
	t.Helper()
	store, err := NewRateStore(filepath.Join(t.TempDir(), "rates.db"))
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestRateStoreFindFreshExcludesExpiredEntries(t *testing.T) {
	store := newTestRateStore(t)
	now := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	if err := store.Upsert(context.Background(), "USD", map[string]currency.CachedRate{
		"EUR": {Rate: 0.85, ExpiresAt: now.Add(time.Hour)},
		"SGD": {Rate: 1.35, ExpiresAt: now},
	}); err != nil {
		t.Fatal(err)
	}

	rates, err := store.FindFresh(context.Background(), "USD", []string{"EUR", "SGD"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(rates) != 1 || rates["EUR"].Rate != 0.85 || !rates["EUR"].ExpiresAt.Equal(now.Add(time.Hour)) {
		t.Fatalf("unexpected fresh rates: %#v", rates)
	}
}

func TestRateStoreUpsertReplacesBaseTargetEntry(t *testing.T) {
	store := newTestRateStore(t)
	now := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	if err := store.Upsert(context.Background(), "USD", map[string]currency.CachedRate{
		"EUR": {Rate: 0.85, ExpiresAt: now.Add(time.Hour)},
	}); err != nil {
		t.Fatal(err)
	}
	updated := currency.CachedRate{Rate: 0.86, ExpiresAt: now.Add(2 * time.Hour)}
	if err := store.Upsert(context.Background(), "USD", map[string]currency.CachedRate{"EUR": updated}); err != nil {
		t.Fatal(err)
	}

	rates, err := store.FindFresh(context.Background(), "USD", []string{"EUR"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(rates) != 1 || rates["EUR"] != updated {
		t.Fatalf("unexpected upserted rates: %#v", rates)
	}
}
