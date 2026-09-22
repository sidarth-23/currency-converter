package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/example/currency-watcher/backend/internal/frankfurterclient"
	"github.com/example/currency-watcher/backend/internal/httpapi"
	"github.com/example/currency-watcher/backend/internal/rates"
)

func main() {
	port := envOrDefault("PORT", "8080")
	baseURL := envOrDefault("FRANKFURTER_BASE_URL", "https://api.frankfurter.dev/v2")
	origins := configuredOrigins(envOrDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000"))

	client, err := frankfurterclient.NewClient(baseURL)
	if err != nil {
		log.Fatalf("create Frankfurter client: %v", err)
	}
	fetcher := rates.NewFrankfurterFetcher(client)
	cache := rates.NewCache(fetcher)
	mux := http.NewServeMux()
	httpapi.NewAPI(mux, cache)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsMiddleware(mux, origins),
	}
	log.Printf("currency watcher API listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func configuredOrigins(value string) map[string]struct{} {
	origins := make(map[string]struct{})
	for _, origin := range strings.Split(value, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			origins[origin] = struct{}{}
		}
	}
	return origins
}

func corsMiddleware(next http.Handler, origins map[string]struct{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, allowed := origins[origin]; allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
