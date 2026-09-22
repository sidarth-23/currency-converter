package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeRateSource struct {
	rates map[string]float64
	err   error
}

func (f fakeRateSource) Get(context.Context, string, []string) (map[string]float64, error) {
	return f.rates, f.err
}

func TestHealthAndRatesRoutes(t *testing.T) {
	mux := http.NewServeMux()
	NewAPI(mux, fakeRateSource{rates: map[string]float64{"EUR": 0.85, "SGD": 1.35}})

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
	for _, query := range []string{"base=usd&targets=EUR", "base=USD&targets=EUR,EUR", "base=USD&targets="} {
		mux := http.NewServeMux()
		NewAPI(mux, fakeRateSource{rates: map[string]float64{"EUR": 0.85}})
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/rates?"+query, nil))
		if response.Code != http.StatusUnprocessableEntity {
			t.Fatalf("query %q returned %d: %s", query, response.Code, response.Body.String())
		}
	}

	mux := http.NewServeMux()
	NewAPI(mux, fakeRateSource{err: context.Canceled})
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/rates?base=USD&targets=EUR", nil))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("upstream error returned %d: %s", response.Code, response.Body.String())
	}
}
