package config

import (
	"os"
	"strings"
)

const (
	defaultPort                  = "8080"
	defaultFrankfurterBaseURL    = "https://api.frankfurter.dev/v2"
	defaultRateCacheDatabasePath = "/tmp/currency-watcher/rates.db"
	defaultCORSAllowedOrigins    = "http://localhost:3000"
)

// Config contains process configuration for the HTTP server.
type Config struct {
	Port                  string
	FrankfurterBaseURL    string
	RateCacheDatabasePath string
	CORSAllowedOrigins    []string
}

// Load reads configuration from the environment.
func Load() Config {
	return Config{
		Port:                  envOrDefault("PORT", defaultPort),
		FrankfurterBaseURL:    envOrDefault("FRANKFURTER_BASE_URL", defaultFrankfurterBaseURL),
		RateCacheDatabasePath: envOrDefault("RATE_CACHE_DB_PATH", defaultRateCacheDatabasePath),
		CORSAllowedOrigins:    configuredOrigins(envOrDefault("CORS_ALLOWED_ORIGINS", defaultCORSAllowedOrigins)),
	}
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func configuredOrigins(value string) []string {
	origins := make([]string, 0)
	for _, origin := range strings.Split(value, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}
