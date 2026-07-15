package hosted

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegisteredMuxPostsAndGetsFinalizedReport(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.BearerToken = "test-token"
	cfg.StorageRoot = t.TempDir()
	cfg.OperationalLogRoot = t.TempDir()
	service, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()
	mux := http.NewServeMux()
	if err := service.Register(mux); err != nil {
		t.Fatal(err)
	}
	body := `{"report_version":1,"trigger":"manual_bug_report","submitted_at":"2026-07-13T12:01:00Z","source":{"service":"client","service_instance_id":"550e8400-e29b-41d4-a716-446655440001","environment":"test","build_version":"1.0.0","platform":"windows"},"events":[{"event_id":"550e8400-e29b-41d4-a716-446655440000","timestamp":"2026-07-13T12:00:00Z","level":"info","event":"ingest_accepted","service":"client","schema_version":1,"service_instance_id":"550e8400-e29b-41d4-a716-446655440001"}]}`
	post := httptest.NewRequest(http.MethodPost, "/v1/diagnostic-reports", strings.NewReader(body))
	post.Header.Set("Content-Type", "application/json")
	post.Header.Set("Authorization", "Bearer test-token")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, post)
	if response.Code != http.StatusCreated {
		t.Fatalf("POST status=%d body=%s", response.Code, response.Body)
	}
	var report struct {
		ID string `json:"diagnostic_report_id"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &report); err != nil || report.ID == "" {
		t.Fatalf("POST report=%s", response.Body)
	}
	get := httptest.NewRequest(http.MethodGet, "/v1/diagnostic-reports/"+report.ID, nil)
	get.Header.Set("Authorization", "Bearer test-token")
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, get)
	if response.Code != http.StatusOK {
		t.Fatalf("GET status=%d body=%s", response.Code, response.Body)
	}
	unauthorized := httptest.NewRecorder()
	mux.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/v1/diagnostic-reports/"+report.ID, nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status=%d", unauthorized.Code)
	}
}
func TestOperationalLoggingLifecycleAndFailureDoNotDisableHTTP(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.BearerToken = "test-token"
	cfg.StorageRoot = t.TempDir()
	blocked := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(blocked, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg.OperationalLogRoot = filepath.Join(blocked, "logs")
	service, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !service.LoggingStatus().Degraded {
		t.Fatalf("status=%#v", service.LoggingStatus())
	}
	mux := http.NewServeMux()
	if err := service.Register(mux); err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/diagnostic-reports/missing", nil))
	if response.Code == http.StatusInternalServerError {
		t.Fatalf("status=%d", response.Code)
	}
	if err := service.Close(); err != nil {
		t.Fatal(err)
	}
	if err := service.Close(); err != nil {
		t.Fatal(err)
	}
	if !service.LoggingStatus().Closed {
		t.Fatalf("status=%#v", service.LoggingStatus())
	}
}
