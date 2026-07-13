package jsonlstore

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage"
)

func TestRecordRoundTrip(t *testing.T) {
	payload := json.RawMessage("{\"nested\":{\"count\":3},\"ok\":true}")
	if !json.Valid(payload) {
		t.Fatal("valid payload fixture is not valid JSON")
	}
	want := storage.Record{EventID: "evt-1", SchemaVersion: "v1", Timestamp: time.Date(2026, 7, 13, 12, 0, 1, 123, time.UTC), IngestedAt: time.Date(2026, 7, 13, 12, 0, 2, 456, time.UTC), Environment: "prod", BuildVersion: "build-7", Service: "game-server", ServiceInstanceID: "instance-1", Category: "network", Level: "warn", Event: "slow_write", TraceID: "trace-1", RequestID: "request-1", SessionID: "session-1", RoomID: "room-1", MatchID: "match-1", PlayerID: "Player-1", AccountID: "account-1", DiagnosticReportID: "report-1", AuditEventID: "audit-1", IdempotencyKey: "idem-1", AuditRequired: true, Payload: payload}
	encoded, err := EncodeRecord(want)
	if err != nil {
		t.Fatalf("EncodeRecord() error = %v", err)
	}
	if encoded[len(encoded)-1] != '\n' || bytes.Count(encoded, []byte{'\n'}) != 1 {
		t.Fatalf("encoded record is not exactly one newline-terminated object: %q", encoded)
	}
	encodedText := string(encoded)
	if !strings.Contains(encodedText, "\"nested\"") {
		t.Fatalf("payload was not preserved as raw JSON: %s", encoded)
	}
	if strings.Contains(encodedText, "\\\"nested\\\"") {
		t.Fatalf("payload was quoted or escaped: %s", encoded)
	}
	got, err := DecodeRecord(encoded)
	if err != nil {
		t.Fatalf("DecodeRecord() error = %v", err)
	}
	if got.EventID != want.EventID || got.SchemaVersion != want.SchemaVersion || !got.Timestamp.Equal(want.Timestamp) || !got.IngestedAt.Equal(want.IngestedAt) {
		t.Fatalf("identity/timestamps changed: %#v", got)
	}
	if got.Environment != want.Environment || got.BuildVersion != want.BuildVersion || got.Service != want.Service || got.ServiceInstanceID != want.ServiceInstanceID || got.Category != want.Category || got.Level != want.Level || got.Event != want.Event {
		t.Fatalf("primary projection changed: %#v", got)
	}
	if got.TraceID != want.TraceID || got.RequestID != want.RequestID || got.SessionID != want.SessionID || got.RoomID != want.RoomID || got.MatchID != want.MatchID || got.PlayerID != want.PlayerID || got.AccountID != want.AccountID {
		t.Fatalf("correlation projection changed: %#v", got)
	}
	if got.DiagnosticReportID != want.DiagnosticReportID || got.AuditEventID != want.AuditEventID || got.IdempotencyKey != want.IdempotencyKey || got.AuditRequired != want.AuditRequired || !bytes.Equal(got.Payload, want.Payload) {
		t.Fatalf("audit/payload projection changed: %#v", got)
	}
}

func TestDecodeRecordRejectsMalformedJSON(t *testing.T) {
	_, err := DecodeRecord([]byte("{\"event_id\":"))
	if err == nil || !strings.Contains(err.Error(), "decode record") {
		t.Fatalf("DecodeRecord() error = %v, want contextual error", err)
	}
}

func TestEncodeRecordRejectsMalformedPayload(t *testing.T) {
	malformed := json.RawMessage("{\"broken\"")
	if json.Valid(malformed) {
		t.Fatal("malformed payload fixture unexpectedly became valid")
	}
	_, err := EncodeRecord(storage.Record{Payload: malformed})
	if err == nil || !strings.Contains(err.Error(), "encode record") {
		t.Fatalf("EncodeRecord() error = %v, want contextual error", err)
	}
}

func TestRecordRoundTripPreservesNilPayload(t *testing.T) {
	encoded, err := EncodeRecord(storage.Record{EventID: "nil-payload"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeRecord(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Payload) != 0 {
		t.Fatalf("nil payload decoded as %q", got.Payload)
	}
}
