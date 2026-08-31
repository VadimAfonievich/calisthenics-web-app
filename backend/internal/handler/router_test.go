package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthzReportsReadyDependencies(t *testing.T) {
	router := NewRouter(Dependencies{
		CORSOrigins: []string{"http://localhost:5173"},
		Health: HealthDependencies{
			Postgres: func(context.Context) error { return nil },
			Redis:    func(context.Context) error { return nil },
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), `"status":"ok"`) {
		t.Fatalf("unexpected response body: %s", response.Body.String())
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("CORS origin = %q", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, "X-Tenant-Slug") {
		t.Fatalf("CORS headers = %q, want X-Tenant-Slug", got)
	}
	if got := response.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q", got)
	}
	if got := response.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q", got)
	}
}

func TestHealthzReportsUnavailablePostgres(t *testing.T) {
	router := NewRouter(Dependencies{Health: HealthDependencies{
		Postgres: func(context.Context) error { return errors.New("connection refused") },
		Redis:    func(context.Context) error { return nil },
	}})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	if !strings.Contains(response.Body.String(), `"code":"POSTGRES_UNAVAILABLE"`) {
		t.Fatalf("unexpected error response: %s", response.Body.String())
	}
}
