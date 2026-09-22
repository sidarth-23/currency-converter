package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/example/currency-watcher/backend/internal/adapters/frankfurter"
	"github.com/example/currency-watcher/backend/internal/adapters/frankfurter/generated"
	sqliteadapter "github.com/example/currency-watcher/backend/internal/adapters/sqlite"
	"github.com/example/currency-watcher/backend/internal/config"
	"github.com/example/currency-watcher/backend/internal/currency"
	"github.com/example/currency-watcher/backend/internal/transport/httpapi"
)

// NewHTTPServer composes the HTTP API and its concrete dependencies.
func NewHTTPServer(cfg config.Config) (*http.Server, error) {
	client, err := generated.NewClient(cfg.FrankfurterBaseURL)
	if err != nil {
		return nil, fmt.Errorf("create Frankfurter client: %w", err)
	}
	provider := frankfurter.NewProvider(client)
	store, err := sqliteadapter.NewRateStore(cfg.RateCacheDatabasePath)
	if err != nil {
		return nil, fmt.Errorf("create rate cache: %w", err)
	}
	rateService := currency.NewRateService(provider, store, time.Hour)
	mux := http.NewServeMux()
	httpapi.NewAPI(mux, rateService, provider)

	return &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: httpapi.WithCORS(mux, cfg.CORSAllowedOrigins),
	}, nil
}
