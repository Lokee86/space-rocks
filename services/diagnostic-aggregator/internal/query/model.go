package query

import (
	"context"
	"encoding/json"
	"time"
)

type Filter struct {
	From               time.Time
	To                 time.Time
	SchemaVersion      string
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
	EventID            string
	DiagnosticReportID string
	AuditEventID       string
	IdempotencyKey     string
	AuditRequired      *bool
	Limit              int
}

type Result struct {
	Events  []json.RawMessage
	Total   uint64
	Limited bool
}

type EventsResponse struct {
	Events  []json.RawMessage `json:"events"`
	Total   uint64             `json:"total"`
	Limited bool               `json:"limited"`
}

type EventQuerier interface {
	Query(context.Context, Filter) (Result, error)
}
