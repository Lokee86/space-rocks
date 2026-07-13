package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/diagnostics"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/events"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestBuildSubmissionIsDeterministicAndCanonical(t *testing.T) {
	now := func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.FixedZone("x", 3600)) }
	i := 0
	ids := []string{
		"123e4567-e89b-12d3-a456-426614174001",
		"123e4567-e89b-12d3-a456-426614174002",
		"123e4567-e89b-12d3-a456-426614174003",
		"123e4567-e89b-12d3-a456-426614174004",
		"123e4567-e89b-12d3-a456-426614174005",
	}
	submission := buildSubmission(now, func() string { value := ids[i]; i++; return value }, "description")
	if submission.Trigger != diagnostics.DiagnosticTriggerDevelopmentSubmission || submission.ReportVersion != 1 || submission.SubmittedAt.Location() != time.UTC {
		t.Fatalf("unexpected submission metadata: %#v", submission)
	}
	if submission.Source.Service != "diagnostic-submit" || submission.Source.ServiceInstanceID != ids[0] || submission.Source.Environment != "development" || submission.Source.BuildVersion != "manual" || submission.Source.Platform != runtime.GOOS+"/"+runtime.GOARCH || submission.Correlation.TraceID != ids[1] || submission.Correlation.SessionID != ids[2] {
		t.Fatalf("unexpected source/correlation: %#v %#v", submission.Source, submission.Correlation)
	}
	wantNames := []string{"diagnostic_submission_started", "development_failure_observed"}
	wantLevels := []string{"info", "error"}
	for index, raw := range submission.Events {
		event, err := events.Decode(raw)
		if err != nil {
			t.Fatalf("event %d: %v", index, err)
		}
		if event.Event != wantNames[index] || event.Level != wantLevels[index] || event.EventID != ids[index+3] || event.Service != "diagnostic-submit" || event.ServiceInstanceID != ids[0] || event.Environment != "development" || event.BuildVersion != "manual" || event.TraceID != ids[1] || event.SessionID != ids[2] || event.Fields["source"] != "manual_cli" {
			t.Fatalf("unexpected event %d: %#v", index, event)
		}
	}
}

func TestRunUsesEnvironmentDefaultsCreateThenGetAndIndentedOutput(t *testing.T) {
	requests := 0
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requests++
		if r.Header.Get("Authorization") != "Bearer token" || r.URL.Path == "" {
			t.Fatalf("unexpected request: %#v", r)
		}
		if requests == 1 && r.Method != http.MethodPost {
			t.Fatalf("create method=%s", r.Method)
		}
		if requests == 2 && r.Method != http.MethodGet {
			t.Fatalf("get method=%s", r.Method)
		}
		status := http.StatusCreated
		if requests == 2 {
			status = http.StatusOK
		}
		body := `{"diagnostic_report_id":"report-1","report_version":1}`
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	var out, errOut bytes.Buffer
	deps := dependencies{httpClient: &http.Client{Transport: transport}, now: time.Now, newUUID: func() string { return "123e4567-e89b-12d3-a456-426614174000" }, environ: func(key string) string {
		if key == "DIAGNOSTIC_AGGREGATOR_URL" {
			return "http://example.test"
		}
		return "token"
	}}
	if code := run([]string{"-description", "safe", "-timeout", "1s"}, &out, &errOut, deps); code != exitOK || requests != 2 || !strings.Contains(out.String(), "\n  \"diagnostic_report_id\"") || errOut.Len() != 0 {
		t.Fatalf("run code=%d requests=%d out=%q err=%q", code, requests, out.String(), errOut.String())
	}
}

func TestRunRequiredConfigAndSanitizedFailure(t *testing.T) {
	deps := defaultDependencies()
	deps.environ = func(string) string { return "" }
	var out, errOut bytes.Buffer
	if code := run(nil, &out, &errOut, deps); code != exitConfiguration || !strings.Contains(errOut.String(), "configuration") {
		t.Fatalf("code=%d err=%q", code, errOut.String())
	}

	deps.environ = func(string) string { return "token" }
	deps.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) { return nil, context.Canceled })}
	errOut.Reset()
	if code := run([]string{"-base-url", "http://example.test", "-timeout", "1s"}, &out, &errOut, deps); code != exitSubmission || strings.Contains(errOut.String(), "token") {
		t.Fatalf("code=%d err=%q", code, errOut.String())
	}
}
