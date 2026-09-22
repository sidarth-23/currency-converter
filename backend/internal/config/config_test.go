package config

import (
	"reflect"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("FRANKFURTER_BASE_URL", "")
	t.Setenv("RATE_CACHE_DB_PATH", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	if got, want := Load(), (Config{
		Port:                  "8080",
		FrankfurterBaseURL:    "https://api.frankfurter.dev/v2",
		RateCacheDatabasePath: "/tmp/currency-watcher/rates.db",
		CORSAllowedOrigins:    []string{"http://localhost:3000"},
	}); !reflect.DeepEqual(got, want) {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}
}

func TestLoadTrimsConfiguredValues(t *testing.T) {
	t.Setenv("PORT", " 18080 ")
	t.Setenv("FRANKFURTER_BASE_URL", " https://rates.example.test/v2 ")
	t.Setenv("RATE_CACHE_DB_PATH", " /tmp/rates.db ")
	t.Setenv("CORS_ALLOWED_ORIGINS", " https://one.example , ,https://two.example ")

	if got, want := Load(), (Config{
		Port:                  "18080",
		FrankfurterBaseURL:    "https://rates.example.test/v2",
		RateCacheDatabasePath: "/tmp/rates.db",
		CORSAllowedOrigins:    []string{"https://one.example", "https://two.example"},
	}); !reflect.DeepEqual(got, want) {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}
}
