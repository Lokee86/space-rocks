package diagnostics

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/events"
)

func validSubmissionLimits() SubmissionLimits {
	return SubmissionLimits{
		MaxEmbeddedEvents:       1,
		MaxUserDescriptionBytes: 64,
		MaxFailureMessageBytes:  64,
		MaxContextStringBytes:   64,
	}
}

func TestSubmissionLimitsValidateAcceptsPositiveLimits(t *testing.T) {
	if err := validSubmissionLimits().Validate(); err != nil {
		t.Fatalf("valid limits rejected: %v", err)
	}
}

func TestSubmissionLimitsValidateRejectsEachInvalidLimit(t *testing.T) {
	tests := []struct {
		name string
		set  func(*SubmissionLimits)
		code ValidationCode
	}{
		{"embedded events", func(l *SubmissionLimits) { l.MaxEmbeddedEvents = 0 }, ValidationCodeEmbeddedEventsLimit},
		{"user description", func(l *SubmissionLimits) { l.MaxUserDescriptionBytes = -1 }, ValidationCodeUserDescriptionLimit},
		{"failure message", func(l *SubmissionLimits) { l.MaxFailureMessageBytes = 0 }, ValidationCodeFailureMessageLimit},
		{"context string", func(l *SubmissionLimits) { l.MaxContextStringBytes = -1 }, ValidationCodeContextStringLimit},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			limits := validSubmissionLimits()
			test.set(&limits)
			if _, ok := ValidationCodeOf(limits.Validate()); !ok {
				t.Fatal("expected validation error")
			}
			code, _ := ValidationCodeOf(limits.Validate())
			if code != test.code {
				t.Fatalf("validation code = %q, want %q", code, test.code)
			}
		})
	}
}

func TestValidationErrorCodeExtractionDoesNotExposeValues(t *testing.T) {
	err := &ValidationError{code: ValidationCodeUserDescriptionLimit}
	if got, ok := ValidationCodeOf(err); !ok || got != ValidationCodeUserDescriptionLimit {
		t.Fatalf("unexpected extracted code: %q, %v", got, ok)
	}
	if errors.Is(err, errors.New("submitted secret")) {
		t.Fatal("validation error unexpectedly matched submitted value")
	}
	if got := err.Error(); got != string(ValidationCodeUserDescriptionLimit) {
		t.Fatalf("error exposed unexpected data: %q", got)
	}
	if _, ok := ValidationCodeOf(errors.New("submitted secret")); ok {
		t.Fatal("non-validation error produced a validation code")
	}
}

func validDiagnosticSubmission() DiagnosticSubmission {
	return DiagnosticSubmission{
		ReportVersion: DiagnosticReportVersion,
		Trigger:       DiagnosticTriggerDevelopmentSubmission,
		SubmittedAt:   time.Unix(1, 0),
		Source: DiagnosticSourceContext{
			Service: "client", ServiceInstanceID: "123e4567-e89b-12d3-a456-426614174000",
			Environment: "test", BuildVersion: "build", Platform: "windows",
		},
		Correlation: DiagnosticCorrelationContext{TraceID: "123e4567-e89b-12d3-a456-426614174001"},
		Events:      []json.RawMessage{json.RawMessage("{\"event\":\"ok\"}")},
	}
}

func TestValidateSubmissionEnvelopeAcceptsValidBoundary(t *testing.T) {
	limits := SubmissionLimits{MaxEmbeddedEvents: 1, MaxUserDescriptionBytes: 4, MaxFailureMessageBytes: 3, MaxContextStringBytes: 36}
	submission := validDiagnosticSubmission()
	submission.UserDescription = "four"
	if err := ValidateSubmissionEnvelope(submission, limits); err != nil {
		t.Fatalf("valid boundary rejected: %v", err)
	}
}

