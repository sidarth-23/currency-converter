package currency

import "time"

// Currency is a provider-supported ISO currency.
type Currency struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// CachedRate is an exchange rate and the time it becomes stale.
type CachedRate struct {
	Rate      float64
	ExpiresAt time.Time
}
