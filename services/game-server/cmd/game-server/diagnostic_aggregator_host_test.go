package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	serverlogging "github.com/Lokee86/space-rocks/services/game-server/internal/logging"
)

func TestRegisterDiagnosticAggregatorDisabled(t *testing.T) {
	t.Setenv("DIAGNOSTIC_AGGREGATOR_ENABLED", "false")
	t.Setenv("DIAGNOSTIC_AGGREGATOR_TOKEN", "")
	mux := http.NewServeMux()
	service, err := registerDiagnosticAggregator(mux)
	if err != nil {
		t.Fatal(err)
	}
	defer closeDiagnosticAggregator(service)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/diagnostic-reports", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("status=%d", response.Code)
	}
}

func TestRegisterDiagnosticAggregatorRejectsInvalidEnabledConfig(t *testing.T) {
	t.Setenv("DIAGNOSTIC_AGGREGATOR_ENABLED", "true")
	t.Setenv("DIAGNOSTIC_AGGREGATOR_TOKEN", "bad token")
	if _, err := registerDiagnosticAggregator(http.NewServeMux()); err == nil {
		t.Fatal("expected invalid configuration error")
	}
}

func TestRegisterDiagnosticAggregatorAuthenticatedPostGet(t *testing.T) {
	t.Setenv("DIAGNOSTIC_AGGREGATOR_ENABLED", "true")
	t.Setenv("DIAGNOSTIC_AGGREGATOR_TOKEN", "test-token")
	t.Setenv("DIAGNOSTIC_AGGREGATOR_STORAGE_ROOT", t.TempDir())
	t.Setenv("DIAGNOSTIC_AGGREGATOR_LOG_ROOT", t.TempDir())
	mux := http.NewServeMux()
	service, err := registerDiagnosticAggregator(mux)
	if err != nil {
		t.Fatal(err)
	}
	defer closeDiagnosticAggregator(service)
	body := `{"report_version":1,"trigger":"manual_bug_report","submitted_at":"2026-07-13T12:01:00Z","source":{"service":"client","service_instance_id":"550e8400-e29b-41d4-a716-446655440001","environment":"test","build_version":"1.0.0","platform":"windows"},"events":[{"event_id":"550e8400-e29b-41d4-a716-446655440000","timestamp":"2026-07-13T12:00:00Z","level":"info","event":"ingest_accepted","service":"client","schema_version":1,"service_instance_id":"550e8400-e29b-41d4-a716-446655440001"}]}`
	request := httptest.NewRequest(http.MethodPost, "/v1/diagnostic-reports", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer test-token")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("POST status=%d body=%s", response.Code, response.Body)
	}
	get := httptest.NewRequest(http.MethodGet, response.Header().Get("Location"), nil)
	get.Header.Set("Authorization", "Bearer test-token")
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, get)
	if response.Code != http.StatusOK {
		t.Fatalf("GET status=%d body=%s", response.Code, response.Body)
	}
}
func TestGameServerAndHostedAggregatorUseSeparateOperationalFiles(t *testing.T) {
	t.Setenv(envBuildVersion, "test-build")
	t.Setenv(envEnvironment, "test")
	identity, err := loadLoggingIdentity(serverlogging.ServiceName)
	if err != nil {
		t.Fatal(err)
	}
	if err := serverlogging.ConfigureRuntime(identity); err != nil {
		t.Fatal(err)
	}
	gameDir := filepath.Join(t.TempDir(), "game-server")
	gamePath, err := serverlogging.ConfigureFileOutput(gameDir, "game-server")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = serverlogging.CloseFileOutput() })

	aggregatorDir := filepath.Join(t.TempDir(), "diagnostic-aggregator")
	t.Setenv("DIAGNOSTIC_AGGREGATOR_ENABLED", "true")
	t.Setenv("DIAGNOSTIC_AGGREGATOR_TOKEN", "test-token")
	t.Setenv("DIAGNOSTIC_AGGREGATOR_STORAGE_ROOT", t.TempDir())
	t.Setenv("DIAGNOSTIC_AGGREGATOR_LOG_ROOT", aggregatorDir)
	service, err := registerDiagnosticAggregator(http.NewServeMux())
	if err != nil {
		t.Fatal(err)
	}
	defer closeDiagnosticAggregator(service)

	aggregatorPath := filepath.Join(aggregatorDir, "diagnostic-aggregator.jsonl.open")
	if gamePath == aggregatorPath {
		t.Fatalf("shared active path %q", gamePath)
	}
	for _, path := range []string{gamePath, aggregatorPath} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("active path %q: %v", path, err)
		}
	}
}