func TestValidateSubmissionEnvelopeRejectionClasses(t *testing.T) {
	base := validDiagnosticSubmission()
	limits := SubmissionLimits{MaxEmbeddedEvents: 1, MaxUserDescriptionBytes: 10, MaxFailureMessageBytes: 10, MaxContextStringBytes: 40}
	tests := []struct {
		name   string
		mutate func(*DiagnosticSubmission, *SubmissionLimits)
		code   ValidationCode
	}{
		{"report version", func(s *DiagnosticSubmission, _ *SubmissionLimits) { s.ReportVersion = 99 }, ValidationCodeInvalidReportVersion},
		{"trigger", func(s *DiagnosticSubmission, _ *SubmissionLimits) { s.Trigger = "unknown" }, ValidationCodeUnsupportedTrigger},
		{"submitted at", func(s *DiagnosticSubmission, _ *SubmissionLimits) { s.SubmittedAt = time.Time{} }, ValidationCodeMissingSubmittedAt},
		{"source", func(s *DiagnosticSubmission, _ *SubmissionLimits) { s.Source.Service = "" }, ValidationCodeMissingSourceField},
		{"source length", func(s *DiagnosticSubmission, l *SubmissionLimits) {
			l.MaxContextStringBytes = 3
			s.Source.Service = "long"
		}, ValidationCodeContextStringExceeded},
		{"uuid", func(s *DiagnosticSubmission, _ *SubmissionLimits) { s.Source.ServiceInstanceID = "not-uuid" }, ValidationCodeInvalidCorrelationID},
		{"events", func(s *DiagnosticSubmission, _ *SubmissionLimits) { s.Events = append(s.Events, json.RawMessage(`{}`)) }, ValidationCodeEmbeddedEventsExceeded},
		{"description", func(s *DiagnosticSubmission, l *SubmissionLimits) {
			l.MaxUserDescriptionBytes = 1
			s.UserDescription = "too long"
		}, ValidationCodeUserDescriptionExceeded},
		{"failure mode", func(s *DiagnosticSubmission, _ *SubmissionLimits) { s.Failure = &DiagnosticFailureContext{} }, ValidationCodeMissingFailureMode},
		{"failure message", func(s *DiagnosticSubmission, l *SubmissionLimits) {
			l.MaxFailureMessageBytes = 1
			s.Failure = &DiagnosticFailureContext{FailureMode: "crash", Message: "long"}
		}, ValidationCodeFailureMessageExceeded},
		{"context", func(s *DiagnosticSubmission, l *SubmissionLimits) {
			l.MaxContextStringBytes = 1
			s.Correlation.TraceID = "123e4567-e89b-12d3-a456-426614174001"
		}, ValidationCodeContextStringExceeded},
		{"empty event", func(s *DiagnosticSubmission, _ *SubmissionLimits) { s.Events = []json.RawMessage{nil} }, ValidationCodeEmptyEmbeddedEvent},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			submission, testLimits := base, limits
			test.mutate(&submission, &testLimits)
			if code, ok := ValidationCodeOf(ValidateSubmissionEnvelope(submission, testLimits)); !ok || code != test.code {
				t.Fatalf("validation code = %q, want %q", code, test.code)
			}
		})
	}
}

const validDiagnosticEvent = "{\"event_id\":\"550e8400-e29b-41d4-a716-446655440000\",\"timestamp\":\"2026-07-13T12:00:00Z\",\"level\":\"info\",\"event\":\"ingest_accepted\",\"service\":\"log-aggregator\",\"schema_version\":1,\"service_instance_id\":\"550e8400-e29b-41d4-a716-446655440001\",\"fields\":{\"count\":1}}"

func TestDecodeSubmissionEventsDecodesValidEvents(t *testing.T) {
	submission := validDiagnosticSubmission()
	submission.Events = []json.RawMessage{json.RawMessage(validDiagnosticEvent)}
	decoded, err := DecodeSubmissionEvents(submission, validSubmissionLimits())
	if err != nil || len(decoded) != 1 || decoded[0].Event != "ingest_accepted" {
		t.Fatalf("decoded events = %#v, err = %v", decoded, err)
	}
}

