package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/events"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/redaction"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage"
	"github.com/google/uuid"
)

var ErrStorageAppend = errors.New("ingestion: storage append outcome unknown")

const (
	CodeEventTooLarge     = "event_too_large"
	CodeRedactionRejected = "redaction_rejected"
)

type Clock func() time.Time
type UUIDGenerator func() uuid.UUID

type Processor struct {
	store    storage.EventStore
	policy   redaction.Policy
	clock    Clock
	newUUID  UUIDGenerator
	maxBytes int
}

func NewProcessor(store storage.EventStore, policy redaction.Policy, clock Clock, newUUID UUIDGenerator, maxBytes int) *Processor {
	if clock == nil {
		clock = time.Now
	}
	if newUUID == nil {
		newUUID = uuid.New
	}
	return &Processor{store: store, policy: policy, clock: clock, newUUID: newUUID, maxBytes: maxBytes}
}

func (p *Processor) IngestBatch(ctx context.Context, rawEvents []json.RawMessage) (BatchResult, error) {
	result := BatchResult{BatchID: p.newUUID().String()}
	records := make([]storage.Record, 0, len(rawEvents))
	for index, raw := range rawEvents {
		if err := ctx.Err(); err != nil {
			return BatchResult{}, err
		}
		if p.maxBytes > 0 && len(raw) > p.maxBytes {
			result.Rejected = append(result.Rejected, rejection(index, CodeEventTooLarge, raw))
			continue
		}
		findings, err := redaction.Inspect(raw, p.policy)
		if err != nil {
			result.Rejected = append(result.Rejected, EventRejection{Index: index, Code: events.CodeMalformedJSON})
			continue
		}
		if len(findings) > 0 {
			finding := findings[0]
			result.Rejected = append(result.Rejected, EventRejection{Index: index, Code: CodeRedactionRejected, EventID: safeEventID(raw), RuleID: finding.RuleID, NormalizedFieldPath: finding.NormalizedFieldPath, ReasonCode: finding.ReasonCode})
			continue
		}
		event, err := events.Decode(raw)
		if err != nil {
			var validationErr *events.ValidationError
			code := events.CodeMalformedJSON
			if errors.As(err, &validationErr) {
				code = validationErr.Code
			}
			result.Rejected = append(result.Rejected, rejection(index, code, raw))
			continue
		}
		record, err := project(event, p.clock())
		if err != nil {
			result.Rejected = append(result.Rejected, rejection(index, "invalid_timestamp", raw))
			continue
		}
		records = append(records, record)
	}
	result.Accepted = len(records)
	if len(records) == 0 {
		return result, nil
	}
	if err := ctx.Err(); err != nil {
		return BatchResult{}, err
	}
	if p.store == nil {
		return BatchResult{}, fmt.Errorf("%w: nil event store", ErrStorageAppend)
	}
	if err := p.store.AppendBatch(ctx, records); err != nil {
		return BatchResult{}, fmt.Errorf("%w: %v", ErrStorageAppend, err)
	}
	return result, nil
}

func project(event events.Event, ingestedAt time.Time) (storage.Record, error) {
	timestamp, err := time.Parse(time.RFC3339, event.Timestamp)
	if err != nil {
		return storage.Record{}, err
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return storage.Record{}, err
	}
	return storage.Record{
		EventID: event.EventID, SchemaVersion: strconv.Itoa(event.SchemaVersion), Timestamp: timestamp,
		IngestedAt: ingestedAt.UTC(), Environment: event.Environment, BuildVersion: event.BuildVersion,
		Service: event.Service, ServiceInstanceID: event.ServiceInstanceID, Category: event.Category,
		Level: event.Level, Event: event.Event, TraceID: event.TraceID, RequestID: event.RequestID,
		SessionID: event.SessionID, RoomID: event.RoomID, MatchID: event.MatchID, PlayerID: event.PlayerID,
		AccountID: event.AccountID, DiagnosticReportID: event.DiagnosticReportID, AuditEventID: event.AuditEventID,
		IdempotencyKey: event.IdempotencyKey, AuditRequired: event.AuditRequired, Payload: payload,
	}, nil
}

func rejection(index int, code string, raw json.RawMessage) EventRejection {
	return EventRejection{Index: index, Code: code, EventID: safeEventID(raw)}
}

func safeEventID(raw json.RawMessage) string {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil {
		return ""
	}
	var value string
	if json.Unmarshal(object["event_id"], &value) != nil {
		return ""
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		return ""
	}
	return parsed.String()
}
