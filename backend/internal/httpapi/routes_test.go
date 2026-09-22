package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/currency-watcher/backend/internal/rates"
)

type fakeRateSource struct {
	rates map[string]float64
	err   error
}

func (f fakeRateSource) Get(context.Context, string, []string) (map[string]float64, error) {
	return f.rates, f.err
}

type fakeCurrencySource struct {
	currencies []rates.Currency
	err        error
}

func (f fakeCurrencySource) FetchCurrencies(context.Context) ([]rates.Currency, error) {
	return f.currencies, f.err
}

func TestHealthAndRatesRoutes(t *testing.T) {
	mux := http.NewServeMux()
	NewAPI(mux, fakeRateSource{rates: map[string]float64{"EUR": 0.85, "SGD": 1.35}}, fakeCurrencySource{})

	health := httptest.NewRecorder()
	mux.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if health.Code != http.StatusOK || !strings.Contains(health.Body.String(), `"status":"ok"`) {
		t.Fatalf("unexpected health response: %d %s", health.Code, health.Body.String())
	}

	rates := httptest.NewRecorder()
	mux.ServeHTTP(rates, httptest.NewRequest(http.MethodGet, "/api/rates?base=USD&targets=EUR,SGD", nil))
	if rates.Code != http.StatusOK {
		t.Fatalf("unexpected rates status: %d %s", rates.Code, rates.Body.String())
	}
	if !strings.Contains(rates.Body.String(), `"base":"USD"`) || !strings.Contains(rates.Body.String(), `"EUR":0.85`) {
		t.Fatalf("unexpected rates response: %s", rates.Body.String())
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
		NewAPI(mux, fakeRateSource{rates: map[string]float64{"EUR": 0.85}}, fakeCurrencySource{})
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
	NewAPI(mux, fakeRateSource{}, fakeCurrencySource{currencies: []rates.Currency{
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
