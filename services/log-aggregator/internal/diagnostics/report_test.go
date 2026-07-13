package diagnostics

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/events"
)

func TestDiagnosticTriggerValidity(t *testing.T) {
	supported := []DiagnosticTrigger{
		DiagnosticTriggerManualBugReport,
		DiagnosticTriggerDevelopmentSubmission,
		DiagnosticTriggerCrash,
		DiagnosticTriggerStartupFailure,
		DiagnosticTriggerUnrecoverableState,
		DiagnosticTriggerRecoveryExhausted,
	}
	for _, trigger := range supported {
		if !trigger.Valid() {
			t.Errorf("expected supported trigger %q to be valid", trigger)
		}
	}
	for _, trigger := range []DiagnosticTrigger{"", "unknown", "crash "} {
		if trigger.Valid() {
			t.Errorf("expected trigger %q to be invalid", trigger)
		}
	}
}

func TestDiagnosticSubmissionJSONContract(t *testing.T) {
	submission := DiagnosticSubmission{
		ReportVersion: DiagnosticReportVersion,
		Trigger:       DiagnosticTriggerCrash,
		SubmittedAt:   time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC),
		Source: DiagnosticSourceContext{
			Service: "game-server", ServiceInstanceID: "game-1", Environment: "staging", BuildVersion: "build-7", Platform: "linux",
		},
		Correlation:     DiagnosticCorrelationContext{TraceID: "trace-1", RequestID: "request-1", PlayerID: "player-1", AccountID: "account-1"},
		UserDescription: "crashed while joining",
		Failure:         &DiagnosticFailureContext{FailureMode: "panic", ErrorCode: "E_PANIC"},
		Events:          []json.RawMessage{json.RawMessage(`{"event":"crash","service":"game-server"}`)},
	}
	data, err := json.Marshal(submission)
	if err != nil {
		t.Fatal(err)
	}
	var decoded DiagnosticSubmission
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ReportVersion != DiagnosticReportVersion || decoded.Trigger != DiagnosticTriggerCrash || len(decoded.Events) != 1 {
		t.Fatalf("submission JSON contract did not round-trip: %+v", decoded)
	}
	if decoded.Source.Platform != "linux" {
		t.Errorf("expected submission platform linux, got %q", decoded.Source.Platform)
	}
	if decoded.Correlation.PlayerID != "player-1" {
		t.Errorf("expected submission player ID player-1, got %q", decoded.Correlation.PlayerID)
	}
	if decoded.Correlation.AccountID != "account-1" {
		t.Errorf("expected submission account ID account-1, got %q", decoded.Correlation.AccountID)
	}
}

func TestDiagnosticReportJSONContract(t *testing.T) {
	report := DiagnosticReport{
		DiagnosticReportID: "report-1",
		ReportVersion:      DiagnosticReportVersion,
		Trigger:            DiagnosticTriggerDevelopmentSubmission,
		CreatedAt:          time.Date(2026, 7, 13, 12, 1, 0, 0, time.UTC),
		SubmittedAt:        time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC),
		Source:             DiagnosticSourceContext{Service: "game-server", Platform: "linux"},
		Correlation:        DiagnosticCorrelationContext{PlayerID: "player-1", AccountID: "account-1"},
		Events:             []events.Event{{EventID: "event-1", Service: "game-server"}},
		Summary: DiagnosticReportSummary{
			SubmittedEventCount: 2,
			AcceptedEventCount:  1,
			RejectedEventCount:  1,
			RedactedFieldCount:  3,
			DroppedFieldCount:   1,
			Truncated:           true,
			TruncatedEventCount: 1,
			Services:            []string{"game-server"},
			EventTimeRange:      TimeRange{From: time.Unix(1, 0).UTC(), To: time.Unix(2, 0).UTC()},
		},
	}
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	var shape map[string]json.RawMessage
	if err := json.Unmarshal(data, &shape); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"diagnostic_report_id", "report_version", "trigger", "created_at", "submitted_at", "source", "correlation", "events", "summary"} {
		if _, ok := shape[key]; !ok {
			t.Errorf("report JSON missing %q", key)
		}
	}
	var decoded DiagnosticReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Summary.AcceptedEventCount != 1 || len(decoded.Events) != 1 {
		t.Fatalf("report JSON contract did not round-trip: %+v", decoded)
	}
	if decoded.Source.Platform != "linux" {
		t.Errorf("expected report platform linux, got %q", decoded.Source.Platform)
	}
	if decoded.Correlation.PlayerID != "player-1" {
		t.Errorf("expected report player ID player-1, got %q", decoded.Correlation.PlayerID)
	}
	if decoded.Correlation.AccountID != "account-1" {
		t.Errorf("expected report account ID account-1, got %q", decoded.Correlation.AccountID)
	}
}
