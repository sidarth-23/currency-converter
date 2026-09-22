package rates

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/example/currency-watcher/backend/internal/frankfurterclient"
)

// RateFetcher retrieves rates from an upstream provider.
type RateFetcher interface {
	Fetch(ctx context.Context, base string, targets []string) (map[string]float64, error)
}

// FrankfurterFetcher adapts the generated Frankfurter client to RateFetcher.
type FrankfurterFetcher struct {
	client *frankfurterclient.Client
}

func NewFrankfurterFetcher(client *frankfurterclient.Client) *FrankfurterFetcher {
	return &FrankfurterFetcher{client: client}
}

func (f *FrankfurterFetcher) Fetch(ctx context.Context, base string, targets []string) (map[string]float64, error) {
	if f == nil || f.client == nil {
		return nil, fmt.Errorf("frankfurter client is nil")
	}
	params := frankfurterclient.GetRatesParams{
		Base:   frankfurterclient.NewOptString(base),
		Quotes: frankfurterclient.NewOptString(strings.Join(targets, ",")),
	}
	result, err := f.client.GetRates(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetch frankfurter rates: %w", err)
	}
	jsonResult, ok := result.(*frankfurterclient.GetRatesOKApplicationJSON)
	if !ok || jsonResult == nil {
		return nil, fmt.Errorf("fetch frankfurter rates: unsupported response")
	}

	rates := make(map[string]float64, len(targets))
	for _, row := range *jsonResult {
		if row.Base == base && row.Quote == base && row.Rate == 1 {
			continue
		}
		if row.Base != base {
			continue
		}
		if _, exists := rates[row.Quote]; exists {
			return nil, fmt.Errorf("fetch frankfurter rates: duplicate quote %q", row.Quote)
		}
		rates[row.Quote] = row.Rate
	}
	for _, target := range targets {
		if _, ok := rates[target]; !ok {
			return nil, fmt.Errorf("fetch frankfurter rates: missing quote %q", target)
		}
	}
	return rates, nil
}

func normalizeKey(base string, targets []string) (string, string, []string, error) {
	normalizedBase := strings.ToUpper(strings.TrimSpace(base))
	if normalizedBase == "" {
		return "", "", nil, fmt.Errorf("base currency is empty")
	}
	seen := make(map[string]struct{}, len(targets))
	normalizedTargets := make([]string, 0, len(targets))
	for _, target := range targets {
		normalized := strings.ToUpper(strings.TrimSpace(target))
		if normalized == "" {
			return "", "", nil, fmt.Errorf("target currency is empty")
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		normalizedTargets = append(normalizedTargets, normalized)
	}
	if len(normalizedTargets) == 0 {
		return "", "", nil, fmt.Errorf("target currencies are empty")
	}
	sort.Strings(normalizedTargets)
	return normalizedBase + "|" + strings.Join(normalizedTargets, ","), normalizedBase, normalizedTargets, nil
}
