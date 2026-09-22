package rates

import (
	"context"
	"fmt"
	"strings"

	"github.com/example/currency-watcher/backend/internal/generated"
)

// RateFetcher retrieves rates from an upstream provider.
type RateFetcher interface {
	Fetch(ctx context.Context, base string, targets []string) (map[string]float64, error)
}

// Currency is a provider-supported ISO currency.
type Currency struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// FrankfurterFetcher adapts the generated Frankfurter client to RateFetcher.
type FrankfurterFetcher struct {
	client *generated.Client
}

func NewFrankfurterFetcher(client *generated.Client) *FrankfurterFetcher {
	return &FrankfurterFetcher{client: client}
}

func (f *FrankfurterFetcher) Fetch(ctx context.Context, base string, targets []string) (map[string]float64, error) {
	if f == nil || f.client == nil {
		return nil, fmt.Errorf("frankfurter client is nil")
	}
	params := generated.GetRatesParams{
		Base:   generated.NewOptString(base),
		Quotes: generated.NewOptString(strings.Join(targets, ",")),
	}
	result, err := f.client.GetRates(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetch frankfurter rates: %w", err)
	}
	jsonResult, ok := result.(*generated.GetRatesOKApplicationJSON)
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

// FetchCurrencies retrieves the active currencies supported by Frankfurter.
func (f *FrankfurterFetcher) FetchCurrencies(ctx context.Context) ([]Currency, error) {
	if f == nil || f.client == nil {
		return nil, fmt.Errorf("frankfurter client is nil")
	}
	result, err := f.client.GetCurrencies(ctx, generated.GetCurrenciesParams{})
	if err != nil {
		return nil, fmt.Errorf("fetch frankfurter currencies: %w", err)
	}
	jsonResult, ok := result.(*generated.GetCurrenciesOKApplicationJSON)
	if !ok || jsonResult == nil {
		return nil, fmt.Errorf("fetch frankfurter currencies: unsupported response")
	}

	seen := make(map[string]struct{}, len(*jsonResult))
	currencies := make([]Currency, 0, len(*jsonResult))
	for _, row := range *jsonResult {
		if _, exists := seen[row.IsoCode]; exists {
			return nil, fmt.Errorf("fetch frankfurter currencies: duplicate code %q", row.IsoCode)
		}
		seen[row.IsoCode] = struct{}{}
		currencies = append(currencies, Currency{Code: row.IsoCode, Name: row.Name})
	}
	return currencies, nil
}
