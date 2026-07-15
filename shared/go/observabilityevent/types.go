package observabilityevent

import (
	"io"
	"time"
)

// Sink accepts one fully serialized canonical record and a safe human-readable
// console rendering. Implementations must not reshape the JSON bytes.
type Sink interface {
	WriteRecord(jsonLine []byte, consoleLine string) error
}

// Config binds an emitter to one generated service identity and one output
// sink. Service is a stable registry key, not an emitted display name.
type Config struct {
	Service           ServiceKey
	Environment       string
	BuildVersion      string
	ServiceInstanceID string
	WorkerID          string
	PID               int
	Development       bool
	Sink              Sink
	WarningWriter     io.Writer
	Now               func() time.Time
	NewEventID        func() (string, error)
}

// Context contains only schema-approved optional correlation fields.
type Context struct {
	TraceID            string
	SessionID          string
	RoomID             string
	PlayerID           string
	AccountID          string
	MatchID            string
	RequestID          string
	DiagnosticReportID string
	AuditEventID       string
	Route              string
	PacketType         string
	DurationMS         *float64
}

// Fields is the bounded scalar free-form field collection.
type Fields map[string]any

// Request describes an ordinary canonical domain event.
type Request struct {
	Event   EventName
	Message string
	Context Context
	Fields  Fields
}

// LegacyRequest is available only to compatibility adapters. It is the sole
// path allowed to emit the generated bridge-only log_message event.
type LegacyRequest struct {
	Level       Level
	Category    string
	LegacyEvent string
	Message     string
	Context     Context
	Fields      Fields
}

// Result reports one emission attempt without exposing rejected values.
type Result struct {
	Accepted      bool
	Redacted      bool
	RejectionCode RejectionCode
	RejectedKey   string
	WriteFailed   bool
}

// Status is the emitter-owned operational state.
type Status struct {
	AcceptedCount     uint64
	RejectedCount     uint64
	RedactedCount     uint64
	WriteFailureCount uint64
	LastRejectionCode RejectionCode
	LastWriteError    string
}
