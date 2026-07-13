package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/diagnosticapi"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/diagnostics"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/events"
)

type e2eReportService struct {
	report      diagnostics.DiagnosticReport
	createCalls int
	getCalls    int
}

func (s *e2eReportService) Create(_ context.Context, submission diagnosticapi.DiagnosticReportCreateRequest) (diagnosticapi.DiagnosticReport, error) {
	s.createCalls++
	decoded := make([]events.Event, 0, len(submission.Events))
	for _, raw := range submission.Events {
		event, err := events.Decode(raw)
		if err != nil {
			return diagnostics.DiagnosticReport{}, err
		}
		decoded = append(decoded, event)
	}
	s.report = diagnostics.DiagnosticReport{
		DiagnosticReportID: "123e4567-e89b-12d3-a456-426614174099",
		ReportVersion:      submission.ReportVersion,
		Trigger:            submission.Trigger,
		CreatedAt:          submission.SubmittedAt,
		SubmittedAt:        submission.SubmittedAt,
		Source:             submission.Source,
		Correlation:        submission.Correlation,
		UserDescription:    submission.UserDescription,
		Events:             decoded,
		Summary: diagnostics.DiagnosticReportSummary{
			SubmittedEventCount: uint64(len(decoded)), AcceptedEventCount: uint64(len(decoded)),
		},
	}
	return s.report, nil
}

func (s *e2eReportService) Get(_ context.Context, reportID string) (diagnosticapi.DiagnosticReport, error) {
	s.getCalls++
	if reportID != s.report.DiagnosticReportID {
		return diagnostics.DiagnosticReport{}, diagnosticapi.ErrReportNotFound
	}
	return s.report, nil
}

func TestDiagnosticSubmitEndToEnd(t *testing.T) {
	service := &e2eReportService{}
	handler, err := diagnosticapi.NewHandler(service, diagnosticapi.HandlerConfig{
		MaxRequestBytes: 1 << 20,
		Authorize:       func(r *http.Request) bool { return r.Header.Get("Authorization") == "Bearer expected-token" },
	})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()

	ids := []string{
		"123e4567-e89b-12d3-a456-426614174001", "123e4567-e89b-12d3-a456-426614174002",
		"123e4567-e89b-12d3-a456-426614174003", "123e4567-e89b-12d3-a456-426614174004",
		"123e4567-e89b-12d3-a456-426614174005",
	}
	index := 0
	deps := dependencies{
		httpClient: server.Client(),
		now:        func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) },
		newUUID:    func() string { value := ids[index]; index++; return value },
		environ: func(name string) string {
			if name == "DIAGNOSTIC_AGGREGATOR_URL" {
				return server.URL
			}
			if name == "DIAGNOSTIC_AGGREGATOR_TOKEN" {
				return "expected-token"
			}
			return ""
		},
	}
	var stdout, stderr strings.Builder
	if code := run([]string{"-description", "end-to-end check", "-timeout", "2s"}, &stdout, &stderr, deps); code != exitOK {
		t.Fatalf("run code=%d stderr=%q", code, stderr.String())
	}
	if service.createCalls != 1 || service.getCalls != 1 {
		t.Fatalf("service calls create=%d get=%d", service.createCalls, service.getCalls)
	}
	var output diagnostics.DiagnosticReport
	if err := json.Unmarshal([]byte(stdout.String()), &output); err != nil {
		t.Fatalf("stdout JSON: %v", err)
	}
	if output.DiagnosticReportID != service.report.DiagnosticReportID || output.Trigger != diagnostics.DiagnosticTriggerDevelopmentSubmission || output.UserDescription != "end-to-end check" {
		t.Fatalf("unexpected report: %#v", output)
	}
	if output.Correlation.TraceID != "123e4567-e89b-12d3-a456-426614174002" || output.Correlation.SessionID != "123e4567-e89b-12d3-a456-426614174003" || len(output.Events) != 2 || output.Summary.AcceptedEventCount != 2 {
		t.Fatalf("unexpected finalized report: %#v", output)
	}
	if len(stdout.String()) == 0 || !strings.Contains(stdout.String(), "\n  \"diagnostic_report_id\"") || stderr.Len() != 0 {
		t.Fatalf("unexpected output stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}