func TestDecodeSubmissionEventsRejectsInvalidCanonicalEventsWithStableCode(t *testing.T) {
	var eventDocument map[string]any
	if err := json.Unmarshal([]byte(validDiagnosticEvent), &eventDocument); err != nil {
		t.Fatal(err)
	}
	eventDocument["unknown"] = "value"
	unknownField, err := json.Marshal(eventDocument)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(unknownField) {
		t.Fatal("unknown-field fixture is not valid JSON")
	}

	incomplete, err := json.Marshal(map[string]string{
		"event_id": "550e8400-e29b-41d4-a716-446655440000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(incomplete) {
		t.Fatal("incomplete-envelope fixture is not valid JSON")
	}

	malformed := append([]byte(nil), incomplete[:len(incomplete)-1]...)
	if json.Valid(malformed) {
		t.Fatal("malformed fixture is valid JSON")
	}

	tests := [][]byte{unknownField, malformed, incomplete}
	for _, raw := range tests {
		submission := validDiagnosticSubmission()
		submission.Events = []json.RawMessage{json.RawMessage(raw)}
		_, err := DecodeSubmissionEvents(submission, validSubmissionLimits())
		code, ok := ValidationCodeOf(err)
		if !ok || code != ValidationCodeInvalidEmbeddedEvent {
			t.Fatalf("raw event produced code %q, want %q", code, ValidationCodeInvalidEmbeddedEvent)
		}
	}
}

func TestDecodeSubmissionEventsReturnsIsolatedSlice(t *testing.T) {
	submission := validDiagnosticSubmission()
	submission.Events = []json.RawMessage{json.RawMessage(validDiagnosticEvent)}
	decoded, err := DecodeSubmissionEvents(submission, validSubmissionLimits())
	if err != nil {
		t.Fatal(err)
	}
	decoded[0].Event = "changed"
	again, err := DecodeSubmissionEvents(submission, validSubmissionLimits())
	if err != nil || again[0].Event != "ingest_accepted" {
		t.Fatalf("decoded result shared mutable state: %#v, err = %v", again, err)
	}
}

func validDiagnosticReport() DiagnosticReport {
	eventTime := "2026-07-13T12:00:00Z"
	return DiagnosticReport{
		DiagnosticReportID: "123e4567-e89b-12d3-a456-426614174002",
		ReportVersion:      DiagnosticReportVersion,
		Trigger:            DiagnosticTriggerManualBugReport,
		CreatedAt:          time.Unix(2, 0),
		SubmittedAt:        time.Unix(1, 0),
		Source: DiagnosticSourceContext{
			Service: "client", ServiceInstanceID: "123e4567-e89b-12d3-a456-426614174000",
			Environment: "test", BuildVersion: "build", Platform: "windows",
		},
		Events: []events.Event{{
			EventID: "550e8400-e29b-41d4-a716-446655440000", Timestamp: eventTime,
			Level: "info", Event: "ingest_accepted", Service: "log-aggregator",
			SchemaVersion: 1, ServiceInstanceID: "550e8400-e29b-41d4-a716-446655440001",
		}},
		Summary: DiagnosticReportSummary{
			SubmittedEventCount: 1,
			AcceptedEventCount:  1,
			EventTimeRange:      TimeRange{From: time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC), To: time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)},
		},
	}
}

func TestValidateDiagnosticReportAcceptsValidReport(t *testing.T) {
	report := validDiagnosticReport()
	limits := SubmissionLimits{MaxEmbeddedEvents: 1, MaxUserDescriptionBytes: 10, MaxFailureMessageBytes: 10, MaxContextStringBytes: 40}
	if err := ValidateDiagnosticReport(report, limits); err != nil {
		t.Fatalf("valid report rejected: %v", err)
	}
}

func TestValidateDiagnosticReportRejectsInvariantClasses(t *testing.T) {
	limits := SubmissionLimits{MaxEmbeddedEvents: 1, MaxUserDescriptionBytes: 10, MaxFailureMessageBytes: 10, MaxContextStringBytes: 40}
	tests := []struct {
		name   string
		mutate func(*DiagnosticReport)
		code   ValidationCode
	}{
		{"report id", func(r *DiagnosticReport) { r.DiagnosticReportID = "bad" }, ValidationCodeInvalidReportID},
		{"version", func(r *DiagnosticReport) { r.ReportVersion = 2 }, ValidationCodeReportInvariant},
		{"source uuid", func(r *DiagnosticReport) { r.Source.ServiceInstanceID = "bad" }, ValidationCodeReportInvariant},
		{"accepted count", func(r *DiagnosticReport) { r.Summary.AcceptedEventCount = 2 }, ValidationCodeReportInvariant},
		{"submitted count", func(r *DiagnosticReport) { r.Summary.SubmittedEventCount = 2 }, ValidationCodeReportInvariant},
		{"truncated false count", func(r *DiagnosticReport) { r.Summary.TruncatedEventCount = 1 }, ValidationCodeReportInvariant},
		{"truncated true without count", func(r *DiagnosticReport) { r.Summary.Truncated = true }, ValidationCodeReportInvariant},
		{"invalid stored event", func(r *DiagnosticReport) { r.Events[0].Event = "Not-Valid" }, ValidationCodeInvalidReportEvent},
		{"mismatched event range", func(r *DiagnosticReport) { r.Summary.EventTimeRange.From = time.Unix(1, 0) }, ValidationCodeReportInvariant},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			report := validDiagnosticReport()
			test.mutate(&report)
			code, ok := ValidationCodeOf(ValidateDiagnosticReport(report, limits))
			if !ok || code != test.code {
				t.Fatalf("code = %q, want %q", code, test.code)
			}
		})
	}
}
