package query

import (
	"net/http"
	"strconv"
	"time"
)

const EventsPath = "/v1/events"

func NewEventsHandler(querier EventQuerier) http.Handler {
	return NewEventsHandlerWithPolicy(querier, LimitPolicy{})
}

func NewEventsHandlerWithPolicy(querier EventQuerier, policy LimitPolicy) http.Handler {
	policy = policy.normalized()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"code": "method_not_allowed"})
			return
		}

		filter, err := parseFilter(r, policy)
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

func parseFilter(r *http.Request, policy LimitPolicy) (Filter, error) {
	policy = policy.normalized()
	values := r.URL.Query()
	allowed := map[string]bool{
		"from": true, "to": true, "schema_version": true, "environment": true,
		"build_version": true, "service": true, "service_instance_id": true,
		"category": true, "level": true, "event": true, "trace_id": true, "request_id": true,
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
		"schema_version", "environment", "build_version", "category",
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
		SchemaVersion:      values.Get("schema_version"),
		Environment:        values.Get("environment"),
		BuildVersion:       values.Get("build_version"),
		Service:            values.Get("service"),
		ServiceInstanceID:  values.Get("service_instance_id"),
		Category:           values.Get("category"),
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
	for _, value := range []string{filter.ServiceInstanceID, filter.TraceID, filter.EventID, filter.DiagnosticReportID, filter.AuditEventID} {
		if value != "" && !validUUID(value) {
			return Filter{}, errInvalidFilter
		}
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
		if parseErr != nil || value <= 0 || value > policy.Maximum {
			return Filter{}, errInvalidFilter
		}
		filter.Limit = value
	} else {
		filter.Limit = policy.Default
	}
	return filter, nil
}

var errInvalidFilter = &filterError{}

type filterError struct{}

func (*filterError) Error() string { return "invalid_filter" }
