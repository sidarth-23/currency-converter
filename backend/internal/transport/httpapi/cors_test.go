package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithCORSHandlesAllowedPreflight(t *testing.T) {
	handler := WithCORS(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}), []string{"http://localhost:3000"})
	request := httptest.NewRequest(http.MethodOptions, "/api/rates", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent || response.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("unexpected preflight response: %d %#v", response.Code, response.Header())
	}
}

func TestWithCORSLeavesDisallowedRequestsUnchanged(t *testing.T) {
	handler := WithCORS(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}), []string{"http://localhost:3000"})
	request := httptest.NewRequest(http.MethodOptions, "/api/rates", nil)
	request.Header.Set("Origin", "https://other.example")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusTeapot || response.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("unexpected disallowed response: %d %#v", response.Code, response.Header())
	}
}
