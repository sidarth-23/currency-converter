package httpapi

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)
// RateSource provides cached exchange rates to HTTP handlers.
type RateSource interface {
	Get(ctx context.Context, base string, targets []string) (map[string]float64, error)
}

type HealthOutput struct {
	Body struct {
		Status string `json:"status"`
	}
}

type RatesInput struct {
	Base    string `query:"base" required:"true"`
	Targets string `query:"targets" required:"true"`
}

type RatesOutput struct {
	Body struct {
		Base  string             `json:"base"`
		Rates map[string]float64 `json:"rates"`
	}
}

// NewAPI builds the service routes on the provided standard-library mux.
func NewAPI(mux *http.ServeMux, fetcher RateSource) huma.API {
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
	}, ratesHandler(fetcher))
	return api
}

var currencyCodePattern = regexp.MustCompile(`^[A-Z]{3}$`)

func ratesHandler(fetcher RateSource) func(context.Context, *RatesInput) (*RatesOutput, error) {
	return func(ctx context.Context, input *RatesInput) (*RatesOutput, error) {
		base := input.Base
		if !currencyCodePattern.MatchString(base) {
			return nil, huma.Error422UnprocessableEntity("base must be exactly three uppercase ASCII letters")
		}
		rawTargets := strings.Split(input.Targets, ",")
		if len(rawTargets) == 0 || (len(rawTargets) == 1 && rawTargets[0] == "") {
			return nil, huma.Error422UnprocessableEntity("targets must not be empty")
		}
		seen := make(map[string]struct{}, len(rawTargets))
		targets := make([]string, 0, len(rawTargets))
		for _, target := range rawTargets {
			if !currencyCodePattern.MatchString(target) {
				return nil, huma.Error422UnprocessableEntity("each target must be exactly three uppercase ASCII letters")
			}
			if _, duplicate := seen[target]; duplicate {
				return nil, huma.Error422UnprocessableEntity("targets must not contain duplicates")
			}
			seen[target] = struct{}{}
			targets = append(targets, target)
		}
		if base == "" || len(targets) == 0 {
			return nil, huma.Error422UnprocessableEntity("base and targets are required")
		}
		result, err := fetcher.Get(ctx, base, targets)
		if err != nil {
			return nil, huma.Error500InternalServerError("unable to fetch exchange rates")
		}
		output := &RatesOutput{}
		output.Body.Base = base
		output.Body.Rates = result
		return output, nil
	}
}
