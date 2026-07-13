package diagnostics

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage"
)

const (
	BundleVersion     = "1"
	DefaultEventLimit = 100
	MaximumEventLimit = 1000
)

var (
	ErrInvalidTraceID       = errors.New("diagnostics: invalid trace id")
	ErrInvalidLimit         = errors.New("diagnostics: invalid event limit")
	ErrNoEvents             = errors.New("diagnostics: no events for trace id")
	ErrSanitizerRequired    = errors.New("diagnostics: sanitizer is required")
	ErrInvalidReportID          = errors.New("diagnostics: generated report id is not a uuid")
	ErrInvalidSanitizedPayload  = errors.New("diagnostics: sanitizer returned invalid json")
)

type NotFoundError struct {
	TraceID string
}

func (e *NotFoundError) Error() string { return fmt.Sprintf("diagnostics: no events for trace id %q", e.TraceID) }
func (e *NotFoundError) Unwrap() error { return ErrNoEvents }

type EventQuerier interface {
	Query(context.Context, storage.Query) (storage.QueryResult, error)
}

type SanitizationSummary struct {
	RedactedFields uint64 `json:"redacted_fields"`
	DroppedFields  uint64 `json:"dropped_fields"`
}

type PayloadSanitizer func(json.RawMessage) (json.RawMessage, SanitizationSummary, error)
type UUIDGenerator func() (string, error)
type Clock func() time.Time

type Builder struct {
	Store        EventQuerier
	DefaultLimit int
	MaximumLimit int
	NewUUID      UUIDGenerator
	Now          Clock
	Sanitize     PayloadSanitizer
}

type Bundle struct {
	Version            string              `json:"version"`
	DiagnosticReportID string              `json:"diagnostic_report_id"`
	CreatedAt          time.Time           `json:"created_at"`
	TraceID            string              `json:"trace_id"`
	Events             []Event             `json:"events"`
	TotalEventCount    uint64              `json:"total_event_count"`
	IncludedEventCount int                 `json:"included_event_count"`
	Limited            bool                `json:"limited"`
	EventTimeRange     TimeRange           `json:"event_time_range"`
	Sanitization       SanitizationSummary `json:"sanitization"`
	Services           []string            `json:"services"`
	ServiceInstances   []string            `json:"service_instances"`
	Environments       []string            `json:"environments"`
	Builds             []string            `json:"builds"`
	RequestIDs         []string            `json:"request_ids"`
	SessionIDs         []string            `json:"session_ids"`
	RoomIDs            []string            `json:"room_ids"`
	MatchIDs           []string            `json:"match_ids"`
	PlayerIDs          []string            `json:"player_ids"`
	AccountIDs         []string            `json:"account_ids"`
}

