package hosted

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/diagnosticapi"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage/jsonlstore"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	body := `{"report_version":1,"trigger":"manual_bug_report","submitted_at":"2026-07-13T12:01:00Z","source":{"service":"client","service_instance_id":"550e8400-e29b-41d4-a716-446655440001","environment":"test","build_version":"1.0.0","platform":"windows"},"correlation":{"trace_id":"550e8400-e29b-41d4-a716-446655440040"},"events":[{"event_id":"550e8400-e29b-41d4-a716-446655440000","timestamp":"2026-07-13T12:00:00Z","level":"info","event":"ingest_accepted","service":"client","schema_version":1,"service_instance_id":"550e8400-e29b-41d4-a716-446655440001"}]}`
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
	if err := service.Close(); err != nil {
		t.Fatal(err)
	}
	records := readLifecycleRecords(t, cfg)
	var accepted, stored lifecycleRecord
	for _, record := range records {
		switch record.Event {
		case "log_message":
			t.Fatalf("legacy record=%#v", record)
		case "aggregator_event_accepted":
			accepted = record
		case "diagnostic_report_stored":
			stored = record
		}
	}
	if accepted.Event == "" || stored.Event == "" {
		t.Fatalf("events=%v", eventNames(records))
	}
	if accepted.TraceID != "550e8400-e29b-41d4-a716-446655440040" || stored.TraceID != accepted.TraceID || accepted.RequestID == "" || stored.RequestID != accepted.RequestID {
		t.Fatalf("accepted=%#v stored=%#v", accepted, stored)
	}
	if accepted.Route != diagnosticapi.DiagnosticReportsPath || stored.Route != accepted.Route || accepted.DiagnosticReportID != report.ID || stored.DiagnosticReportID != report.ID {
		t.Fatalf("accepted=%#v stored=%#v", accepted, stored)
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

type lifecycleRecord struct {
	Event              string         `json:"event"`
	TraceID            string         `json:"trace_id"`
	Fields             map[string]any `json:"fields"`
	RequestID          string         `json:"request_id"`
	Route              string         `json:"route"`
	DiagnosticReportID string         `json:"diagnostic_report_id"`
}

func lifecycleConfig(t *testing.T) Config {
	t.Helper()
	config := DefaultConfig()
	config.Enabled = true
	config.BearerToken = "test-token"
	config.StorageRoot = t.TempDir()
	config.OperationalLogRoot = t.TempDir()
	return config
}

func readLifecycleRecords(t *testing.T, config Config) []lifecycleRecord {
	t.Helper()
	data, err := os.ReadFile(operationalLogConfig(config).ActivePath())
	if err != nil {
		t.Fatal(err)
	}
	var records []lifecycleRecord
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var record lifecycleRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("record=%q err=%v", line, err)
		}
		records = append(records, record)
	}
	return records
}

func eventNames(records []lifecycleRecord) []string {
	names := make([]string, len(records))
	for index, record := range records {
		names[index] = record.Event
	}
	return names
}

func TestHostedLifecycleUsesCanonicalOrderAndSeparateTraces(t *testing.T) {
	config := lifecycleConfig(t)
	traces := []string{"550e8400-e29b-41d4-a716-446655440030", "550e8400-e29b-41d4-a716-446655440031"}
	index := 0
	deps := defaultServiceDependencies
	deps.newUUID = func() string { value := traces[index]; index++; return value }
	service, err := newWithDependencies(config, deps)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Close(); err != nil {
		t.Fatal(err)
	}
	if err := service.Close(); err != nil {
		t.Fatal(err)
	}
	records := readLifecycleRecords(t, config)
	want := []string{"service_starting", "service_started", "service_stopping", "service_stopped"}
	if got := eventNames(records); !equalEventNames(got, want) {
		t.Fatalf("events=%v", got)
	}
	if records[0].TraceID != traces[0] || records[1].TraceID != traces[0] || records[2].TraceID != traces[1] || records[3].TraceID != traces[1] || traces[0] == traces[1] {
		t.Fatalf("records=%#v", records)
	}
}

func TestHostedInitializationFailuresDoNotEmitStarted(t *testing.T) {
	t.Run("store", func(t *testing.T) {
		config := lifecycleConfig(t)
		deps := defaultServiceDependencies
		deps.newUUID = func() string { return "550e8400-e29b-41d4-a716-446655440032" }
		deps.newStore = func(jsonlstore.Config) (reportStore, error) { return nil, errors.New("store unavailable") }
		if _, err := newWithDependencies(config, deps); err == nil {
			t.Fatal("expected store failure")
		}
		if got := eventNames(readLifecycleRecords(t, config)); !equalEventNames(got, []string{"service_starting", "dependency_initialization_failed"}) {
			t.Fatalf("events=%v", got)
		}
	})

	t.Run("handler", func(t *testing.T) {
		config := lifecycleConfig(t)
		deps := defaultServiceDependencies
		deps.newUUID = func() string { return "550e8400-e29b-41d4-a716-446655440033" }
		deps.newHandler = func(diagnosticapi.ReportService, diagnosticapi.HandlerConfig) (http.Handler, error) {
			return nil, errors.New("handler unavailable")
		}
		if _, err := newWithDependencies(config, deps); err == nil {
			t.Fatal("expected handler failure")
		}
		records := readLifecycleRecords(t, config)
		if got := eventNames(records); !equalEventNames(got, []string{"service_starting", "service_startup_failed"}) {
			t.Fatalf("events=%v", got)
		}
		if records[1].Fields["failure_stage"] != "handler_initialization" {
			t.Fatalf("fields=%#v", records[1].Fields)
		}
	})
}

type failingCloseStore struct{ closeErr error }

func (*failingCloseStore) Save(context.Context, storage.Report) error { return nil }
func (*failingCloseStore) Get(context.Context, string) (storage.Report, error) {
	return storage.Report{}, storage.ErrReportNotFound
}
func (*failingCloseStore) DeleteExpired(context.Context, time.Time) (int, error) { return 0, nil }
func (store *failingCloseStore) Close() error                                    { return store.closeErr }
func (*failingCloseStore) EnforceRetention(context.Context) (int, error)         { return 0, nil }

func TestHostedStoreCloseFailureStillEmitsStopped(t *testing.T) {
	config := lifecycleConfig(t)
	deps := defaultServiceDependencies
	ids := []string{"550e8400-e29b-41d4-a716-446655440034", "550e8400-e29b-41d4-a716-446655440035"}
	index := 0
	deps.newUUID = func() string { value := ids[index]; index++; return value }
	deps.newStore = func(jsonlstore.Config) (reportStore, error) {
		return &failingCloseStore{closeErr: errors.New("close unavailable")}, nil
	}
	service, err := newWithDependencies(config, deps)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Close(); err == nil {
		t.Fatal("expected close failure")
	}
	if got := eventNames(readLifecycleRecords(t, config)); !equalEventNames(got, []string{"service_starting", "service_started", "service_stopping", "aggregator_storage_failed", "service_stopped"}) {
		t.Fatalf("events=%v", got)
	}
}

func equalEventNames(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}
