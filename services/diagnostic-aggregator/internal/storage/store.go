// Package storage defines the persistence boundary for normalized observability
// records. It intentionally contains no backend implementation or diagnostic
// bundle policy.
package storage

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// ErrClosed is returned when an operation is attempted on a closed store.
var ErrClosed = errors.New("storage: event store is closed")

// Record is the storage-owned projection of one normalized log event. The
// searchable fields are copied out of the normalized event so stores need not
// interpret payload JSON. UUID-backed identifiers remain validated canonical
// strings at this boundary; storage neither creates nor validates identifiers.
//
// Validation and redaction must be completed by the ingestion owner before a
// record is passed to EventStore.AppendBatch. A store must not be used to
// bypass those controls.
type Record struct {
	EventID            string
	SchemaVersion      string
	Timestamp          time.Time
	IngestedAt         time.Time
	Environment        string
	BuildVersion       string
	Service            string
	ServiceInstanceID  string
	Category           string
	Level              string
	Event              string
	TraceID            string
	RequestID          string
	SessionID          string
	RoomID             string
	MatchID            string
	PlayerID           string
	AccountID          string
	DiagnosticReportID string
	AuditEventID       string
	IdempotencyKey     string
	AuditRequired      bool
	Payload            json.RawMessage
}

// Query describes the storage-neutral filters supported by log consumers.
// From and To are inclusive time bounds when non-zero. Limit is a maximum
// result count; zero means the store's default policy. Query results are
// deterministic: records are ordered chronologically by Timestamp, with
// EventID as the tie-breaker.
type Query struct {
	From               time.Time
	To                 time.Time
	SchemaVersion      string
	Service            string
	ServiceInstanceID  string
	Environment        string
	BuildVersion       string
	Category           string
	Level              string
	Event              string
	TraceID            string
	RequestID          string
	SessionID          string
	RoomID             string
	MatchID            string
	PlayerID           string
	AccountID          string
	EventID            string
	DiagnosticReportID string
	AuditEventID       string
	IdempotencyKey     string
	AuditRequired      *bool
	Limit              int
}

// QueryResult is the stable result shape returned to query and diagnostic
// consumers. Records are chronologically ordered as specified by Query.
// Total is the number of matches before applying Limit. Limited reports that
// the returned Records were truncated. Diagnostic-bundle construction is
// outside the storage boundary.
type QueryResult struct {
	Records []Record
	Total   uint64
	Limited bool
}

// Status is the operational snapshot exposed to health and diagnostics
// consumers. Timestamps are zero when the store has no records.
type Status struct {
	Ready       bool
	Degraded    bool
	RecordCount uint64
	ByteCount   uint64
	Oldest      time.Time
	Newest      time.Time
}

// EventStore is the persistence seam for normalized observability records.
// AppendBatch must preserve input order. If it returns an error, callers must
// treat the write outcome as unknown: this contract does not promise
// transactional disk semantics. Implementations must preserve the projection
// supplied by ingestion and must not perform validation, redaction, ID
// generation, diagnostic-bundle policy, or expose backend-specific status
// fields through Status.
type EventStore interface {
	AppendBatch(ctx context.Context, records []Record) error
	Query(ctx context.Context, query Query) (QueryResult, error)
	Status(ctx context.Context) (Status, error)
	Close() error
}
