package query

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

const EventsPath = "/v1/events"

type Filter struct {
	From               time.Time
	To                 time.Time
	Service            string
	ServiceInstanceID  string
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
	Total   uint64           `json:"total"`
	Limited bool             `json:"limited"`
}

type EventQuerier interface {
	Query(context.Context, Filter) (Result, error)
}

func NewEventsHandler(querier EventQuerier) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"code": "method_not_allowed"})
			return
		}

		filter, err := parseFilter(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"code": err.Error()})
			return
		}
		result, err := querier.Query(r.Context(), filter)
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": "query_unavailable"})
			return
		}
		writeJSON(w, http.StatusOK, EventsResponse{Events: result.Events, Total: result.Total, Limited: result.Limited})
	})
}

func parseFilter(r *http.Request) (Filter, error) {
	values := r.URL.Query()
	allowed := map[string]bool{
		"from": true, "to": true, "service": true, "service_instance_id": true,
		"level": true, "event": true, "trace_id": true, "request_id": true,
		"session_id": true, "room_id": true, "match_id": true, "player_id": true,
		"account_id": true, "event_id": true, "diagnostic_report_id": true,
		"audit_event_id": true, "idempotency_key": true, "audit_required": true,
		"limit": true,
	}
	for name := range values {
		if !allowed[name] {
			return Filter{}, errInvalidFilter
		}
		if len(values[name]) != 1 {
			return Filter{}, errInvalidFilter
		}
	}
	for _, name := range []string{
		"service", "service_instance_id", "level", "event", "trace_id",
		"request_id", "session_id", "room_id", "match_id", "player_id",
		"account_id", "event_id", "diagnostic_report_id", "audit_event_id",
		"idempotency_key",
	} {
		if values[name] != nil && values[name][0] == "" {
			return Filter{}, errInvalidFilter
		}
	}

	filter := Filter{
		Service:            values.Get("service"),
		ServiceInstanceID:  values.Get("service_instance_id"),
		Level:              values.Get("level"),
		Event:              values.Get("event"),
		TraceID:            values.Get("trace_id"),
		RequestID:          values.Get("request_id"),
		SessionID:          values.Get("session_id"),
		RoomID:             values.Get("room_id"),
		MatchID:            values.Get("match_id"),
		PlayerID:           values.Get("player_id"),
		AccountID:          values.Get("account_id"),
		EventID:            values.Get("event_id"),
		DiagnosticReportID: values.Get("diagnostic_report_id"),
		AuditEventID:       values.Get("audit_event_id"),
		IdempotencyKey:     values.Get("idempotency_key"),
	}
	var err error
	if raw, present := values["from"]; present {
		if len(raw) != 1 || raw[0] == "" {
			return Filter{}, errInvalidFilter
		}
		filter.From, err = time.Parse(time.RFC3339, raw[0])
		if err != nil {
			return Filter{}, errInvalidFilter
		}
	}
	if raw, present := values["to"]; present {
		if len(raw) != 1 || raw[0] == "" {
			return Filter{}, errInvalidFilter
		}
		filter.To, err = time.Parse(time.RFC3339, raw[0])
		if err != nil {
			return Filter{}, errInvalidFilter
		}
	}
	if !filter.From.IsZero() && !filter.To.IsZero() && filter.From.After(filter.To) {
		return Filter{}, errInvalidFilter
	}
	if raw, present := values["audit_required"]; present {
		if len(raw) != 1 || raw[0] == "" {
			return Filter{}, errInvalidFilter
		}
		value, parseErr := strconv.ParseBool(raw[0])
		if parseErr != nil {
			return Filter{}, errInvalidFilter
		}
		filter.AuditRequired = &value
	}
	if raw, present := values["limit"]; present {
		if len(raw) != 1 || raw[0] == "" {
			return Filter{}, errInvalidFilter
		}
		value, parseErr := strconv.Atoi(raw[0])
		if parseErr != nil || value <= 0 {
			return Filter{}, errInvalidFilter
		}
		filter.Limit = value
	}
	return filter, nil
}

var errInvalidFilter = &filterError{}

type filterError struct{}

func (*filterError) Error() string { return "invalid_filter" }

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