type Event struct {
	EventID   string          `json:"event_id"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type summarySets struct {
	services, instances, environments, builds map[string]struct{}
	requests, sessions, rooms, matches        map[string]struct{}
	players, accounts                         map[string]struct{}
}

func newSummarySets() summarySets {
	return summarySets{
		services: make(map[string]struct{}), instances: make(map[string]struct{}),
		environments: make(map[string]struct{}), builds: make(map[string]struct{}),
		requests: make(map[string]struct{}), sessions: make(map[string]struct{}),
		rooms: make(map[string]struct{}), matches: make(map[string]struct{}),
		players: make(map[string]struct{}), accounts: make(map[string]struct{}),
	}
}

func (b Builder) Build(ctx context.Context, traceID string, limit int) (Bundle, error) {
	if !validUUID(traceID) {
		return Bundle{}, ErrInvalidTraceID
	}
	limit, err := b.eventLimit(limit)
	if err != nil {
		return Bundle{}, err
	}
	if b.Store == nil {
		return Bundle{}, errors.New("diagnostics: nil event querier")
	}
	if b.Sanitize == nil {
		return Bundle{}, ErrSanitizerRequired
	}

	result, err := b.Store.Query(ctx, storage.Query{TraceID: traceID, Limit: limit})
	if err != nil {
		return Bundle{}, err
	}
	if len(result.Records) == 0 {
		return Bundle{}, &NotFoundError{TraceID: traceID}
	}

	newUUID := b.NewUUID
	if newUUID == nil {
		newUUID = newUUIDv4
	}
	reportID, err := newUUID()
	if err != nil {
		return Bundle{}, err
	}
	if !validUUID(reportID) {
		return Bundle{}, fmt.Errorf("%w: %q", ErrInvalidReportID, reportID)
	}
	now := b.Now
	if now == nil {
		now = time.Now
	}

	bundle := Bundle{
		Version: BundleVersion, DiagnosticReportID: reportID, CreatedAt: now().UTC(),
		TraceID: traceID, TotalEventCount: result.Total,
		IncludedEventCount: len(result.Records), Limited: result.Limited,
	}
	sets := newSummarySets()
	for _, record := range result.Records {
		payload, summary, err := b.Sanitize(copyPayload(record.Payload))
		if err != nil {
			return Bundle{}, err
		}
		if len(payload) == 0 || !json.Valid(payload) || !isJSONObject(payload) {
			return Bundle{}, ErrInvalidSanitizedPayload
		}
		bundle.Sanitization.RedactedFields += summary.RedactedFields
		bundle.Sanitization.DroppedFields += summary.DroppedFields
		bundle.Events = append(bundle.Events, Event{
			EventID: record.EventID, Timestamp: record.Timestamp,
			Payload: copyPayload(payload),
		})
		if bundle.EventTimeRange.From.IsZero() || record.Timestamp.Before(bundle.EventTimeRange.From) {
			bundle.EventTimeRange.From = record.Timestamp
		}
		if record.Timestamp.After(bundle.EventTimeRange.To) {
			bundle.EventTimeRange.To = record.Timestamp
		}
		addSummaries(&sets, record)
	}
	assignSummaries(&bundle, sets)
	return bundle, nil
}

func (b Builder) eventLimit(limit int) (int, error) {
	if limit == 0 {
		limit = b.DefaultLimit
	}
	if limit == 0 {
		limit = DefaultEventLimit
	}
	maximum := b.MaximumLimit
	if maximum == 0 {
		maximum = MaximumEventLimit
	}
	if limit < 1 || limit > maximum {
		return 0, ErrInvalidLimit
	}
	return limit, nil
}

func addSummaries(sets *summarySets, record storage.Record) {
	values := []struct{ set map[string]struct{}; value string }{
		{sets.services, record.Service}, {sets.instances, record.ServiceInstanceID},
		{sets.environments, record.Environment}, {sets.builds, record.BuildVersion},
		{sets.requests, record.RequestID}, {sets.sessions, record.SessionID},
		{sets.rooms, record.RoomID}, {sets.matches, record.MatchID},
		{sets.players, record.PlayerID}, {sets.accounts, record.AccountID},
	}
	for _, item := range values {
		if item.value != "" {
			item.set[item.value] = struct{}{}
		}
	}
}

func assignSummaries(bundle *Bundle, sets summarySets) {
	bundle.Services = sortedValues(sets.services)
	bundle.ServiceInstances = sortedValues(sets.instances)
	bundle.Environments = sortedValues(sets.environments)
	bundle.Builds = sortedValues(sets.builds)
	bundle.RequestIDs = sortedValues(sets.requests)
	bundle.SessionIDs = sortedValues(sets.sessions)
	bundle.RoomIDs = sortedValues(sets.rooms)
	bundle.MatchIDs = sortedValues(sets.matches)
	bundle.PlayerIDs = sortedValues(sets.players)
	bundle.AccountIDs = sortedValues(sets.accounts)
}

func sortedValues(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func copyPayload(payload []byte) json.RawMessage {
	return json.RawMessage(append([]byte(nil), payload...))
}

func isJSONObject(payload json.RawMessage) bool {
	var object map[string]json.RawMessage
	return json.Unmarshal(payload, &object) == nil && object != nil
}

func validUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return false
	}
	for index, character := range value {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			continue
		}
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f') || (character >= 'A' && character <= 'F')) {
			return false
		}
	}
	return true
}

func newUUIDv4() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	raw := hex.EncodeToString(bytes[:])
	return raw[:8] + "-" + raw[8:12] + "-" + raw[12:16] + "-" + raw[16:20] + "-" + raw[20:], nil
}
