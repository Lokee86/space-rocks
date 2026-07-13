package diagnosticclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/diagnostics"
)

func testReport() diagnostics.DiagnosticReport {
	return diagnostics.DiagnosticReport{DiagnosticReportID: "report-1", ReportVersion: 1, Trigger: diagnostics.DiagnosticTriggerManualBugReport}
}

func newTestClient(t *testing.T, handler http.Handler, max int64) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := New(Config{BaseURL: server.URL, BearerToken: "secret", HTTPClient: server.Client(), MaxResponseBytes: max})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func reportJSON(t *testing.T) string {
	t.Helper()
	return `{"diagnostic_report_id":"report-1","report_version":1,"trigger":"manual_bug_report"}`
}

func TestCreateAndGetRoutesHeadersAndDecode(t *testing.T) {
	report := reportJSON(t)
	calls := 0
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method == http.MethodPost {
			if r.URL.Path != "/v1/diagnostic-reports" || r.Header.Get("Content-Type") != "application/json" || r.Header.Get("Authorization") != "Bearer secret" {
				t.Errorf("unexpected create request: %s %s", r.Method, r.URL.String())
			}
			if _, err := io.ReadAll(r.Body); err != nil {
				t.Error(err)
			}
			w.WriteHeader(http.StatusCreated)
		} else {
			if r.Method != http.MethodGet || r.URL.Path != "/v1/diagnostic-reports/report-1" || r.Header.Get("Authorization") != "Bearer secret" {
				t.Errorf("unexpected get request: %s %s", r.Method, r.URL.String())
			}
			w.WriteHeader(http.StatusOK)
		}
		_, _ = w.Write([]byte(report))
	}), 1024)

	got, err := client.Create(context.Background(), diagnostics.DiagnosticSubmission{})
	if err != nil || got.DiagnosticReportID != "report-1" {
		t.Fatalf("Create() = %#v, %v", got, err)
	}
	got, err = client.Get(context.Background(), "report-1")
	if err != nil || got.DiagnosticReportID != "report-1" || calls != 2 {
		t.Fatalf("Get() = %#v, %v; calls=%d", got, err, calls)
	}
}

func TestErrorsAreStableAndDoNotExposeBody(t *testing.T) {
	tests := []struct {
		name string
		body string
		want error
	}{
		{"malformed", "not json", ErrMalformedJSON},
		{"trailing", reportJSON(t) + reportJSON(t), ErrTrailingJSON},
		{"oversized", strings.Repeat("x", 20), ErrResponseTooLarge},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			max := int64(1024)
			if tc.name == "oversized" {
				max = 10
			}
			client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tc.body))
			}), max)
			_, err := client.Get(context.Background(), "id")
			if !errors.Is(err, tc.want) || strings.Contains(err.Error(), tc.body) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestUnexpectedStatusAndCancellation(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("secret response"))
	}), 1024)
	_, err := client.Get(context.Background(), "id")
	if !errors.Is(err, ErrUnexpectedStatus) || strings.Contains(err.Error(), "secret response") {
		t.Fatalf("status error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.Get(ctx, "id")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v", err)
	}
}

func TestNewGuards(t *testing.T) {
	valid := Config{BaseURL: "http://example.test", HTTPClient: http.DefaultClient, MaxResponseBytes: 1}
	for _, mutate := range []func(*Config){
		func(c *Config) { c.BaseURL = "://bad" },
		func(c *Config) { c.BaseURL = "ftp://example.test" },
		func(c *Config) { c.HTTPClient = nil },
		func(c *Config) { c.MaxResponseBytes = 0 },
	} {
		config := valid
		mutate(&config)
		if _, err := New(config); !errors.Is(err, ErrInvalidConfig) {
			t.Fatalf("New(%+v) error = %v", config, err)
		}
	}
	if _, err := New(valid); err != nil {
		t.Fatalf("valid New() error = %v", err)
	}
}
