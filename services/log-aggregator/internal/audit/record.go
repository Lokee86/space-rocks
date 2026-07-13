package audit

import (
	"encoding/json"
	"time"
)

const RecordVersion = "1"

type Record struct {
	Version                 string          `json:"version"`
	AuditEventID            string          `json:"audit_event_id"`
	SourceEventID           string          `json:"source_event_id"`
	SourceTimestamp         time.Time       `json:"source_timestamp"`
	PromotedAt              time.Time       `json:"promoted_at"`
	TraceID                 string          `json:"trace_id"`
	AuditType               string          `json:"audit_type"`
	SourceService           string          `json:"source_service"`
	SourceServiceInstanceID string          `json:"source_service_instance_id"`
	ActorID                 string          `json:"actor_id"`
	ActorType               string          `json:"actor_type"`
	TargetType              string          `json:"target_type"`
	TargetID                string          `json:"target_id"`
	Action                  string          `json:"action"`
	ReasonCode              string          `json:"reason_code"`
	CaseID                  string          `json:"case_id"`
	TransactionID           string          `json:"transaction_id"`
	ResultID                string          `json:"result_id"`
	MatchID                 string          `json:"match_id"`
	AccountID               string          `json:"account_id"`
	Payload                 json.RawMessage `json:"payload"`
}

func cloneRecord(record Record) Record {
	record.Payload = append(json.RawMessage(nil), record.Payload...)
	return record
}
