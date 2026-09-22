package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/example/currency-watcher/backend/internal/currency"
)

// RateSource provides cached exchange rates to HTTP handlers.
type RateSource interface {
	Get(ctx context.Context, base string, targets []string) (map[string]currency.CachedRate, error)
}

// CurrencySource provides provider-supported currencies to HTTP handlers.
type CurrencySource interface {
	FetchCurrencies(ctx context.Context) ([]currency.Currency, error)
}

type HealthOutput struct {
	Body struct {
		Status string `json:"status"`
	}
}

type currencyCode string

// Schema describes currencyCode as three uppercase ASCII letters in the generated HTTP schema.
func (currencyCode) Schema(huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type:        "string",
		Pattern:     "^[A-Z]{3}$",
		Description: "three uppercase ASCII letters",
	}
}

type RatesInput struct {
	Base    currencyCode   `query:"base" required:"true"`
	Targets []currencyCode `query:"targets" required:"true" minItems:"1" uniqueItems:"true"`
}

type RateOutput struct {
	Rate      float64   `json:"rate"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type RatesOutput struct {
	ContentType string `header:"Content-Type"`
	Body        struct {
		Base  string                `json:"base"`
		Rates map[string]RateOutput `json:"rates"`
	}
}

type CurrenciesOutput struct {
	ContentType string              `header:"Content-Type"`
	Body        []currency.Currency `json:"body"`
}

// NewAPI builds the service routes on the provided standard-library mux.
func NewAPI(mux *http.ServeMux, rateSource RateSource, currencySource CurrencySource) huma.API {
	api := humago.NewWithPrefix(mux, "/api", huma.DefaultConfig("Currency Watcher API", "0.1.0"))
	huma.Register(api, huma.Operation{
		OperationID: "getHealth",
		Method:      http.MethodGet,
		Path:        "/health",
	}, func(context.Context, *struct{}) (*HealthOutput, error) {
		output := &HealthOutput{}
		output.Body.Status = "ok"
		return output, nil
	})
	huma.Register(api, huma.Operation{
		OperationID: "getRates",
		Method:      http.MethodGet,
		Path:        "/rates",
	}, ratesHandler(rateSource))
	huma.Register(api, huma.Operation{
		OperationID: "getCurrencies",
		Method:      http.MethodGet,
		Path:        "/currencies",
	}, currenciesHandler(currencySource))
	return api
}

// ratesHandler maps validated rate queries to JSON results and source failures to HTTP 500 responses.
func ratesHandler(rateSource RateSource) func(context.Context, *RatesInput) (*RatesOutput, error) {
	return func(ctx context.Context, input *RatesInput) (*RatesOutput, error) {
		targets := make([]string, len(input.Targets))
		for i, target := range input.Targets {
			targets[i] = string(target)
		}
		result, err := rateSource.Get(ctx, string(input.Base), targets)
		if err != nil {
			return nil, huma.Error500InternalServerError("unable to fetch exchange rates")
		}
		output := &RatesOutput{ContentType: "application/json"}
		output.Body.Base = string(input.Base)
		output.Body.Rates = make(map[string]RateOutput, len(result))
		for target, cachedRate := range result {
			output.Body.Rates[target] = RateOutput{
				Rate:      cachedRate.Rate,
				ExpiresAt: cachedRate.ExpiresAt,
			}
		}
		return output, nil
	}
}

// currenciesHandler maps provider currencies to JSON results and source failures to HTTP 500 responses.
func currenciesHandler(currencySource CurrencySource) func(context.Context, *struct{}) (*CurrenciesOutput, error) {
	return func(ctx context.Context, _ *struct{}) (*CurrenciesOutput, error) {
		currencies, err := currencySource.FetchCurrencies(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("unable to fetch currencies")
		}
		return &CurrenciesOutput{ContentType: "application/json", Body: currencies}, nil
	}
}
