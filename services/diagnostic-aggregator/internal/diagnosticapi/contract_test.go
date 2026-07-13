package diagnosticapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestCreateRequestUsesAgreedJSONContract(t *testing.T) {
	at := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	request := DiagnosticReportCreateRequest{
		ReportVersion: 1, Trigger: "manual_bug_report", SubmittedAt: at,
		Source:          SourceContext{Service: "client", ServiceInstanceID: "client-1", Environment: "test", BuildVersion: "build-7", Platform: "desktop"},
		Correlation:     CorrelationContext{TraceID: "trace-1", RequestID: "request-1", SessionID: "session-1", RoomID: "room-1", MatchID: "match-1", PlayerID: "player-1", AccountID: "account-1"},
		UserDescription: "reproduced after reconnect",
		Failure:         &FailureContext{FailureMode: "connection_lost", ErrorCode: "network_error", Component: "client", Message: "connection lost"},
		Events:          []json.RawMessage{json.RawMessage("{}")},
	}
	encoded, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"report_version", "trigger", "submitted_at", "source", "correlation", "user_description", "failure", "events"} {
		if _, ok := fields[field]; !ok {
			t.Fatalf("missing JSON field %q: %s", field, encoded)
		}
	}
	for _, field := range []string{"source_context", "correlation_context", "failure_context", "report_id"} {
		if _, ok := fields[field]; ok {
			t.Fatalf("legacy JSON field %q present: %s", field, encoded)
		}
	}
	var version int
	if err := json.Unmarshal(fields["report_version"], &version); err != nil || version != 1 {
		t.Fatalf("report_version is not int 1: %s", fields["report_version"])
	}
	var events []json.RawMessage
	if err := json.Unmarshal(fields["events"], &events); err != nil || len(events) != 1 {
		t.Fatalf("events contract mismatch: %s", fields["events"])
	}
	var roundTrip DiagnosticReportCreateRequest
	if err := json.Unmarshal(encoded, &roundTrip); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(roundTrip, request) {
		t.Fatalf("round trip mismatch: got %#v want %#v", roundTrip, request)
	}
}

func TestReportSummaryUsesAgreedKeys(t *testing.T) {
	at := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	encoded, err := json.Marshal(ReportSummary{SubmittedEventCount: 2, AcceptedEventCount: 1, RejectedEventCount: 1, RedactedFieldCount: 3, DroppedFieldCount: 4, Truncated: true, TruncatedEventCount: 5, Services: []string{"client"}, ServiceInstances: []string{"client-1"}, Environments: []string{"test"}, Builds: []string{"build-7"}, EventTimeRange: EventTimeRange{From: at, To: at}})
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"submitted_event_count", "accepted_event_count", "rejected_event_count", "redacted_field_count", "dropped_field_count", "truncated", "truncated_event_count", "services", "service_instances", "environments", "builds", "event_time_range"} {
		if _, ok := fields[field]; !ok {
			t.Fatalf("missing summary field %q: %s", field, encoded)
		}
	}
	var rangeFields map[string]json.RawMessage
	if err := json.Unmarshal(fields["event_time_range"], &rangeFields); err != nil {
		t.Fatal(err)
	}
	if _, ok := rangeFields["from"]; !ok {
		t.Fatal("missing event_time_range.from")
	}
	if _, ok := rangeFields["to"]; !ok {
		t.Fatal("missing event_time_range.to")
	}
}

func TestRoutesAndHandlerContract(t *testing.T) {
	if DiagnosticReportsPath != "/v1/diagnostic-reports" || DiagnosticReportItemPathPrefix != DiagnosticReportsPath+"/" {
		t.Fatal("route constants mismatch")
	}
	if !RequestAuthorizer(func(r *http.Request) bool { return r.Method == http.MethodPost })(&http.Request{Method: http.MethodPost}) {
		t.Fatal("authorizer mismatch")
	}
	if (HandlerConfig{MaxRequestBytes: 1024}).MaxRequestBytes != 1024 {
		t.Fatal("max request bytes mismatch")
	}
	if (ErrorResponse{Code: ErrorCodeMalformedJSON}).Code != "malformed_json" {
		t.Fatal("error response mismatch")
	}
}

func TestStableErrorCodesAndSentinels(t *testing.T) {
	codes := []string{ErrorCodeMethodNotAllowed, ErrorCodeUnauthorized, ErrorCodeUnsupportedMediaType, ErrorCodeMalformedJSON, ErrorCodeTrailingJSON, ErrorCodeRequestTooLarge, ErrorCodeInvalidDiagnosticReport, ErrorCodeDiagnosticReportRejected, ErrorCodeInvalidReportID, ErrorCodeDiagnosticReportNotFound, ErrorCodeServiceUnavailable}
	for _, code := range codes {
		if code == "" {
			t.Fatal("empty error code")
		}
	}
	for _, err := range []error{ErrInvalidReport, ErrRejectedReport, ErrInvalidReportID, ErrReportNotFound, ErrUnavailable} {
		if err == nil || !errors.Is(err, err) {
			t.Fatal("missing service sentinel")
		}
	}
}
