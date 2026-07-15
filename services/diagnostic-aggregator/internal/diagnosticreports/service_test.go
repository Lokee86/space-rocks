package diagnosticreports

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/diagnostics"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/operationcontext"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/redaction"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

type memoryStore struct {
	report          storage.Report
	getErr, saveErr error
}

type recordingEmitter struct {
	requests []observability.Request
	result   observability.Result
}

func (e *recordingEmitter) Emit(request observability.Request) observability.Result {
	e.requests = append(e.requests, request)
	return e.result
}

func (m *memoryStore) Save(_ context.Context, report storage.Report) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.report = report
	return nil
}
func (m *memoryStore) Get(_ context.Context, id string) (storage.Report, error) {
	if m.getErr != nil {
		return storage.Report{}, m.getErr
	}
	if m.report.DiagnosticReportID != id {
		return storage.Report{}, storage.ErrReportNotFound
	}
	return m.report, nil
}
func (*memoryStore) DeleteExpired(context.Context, time.Time) (int, error) { return 0, nil }
func (*memoryStore) Close() error                                          { return nil }

func testService(t *testing.T, store storage.ReportStore) *Service {
	return testServiceWithEmitter(t, store, &recordingEmitter{result: observability.Result{Accepted: true}})
}

func testServiceWithEmitter(t *testing.T, store storage.ReportStore, emitter EventEmitter) *Service {
	t.Helper()
	limits := diagnostics.SubmissionLimits{MaxEmbeddedEvents: 4, MaxUserDescriptionBytes: 100, MaxFailureMessageBytes: 100, MaxContextStringBytes: 100}
	service, err := NewService(store, limits, redaction.DefaultPolicy(), func() time.Time { return time.Unix(100, 0).UTC() }, func() (string, error) { return "550e8400-e29b-41d4-a716-446655440000", nil }, emitter)
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func validSubmission() diagnostics.DiagnosticSubmission {
	return diagnostics.DiagnosticSubmission{ReportVersion: 1, Trigger: diagnostics.DiagnosticTriggerManualBugReport, SubmittedAt: time.Date(2026, 7, 13, 12, 1, 0, 0, time.UTC), Source: diagnostics.DiagnosticSourceContext{Service: "client", ServiceInstanceID: "550e8400-e29b-41d4-a716-446655440001", Environment: "test", BuildVersion: "1.0", Platform: "windows"}, Events: []json.RawMessage{
		json.RawMessage(`{"event_id":"550e8400-e29b-41d4-a716-446655440002","timestamp":"2026-07-13T12:02:00Z","level":"info","event":"ingest_accepted","service":"z-service","schema_version":1,"service_instance_id":"550e8400-e29b-41d4-a716-446655440001","environment":"prod","build_version":"2"}`),
		json.RawMessage(`{"event_id":"550e8400-e29b-41d4-a716-446655440003","timestamp":"2026-07-13T12:00:00Z","level":"info","event":"ingest_accepted","service":"a-service","schema_version":1,"service_instance_id":"550e8400-e29b-41d4-a716-446655440004","environment":"dev","build_version":"1"}`),
	}}
}

func TestCreateGetAndDeterministicSummary(t *testing.T) {
	store := &memoryStore{}
	service := testService(t, store)
	created, err := service.Create(context.Background(), validSubmission())
	if err != nil {
		t.Fatal(err)
	}
	if created.Summary.SubmittedEventCount != 2 || created.Summary.AcceptedEventCount != 2 {
		t.Fatalf("summary counts = %#v", created.Summary)
	}
	if got, want := created.Summary.Services, []string{"a-service", "z-service"}; !equalStrings(got, want) {
		t.Fatalf("services=%v", got)
	}
	if !created.Summary.EventTimeRange.From.Equal(time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)) || !created.Summary.EventTimeRange.To.Equal(time.Date(2026, 7, 13, 12, 2, 0, 0, time.UTC)) {
		t.Fatalf("range=%#v", created.Summary.EventTimeRange)
	}
	loaded, err := service.Get(context.Background(), created.DiagnosticReportID)
	if err != nil || loaded.DiagnosticReportID != created.DiagnosticReportID {
		t.Fatalf("get=%#v err=%v", loaded, err)
	}
}

func TestCreateRejectsUnsafeWithoutValueLeak(t *testing.T) {
	submission := validSubmission()
	submission.Events[0] = json.RawMessage(`{"event_id":"550e8400-e29b-41d4-a716-446655440002","timestamp":"2026-07-13T12:02:00Z","level":"info","event":"ingest_accepted","service":"z-service","schema_version":1,"service_instance_id":"550e8400-e29b-41d4-a716-446655440001","fields":{"password":"secret-value"}}`)
	_, err := testService(t, &memoryStore{}).Create(context.Background(), submission)
	if !errors.Is(err, ErrRejected) || strings.Contains(err.Error(), "secret-value") {
		t.Fatalf("error=%v", err)
	}
}

