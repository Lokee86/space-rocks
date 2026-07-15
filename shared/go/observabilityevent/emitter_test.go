package observabilityevent

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	testInstanceID = "550e8400-e29b-41d4-a716-446655440010"
	testEventID    = "550e8400-e29b-41d4-a716-446655440011"
	testTraceID    = "550e8400-e29b-41d4-a716-446655440012"
)

type captureSink struct {
	payloads [][]byte
	console  []string
	err      error
}

func (sink *captureSink) WriteRecord(payload []byte, console string) error {
	if sink.err != nil {
		return sink.err
	}
	sink.payloads = append(sink.payloads, append([]byte(nil), payload...))
	sink.console = append(sink.console, console)
	return nil
}

func testEmitter(t *testing.T, sink Sink, service ServiceKey) *Emitter {
	t.Helper()
	emitter, err := New(Config{
		Service: service, Environment: "test", BuildVersion: "test-build",
		ServiceInstanceID: testInstanceID, WorkerID: "worker-1", PID: 42,
		Sink: sink, WarningWriter: io.Discard,
		Now:        func() time.Time { return time.Date(2026, 7, 14, 1, 2, 3, 0, time.UTC) },
		NewEventID: func() (string, error) { return testEventID, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	return emitter
}

func TestEmitterWritesOnlyCanonicalEnvelope(t *testing.T) {
	sink := &captureSink{}
	emitter := testEmitter(t, sink, ServiceKeyApiServer)
	result := emitter.Emit(Request{Event: EventNameBuildVersionLoaded, Message: "loaded", Fields: Fields{"mode": "test"}})
	if !result.Accepted || len(sink.payloads) != 1 {
		t.Fatalf("result=%+v payloads=%d", result, len(sink.payloads))
	}
	var record map[string]any
	if err := json.Unmarshal(sink.payloads[0], &record); err != nil {
		t.Fatal(err)
	}
	for key := range record {
		if _, ok := FieldDefinitions[FieldName(key)]; !ok {
			t.Fatalf("unexpected canonical field %q", key)
		}
	}
	for _, key := range RequiredFields {
		if _, ok := record[string(key)]; !ok {
			t.Fatalf("missing required canonical field %q", key)
		}
	}
	if record["service"] != ServiceNameApiServer || record["event_id"] != testEventID {
		t.Fatalf("unexpected identity fields: %#v", record)
	}
	if _, ok := record["msg"]; ok {
		t.Fatal("legacy slog msg field leaked into canonical envelope")
	}
}

func TestBridgeOnlyEventIsConfinedToLegacyPath(t *testing.T) {
	sink := &captureSink{}
	emitter := testEmitter(t, sink, ServiceKeyGameServer)
	ordinary := emitter.Emit(Request{Event: EventNameLogMessage})
	if ordinary.RejectionCode != RejectionCodeBridgeEventForbidden {
		t.Fatalf("ordinary result=%+v", ordinary)
	}
	legacy := emitter.EmitLegacy(LegacyRequest{Level: LevelInfo, Category: "game", Message: "legacy"})
	if !legacy.Accepted || len(sink.payloads) != 1 {
		t.Fatalf("legacy result=%+v payloads=%d", legacy, len(sink.payloads))
	}
}

func TestGeneratedRedactionPolicyRedactsOrRejectsWithoutLeakage(t *testing.T) {
	sink := &captureSink{}
	emitter := testEmitter(t, sink, ServiceKeyApiServer)
	redacted := emitter.Emit(Request{Event: EventNameBuildVersionLoaded, Fields: Fields{"private_profile": "sensitive"}})
	if !redacted.Accepted || !redacted.Redacted {
		t.Fatalf("redacted result=%+v", redacted)
	}
	if !strings.Contains(string(sink.payloads[0]), RedactionReplacementMarker) || strings.Contains(string(sink.payloads[0]), "sensitive") {
		t.Fatalf("unexpected redacted payload %s", sink.payloads[0])
	}
	before := len(sink.payloads)
	rejected := emitter.Emit(Request{Event: EventNameBuildVersionLoaded, Fields: Fields{"password": "do-not-leak"}})
	if rejected.RejectionCode != RejectionCodeUnsafeField || len(sink.payloads) != before {
		t.Fatalf("rejected result=%+v payloads=%d", rejected, len(sink.payloads))
	}
}

func TestStableValidationRejectionCodes(t *testing.T) {
	emitter := testEmitter(t, &captureSink{}, ServiceKeyApiServer)
	cases := []struct {
		name    string
		request Request
		want    RejectionCode
	}{
		{"trace required", Request{Event: EventNameApiRequestStarted}, RejectionCodeTraceRequired},
		{"invalid uuid", Request{Event: EventNameApiRequestStarted, Context: Context{TraceID: "bad"}}, RejectionCodeInvalidUuid},
		{"service", Request{Event: EventNameStorageBackendSelected}, RejectionCodeServiceNotAllowed},
		{"key", Request{Event: EventNameBuildVersionLoaded, Fields: Fields{"Bad-Key": 1}}, RejectionCodeInvalidFieldKey},
		{"type", Request{Event: EventNameBuildVersionLoaded, Fields: Fields{"items": []string{}}}, RejectionCodeInvalidFieldType},
		{"null", Request{Event: EventNameBuildVersionLoaded, Fields: Fields{"item": nil}}, RejectionCodeNullNotAllowed},
		{"string", Request{Event: EventNameBuildVersionLoaded, Fields: Fields{"item": strings.Repeat("x", MaxFreeFormValueBytes+1)}}, RejectionCodeStringLimitExceeded},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := emitter.Emit(test.request).RejectionCode; got != test.want {
				t.Fatalf("got %q want %q", got, test.want)
			}
		})
	}
	// Distinct keys are required to exercise the generated field-count limit.
	fields := Fields{}
	for index := 0; index <= MaxFreeFormFields; index++ {
		fields["item_"+string(rune('a'+index))] = index
	}
	if got := emitter.Emit(Request{Event: EventNameBuildVersionLoaded, Fields: fields}).RejectionCode; got != RejectionCodeFieldLimitExceeded {
		t.Fatalf("field count got %q", got)
	}
}

func TestWriterFailureIsNonRaisingAndOperationallyVisible(t *testing.T) {
	emitter := testEmitter(t, &captureSink{err: errors.New("disk unavailable")}, ServiceKeyApiServer)
	result := emitter.Emit(Request{Event: EventNameBuildVersionLoaded})
	if !result.WriteFailed || result.RejectionCode != RejectionCodeWriteFailed {
		t.Fatalf("result=%+v", result)
	}
	status := emitter.Status()
	if status.WriteFailureCount != 1 || !strings.Contains(status.LastWriteError, "disk unavailable") {
		t.Fatalf("status=%+v", status)
	}
}

func TestTraceRequiredEventAcceptsValidTrace(t *testing.T) {
	emitter := testEmitter(t, &captureSink{}, ServiceKeyApiServer)
	result := emitter.Emit(Request{Event: EventNameApiRequestStarted, Context: Context{TraceID: testTraceID}})
	if !result.Accepted {
		t.Fatalf("result=%+v", result)
	}
}

func TestSharedCrossLanguageFixturesHaveMatchingOutcomes(t *testing.T) {
	payload, err := os.ReadFile("../../contracts/observability/fixtures/emitter_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			ID            string         `json:"id"`
			Mode          string         `json:"mode"`
			Event         string         `json:"event"`
			Level         string         `json:"level"`
			Category      string         `json:"category"`
			Message       string         `json:"message"`
			Context       map[string]any `json:"context"`
			Fields        Fields         `json:"fields"`
			Accepted      bool           `json:"accepted"`
			Redacted      bool           `json:"redacted"`
			RejectionCode RejectionCode  `json:"rejection_code"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(payload, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range fixture.Cases {
		t.Run(testCase.ID, func(t *testing.T) {
			emitter := testEmitter(t, &captureSink{}, ServiceKeyApiServer)
			var result Result
			if testCase.Mode == "legacy" {
				result = emitter.EmitLegacy(LegacyRequest{
					Level: Level(testCase.Level), Category: testCase.Category,
					Message: testCase.Message, Fields: testCase.Fields,
				})
			} else {
				context := Context{}
				if value, ok := testCase.Context["trace_id"].(string); ok {
					context.TraceID = value
				}
				result = emitter.Emit(Request{
					Event: EventName(testCase.Event), Message: testCase.Message,
					Context: context, Fields: testCase.Fields,
				})
			}
			if result.Accepted != testCase.Accepted || result.Redacted != testCase.Redacted || result.RejectionCode != testCase.RejectionCode {
				t.Fatalf("result=%+v fixture=%+v", result, testCase)
			}
		})
	}
}
