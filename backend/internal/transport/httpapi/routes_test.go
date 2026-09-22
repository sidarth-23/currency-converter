package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/example/currency-watcher/backend/internal/currency"
)

type fakeRateSource struct {
	rates map[string]currency.CachedRate
	err   error
}

func (f fakeRateSource) Get(context.Context, string, []string) (map[string]currency.CachedRate, error) {
	return f.rates, f.err
}

type fakeCurrencySource struct {
	currencies []currency.Currency
	err        error
}

func (f fakeCurrencySource) FetchCurrencies(context.Context) ([]currency.Currency, error) {
	return f.currencies, f.err
}

func TestHealthAndRatesRoutes(t *testing.T) {
	expiresAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	mux := http.NewServeMux()
	NewAPI(mux, fakeRateSource{rates: map[string]currency.CachedRate{
		"EUR": {Rate: 0.85, ExpiresAt: expiresAt},
		"SGD": {Rate: 1.35, ExpiresAt: expiresAt},
	}}, fakeCurrencySource{})

	health := httptest.NewRecorder()
	mux.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if health.Code != http.StatusOK || !strings.Contains(health.Body.String(), `"status":"ok"`) {
		t.Fatalf("unexpected health response: %d %s", health.Code, health.Body.String())
	}

	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/rates?base=USD&targets=EUR,SGD", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected rates status: %d %s", response.Code, response.Body.String())
	}
	var output struct {
		Base  string                `json:"base"`
		Rates map[string]RateOutput `json:"rates"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &output); err != nil {
		t.Fatal(err)
	}
	if output.Base != "USD" || output.Rates["EUR"].Rate != 0.85 || !output.Rates["EUR"].ExpiresAt.Equal(expiresAt) {
		t.Fatalf("unexpected rates response: %#v", output)
	}
	if !strings.Contains(response.Body.String(), `"expiresAt":"2026-01-02T03:04:05Z"`) {
		t.Fatalf("expiry was not RFC 3339: %s", response.Body.String())
	}
}
func TestRatesValidationAndUpstreamError(t *testing.T) {
	for _, query := range []string{
		"base=usd&targets=EUR",
		"base=USD&targets=eur",
		"base=USD&targets=EUR,EUR",
		"base=USD&targets=",
	} {
		mux := http.NewServeMux()
		NewAPI(mux, fakeRateSource{rates: map[string]currency.CachedRate{"EUR": {Rate: 0.85}}}, fakeCurrencySource{})
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/rates?"+query, nil))
		if response.Code != http.StatusUnprocessableEntity {
			t.Fatalf("query %q returned %d: %s", query, response.Code, response.Body.String())
		}
	}

	mux := http.NewServeMux()
	NewAPI(mux, fakeRateSource{err: context.Canceled}, fakeCurrencySource{})
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/rates?base=USD&targets=EUR", nil))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("upstream error returned %d: %s", response.Code, response.Body.String())
	}
}

func TestCurrenciesRoute(t *testing.T) {
	mux := http.NewServeMux()
	NewAPI(mux, fakeRateSource{}, fakeCurrencySource{currencies: []currency.Currency{
		{Code: "EUR", Name: "Euro"},
	}})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/currencies", nil)
	request.Header.Set("Accept", "*/*")
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusOK ||
		response.Header().Get("Content-Type") != "application/json" ||
		!strings.Contains(response.Body.String(), `"code":"EUR"`) {
		t.Fatalf("unexpected currencies response: %d %s", response.Code, response.Body.String())
	}

	mux = http.NewServeMux()
	NewAPI(mux, fakeRateSource{}, fakeCurrencySource{err: context.Canceled})
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/currencies", nil))
	if response.Code != http.StatusInternalServerError ||
		!strings.Contains(response.Body.String(), "unable to fetch currencies") {
		t.Fatalf("unexpected currencies error: %d %s", response.Code, response.Body.String())
	}
}
