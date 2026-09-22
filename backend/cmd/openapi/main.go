package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/example/currency-watcher/backend/internal/httpapi"
)

//go:generate go run . ../../../contract/currency-watcher/openapi.yaml

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: openapi <output-path>")
		os.Exit(2)
	}
	output, err := os.Create(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "create output: %v\n", err)
		os.Exit(1)
	}
	defer output.Close()

	mux := http.NewServeMux()
	api := httpapi.NewAPI(mux, nil)
	yaml, err := api.OpenAPI().YAML()
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode OpenAPI: %v\n", err)
		os.Exit(1)
	}
	if _, err := output.Write(yaml); err != nil {
		fmt.Fprintf(os.Stderr, "write output: %v\n", err)
		os.Exit(1)
	}
}
