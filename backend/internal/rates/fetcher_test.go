package rates

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/currency-watcher/backend/internal/frankfurterclient"
)

func TestFrankfurterFetcherUsesTypedRatesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("base") != "USD" || r.URL.Query().Get("quotes") != "EUR,SGD" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"date":"2026-01-01","base":"USD","quote":"USD","rate":1},{"date":"2026-01-01","base":"USD","quote":"EUR","rate":0.85},{"date":"2026-01-01","base":"USD","quote":"SGD","rate":1.35}]`))
	}))
	defer server.Close()
	client, err := frankfurterclient.NewClient(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	result, err := NewFrankfurterFetcher(client).Fetch(context.Background(), "USD", []string{"EUR", "SGD"})
	if err != nil {
		t.Fatal(err)
	}
	if result["EUR"] != 0.85 || result["SGD"] != 1.35 {
		t.Fatalf("unexpected rates: %#v", result)
	}
}

func TestFrankfurterFetcherRejectsUpstreamFailureAndMalformedRows(t *testing.T) {
	tests := []struct {
		name string
		body string
		code int
	}{
		{name: "upstream failure", body: `{"detail":"unavailable"}`, code: http.StatusServiceUnavailable},
		{name: "missing row", body: `[{"date":"2026-01-01","base":"USD","quote":"EUR","rate":0.85}]`, code: http.StatusOK},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(test.code)
				_, _ = w.Write([]byte(test.body))
			}))
			defer server.Close()
			client, err := frankfurterclient.NewClient(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := NewFrankfurterFetcher(client).Fetch(context.Background(), "USD", []string{"EUR", "SGD"}); err == nil {
				t.Fatal("expected fetch error")
			}
		})
	}
}
