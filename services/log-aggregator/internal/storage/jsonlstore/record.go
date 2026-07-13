package jsonlstore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage"
)

// diskRecord is the stable JSONL representation. Payload remains raw JSON so
// persisted records retain their original nested value rather than a string.
type diskRecord struct {
	EventID            string          `json:"event_id"`
	SchemaVersion      string          `json:"schema_version"`
	Timestamp          time.Time       `json:"timestamp"`
	IngestedAt         time.Time       `json:"ingested_at"`
	Environment        string          `json:"environment"`
	BuildVersion       string          `json:"build_version"`
	Service            string          `json:"service"`
	ServiceInstanceID  string          `json:"service_instance_id"`
	Category           string          `json:"category"`
	Level              string          `json:"level"`
	Event              string          `json:"event"`
	TraceID            string          `json:"trace_id"`
	RequestID          string          `json:"request_id"`
	SessionID          string          `json:"session_id"`
	RoomID             string          `json:"room_id"`
	MatchID            string          `json:"match_id"`
	PlayerID           string          `json:"player_id"`
	AccountID          string          `json:"account_id"`
	DiagnosticReportID string          `json:"diagnostic_report_id"`
	AuditEventID       string          `json:"audit_event_id"`
	IdempotencyKey     string          `json:"idempotency_key"`
	AuditRequired      bool            `json:"audit_required"`
	Payload            json.RawMessage `json:"payload"`
}

// EncodeRecord encodes one storage record as exactly one newline-terminated
// JSON object.
func EncodeRecord(record storage.Record) ([]byte, error) {
	encoded, err := json.Marshal(diskRecord{
		EventID: record.EventID, SchemaVersion: record.SchemaVersion,
		Timestamp: record.Timestamp, IngestedAt: record.IngestedAt,
		Environment: record.Environment, BuildVersion: record.BuildVersion,
		Service: record.Service, ServiceInstanceID: record.ServiceInstanceID,
		Category: record.Category, Level: record.Level, Event: record.Event,
		TraceID: record.TraceID, RequestID: record.RequestID, SessionID: record.SessionID,
		RoomID: record.RoomID, MatchID: record.MatchID, PlayerID: record.PlayerID,
		AccountID: record.AccountID, DiagnosticReportID: record.DiagnosticReportID,
		AuditEventID: record.AuditEventID, IdempotencyKey: record.IdempotencyKey,
		AuditRequired: record.AuditRequired, Payload: payloadOrNull(record.Payload),
	})
	if err != nil {
		return nil, fmt.Errorf("jsonlstore: encode record: %w", err)
	}
	return append(encoded, '\n'), nil
}

// DecodeRecord decodes one complete JSONL record and preserves its raw payload.
func DecodeRecord(line []byte) (storage.Record, error) {
	line = bytes.TrimSuffix(line, []byte{'\n'})
	var disk diskRecord
	if err := json.Unmarshal(line, &disk); err != nil {
		return storage.Record{}, fmt.Errorf("jsonlstore: decode record: %w", err)
	}
	payload := append([]byte(nil), disk.Payload...)
	if bytes.Equal(payload, []byte("null")) {
		payload = nil
	}
	return storage.Record{
		EventID: disk.EventID, SchemaVersion: disk.SchemaVersion,
		Timestamp: disk.Timestamp, IngestedAt: disk.IngestedAt,
		Environment: disk.Environment, BuildVersion: disk.BuildVersion,
		Service: disk.Service, ServiceInstanceID: disk.ServiceInstanceID,
		Category: disk.Category, Level: disk.Level, Event: disk.Event,
		TraceID: disk.TraceID, RequestID: disk.RequestID, SessionID: disk.SessionID,
		RoomID: disk.RoomID, MatchID: disk.MatchID, PlayerID: disk.PlayerID,
		AccountID: disk.AccountID, DiagnosticReportID: disk.DiagnosticReportID,
		AuditEventID: disk.AuditEventID, IdempotencyKey: disk.IdempotencyKey,
		AuditRequired: disk.AuditRequired, Payload: payload,
	}, nil
}

func payloadOrNull(payload json.RawMessage) json.RawMessage {
	if len(payload) == 0 {
		return json.RawMessage("null")
	}
	return payload
}
