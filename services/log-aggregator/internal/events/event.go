package events

import (
	"bytes"
	"encoding/json"
	"io"
)

// Event is the canonical decoded observability envelope. Event-specific data
// belongs in Fields; top-level fields are reserved for the shared contract.
type Event struct {
	EventID                 string         `json:"event_id"`
	Timestamp               string         `json:"timestamp"`
	Level                   string         `json:"level"`
	Event                   string         `json:"event"`
	Service                 string         `json:"service"`
	SchemaVersion           int            `json:"schema_version"`
	ServiceInstanceID       string         `json:"service_instance_id"`
	Environment             string         `json:"environment,omitempty"`
	BuildVersion            string         `json:"build_version,omitempty"`
	Category                string         `json:"category,omitempty"`
	Message                 string         `json:"message,omitempty"`
	TraceID                 string         `json:"trace_id,omitempty"`
	RequestID               string         `json:"request_id,omitempty"`
	SessionID               string         `json:"session_id,omitempty"`
	RoomID                  string         `json:"room_id,omitempty"`
	MatchID                 string         `json:"match_id,omitempty"`
	PlayerID                string         `json:"player_id,omitempty"`
	AccountID               string         `json:"account_id,omitempty"`
	Route                   string         `json:"route,omitempty"`
	PacketType              string         `json:"packet_type,omitempty"`
	ErrorCode               string         `json:"error_code,omitempty"`
	FailureMode             string         `json:"failure_mode,omitempty"`
	DurationMS              *float64       `json:"duration_ms,omitempty"`
	DegradedState           string         `json:"degraded_state,omitempty"`
	IdempotencyKey          string         `json:"idempotency_key,omitempty"`
	DiagnosticReportID      string         `json:"diagnostic_report_id,omitempty"`
	BatchID                 string         `json:"batch_id,omitempty"`
	SourceEventID           string         `json:"source_event_id,omitempty"`
	SourceService           string         `json:"source_service,omitempty"`
	SourceServiceInstanceID string         `json:"source_service_instance_id,omitempty"`
	WorkerID                string         `json:"worker_id,omitempty"`
	ProcessID               *int64         `json:"process_id,omitempty"`
	AuditRequired           bool           `json:"audit_required,omitempty"`
	AuditType               string         `json:"audit_type,omitempty"`
	ActorID                 string         `json:"actor_id,omitempty"`
	ActorType               string         `json:"actor_type,omitempty"`
	TargetType              string         `json:"target_type,omitempty"`
	TargetID                string         `json:"target_id,omitempty"`
	Action                  string         `json:"action,omitempty"`
	ReasonCode              string         `json:"reason_code,omitempty"`
	CaseID                  string         `json:"case_id,omitempty"`
	TransactionID           string         `json:"transaction_id,omitempty"`
	ResultID                string         `json:"result_id,omitempty"`
	AuditEventID            string         `json:"audit_event_id,omitempty"`
	Fields                  map[string]any `json:"fields,omitempty"`
}

// Decode strictly decodes exactly one JSON object and validates its envelope.
func Decode(data []byte) (Event, error) {
	var fields map[string]json.RawMessage
	objectDecoder := json.NewDecoder(bytes.NewReader(data))
	if err := objectDecoder.Decode(&fields); err != nil || fields == nil {
		return Event{}, validationError(CodeMalformedJSON)
	}
	for key := range fields {
		if !allowedTopLevelFields[key] {
			return Event{}, validationError(CodeUnknownField)
		}
	}

	var event Event
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&event); err != nil {
		return Event{}, validationError(CodeMalformedJSON)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err == nil || err != io.EOF {
		return Event{}, validationError(CodeTrailingJSON)
	}
	if err := Validate(event); err != nil {
		return Event{}, err
	}
	return event, nil
}

var allowedTopLevelFields = map[string]bool{
	"event_id": true, "timestamp": true, "level": true, "event": true,
	"service": true, "schema_version": true, "service_instance_id": true,
	"environment": true, "build_version": true, "category": true, "message": true,
	"trace_id": true, "request_id": true, "session_id": true, "room_id": true,
	"match_id": true, "player_id": true, "account_id": true, "route": true,
	"packet_type": true, "error_code": true, "failure_mode": true, "duration_ms": true,
	"degraded_state": true, "idempotency_key": true, "diagnostic_report_id": true,
	"batch_id": true, "source_event_id": true, "source_service": true,
	"source_service_instance_id": true, "worker_id": true, "process_id": true,
	"audit_required": true, "audit_type": true, "actor_id": true, "actor_type": true,
	"target_type": true, "target_id": true, "action": true, "reason_code": true,
	"case_id": true, "transaction_id": true, "result_id": true, "audit_event_id": true,
	"fields": true,
}