func TestGetInvalidMissingAndUnavailable(t *testing.T) {
	service := testService(t, &memoryStore{})
	if !errors.Is(func() error { _, e := service.Get(context.Background(), "bad"); return e }(), ErrInvalidID) {
		t.Fatal("invalid id not mapped")
	}
	if _, err := service.Get(context.Background(), "550e8400-e29b-41d4-a716-446655440000"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing=%v", err)
	}
	if _, err := testService(t, &memoryStore{getErr: errors.New("backend secret")}).Get(context.Background(), "550e8400-e29b-41d4-a716-446655440000"); !errors.Is(err, ErrUnavailable) || len(err.Error()) > 100 {
		t.Fatalf("unavailable=%v", err)
	}
}

func TestGetRejectsTrailingPersistedJSON(t *testing.T) {
	store := &memoryStore{}
	service := testService(t, store)
	created, err := service.Create(context.Background(), validSubmission())
	if err != nil {
		t.Fatal(err)
	}
	store.report.RawJSON = append(store.report.RawJSON, []byte(" {}")...)
	if _, err := service.Get(context.Background(), created.DiagnosticReportID); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("trailing payload error=%v", err)
	}
}

func TestNewServiceRequiresDependencies(t *testing.T) {
	limits := diagnostics.SubmissionLimits{MaxEmbeddedEvents: 1, MaxUserDescriptionBytes: 1, MaxFailureMessageBytes: 1, MaxContextStringBytes: 1}
	if _, err := NewService(nil, limits, redaction.Policy{}, time.Now, func() (string, error) { return "", nil }, &recordingEmitter{}); err == nil {
		t.Fatal("expected dependency validation error")
	}
}
func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func testOperationContext() context.Context {
	return operationcontext.With(context.Background(), operationcontext.Values{TraceID: "550e8400-e29b-41d4-a716-446655440010", RequestID: "550e8400-e29b-41d4-a716-446655440011", Route: "/v1/diagnostic-reports"})
}

func TestCreateEmitsAcceptedThenStoredWithOwnedContext(t *testing.T) {
	emitter := &recordingEmitter{result: observability.Result{Accepted: true}}
	service := testServiceWithEmitter(t, &memoryStore{}, emitter)
	created, err := service.Create(testOperationContext(), validSubmission())
	if err != nil {
		t.Fatal(err)
	}
	if len(emitter.requests) != 2 {
		t.Fatalf("requests=%#v", emitter.requests)
	}
	accepted, stored := emitter.requests[0], emitter.requests[1]
	if accepted.Event != observability.EventNameAggregatorEventAccepted || stored.Event != observability.EventNameDiagnosticReportStored {
		t.Fatalf("events=%s,%s", accepted.Event, stored.Event)
	}
	for _, request := range emitter.requests {
		if request.Context.TraceID != "550e8400-e29b-41d4-a716-446655440010" || request.Context.RequestID != "550e8400-e29b-41d4-a716-446655440011" || request.Context.Route != "/v1/diagnostic-reports" || request.Context.DiagnosticReportID != created.DiagnosticReportID {
			t.Fatalf("context=%#v", request.Context)
		}
	}
	if accepted.Fields["submitted_event_count"] != 2 || stored.Fields["accepted_event_count"] != uint64(2) || stored.Fields["source_service"] != "client" {
		t.Fatalf("accepted=%#v stored=%#v", accepted.Fields, stored.Fields)
	}
}

func TestCreateSaveFailureEmitsOnceAndEmissionFailureDoesNotChangeError(t *testing.T) {
	emitter := &recordingEmitter{result: observability.Result{WriteFailed: true}}
	service := testServiceWithEmitter(t, &memoryStore{saveErr: errors.New("backend secret")}, emitter)
	_, err := service.Create(testOperationContext(), validSubmission())
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("error=%v", err)
	}
	if len(emitter.requests) != 2 {
		t.Fatalf("requests=%#v", emitter.requests)
	}
	if emitter.requests[0].Event != observability.EventNameAggregatorEventAccepted || emitter.requests[1].Event != observability.EventNameAggregatorStorageFailed {
		t.Fatalf("events=%s,%s", emitter.requests[0].Event, emitter.requests[1].Event)
	}
	if emitter.requests[1].Context.DiagnosticReportID == "" || emitter.requests[1].Fields["failure_stage"] != "report_save" {
		t.Fatalf("failure=%#v", emitter.requests[1])
	}
}
