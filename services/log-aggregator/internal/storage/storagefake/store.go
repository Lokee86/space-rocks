// Package storagefake provides a concurrency-safe in-memory EventStore for
// integration and consumer tests. It is not a production storage backend.
package storagefake

import (
	"context"
	"sort"
	"sync"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage"
)

// Failures contains optional errors returned by the corresponding fake
// operation. A nil error leaves that operation behaving normally.
type Failures struct {
	Append error
	Query  error
	Status error
}

// Store is an in-memory implementation of storage.EventStore.
type Store struct {
	mu       sync.RWMutex
	records  []storage.Record
	failures Failures
	closed   bool
}

var _ storage.EventStore = (*Store)(nil)

// New creates a Store seeded with optional records. Seed records are copied,
// including their payload bytes, and retain their input order.
func New(seeds ...storage.Record) *Store {
	return &Store{records: cloneRecords(seeds)}
}

// SetFailures configures errors injected by subsequent operations.
func (store *Store) SetFailures(failures Failures) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.failures = failures
}

// AppendBatch appends records in input order and copies payload bytes.
func (store *Store) AppendBatch(ctx context.Context, records []storage.Record) error {
	if err := contextError(ctx); err != nil {
		return err
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	if store.closed {
		return storage.ErrClosed
	}
	if store.failures.Append != nil {
		return store.failures.Append
	}
	store.records = append(store.records, cloneRecords(records)...)
	return nil
}

// Query returns matching records in deterministic chronological order.
func (store *Store) Query(ctx context.Context, query storage.Query) (storage.QueryResult, error) {
	if err := contextError(ctx); err != nil {
		return storage.QueryResult{}, err
	}

	store.mu.RLock()
	defer store.mu.RUnlock()
	if store.closed {
		return storage.QueryResult{}, storage.ErrClosed
	}
	if store.failures.Query != nil {
		return storage.QueryResult{}, store.failures.Query
	}

	matches := make([]storage.Record, 0, len(store.records))
	for _, record := range store.records {
		if matchesQuery(record, query) {
			matches = append(matches, cloneRecord(record))
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Timestamp.Equal(matches[j].Timestamp) {
			return matches[i].EventID < matches[j].EventID
		}
		return matches[i].Timestamp.Before(matches[j].Timestamp)
	})

	result := storage.QueryResult{Total: uint64(len(matches))}
	if query.Limit > 0 && query.Limit < len(matches) {
		result.Limited = true
		matches = matches[:query.Limit]
	}
	result.Records = matches
	return result, nil
}

// Status reports counts and event-time bounds without exposing backend details.
func (store *Store) Status(ctx context.Context) (storage.Status, error) {
	if err := contextError(ctx); err != nil {
		return storage.Status{}, err
	}

	store.mu.RLock()
	defer store.mu.RUnlock()
	if store.failures.Status != nil {
		return storage.Status{}, store.failures.Status
	}

	status := storage.Status{Ready: !store.closed, RecordCount: uint64(len(store.records))}
	for index, record := range store.records {
		status.ByteCount += uint64(len(record.Payload))
		if index == 0 || record.Timestamp.Before(status.Oldest) {
			status.Oldest = record.Timestamp
		}
		if index == 0 || record.Timestamp.After(status.Newest) {
			status.Newest = record.Timestamp
		}
	}
	return status, nil
}

// Close marks the store closed. Closing an already closed store succeeds.
func (store *Store) Close() error {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.closed = true
	return nil
}

// Snapshot returns a deep copy of all records in append order.
func (store *Store) Snapshot() []storage.Record {
	store.mu.RLock()
	defer store.mu.RUnlock()
	return cloneRecords(store.records)
}

func contextError(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func matchesQuery(record storage.Record, query storage.Query) bool {
	if !query.From.IsZero() && record.Timestamp.Before(query.From) ||
		!query.To.IsZero() && record.Timestamp.After(query.To) {
		return false
	}
	return equalOrEmpty(record.Service, query.Service) &&
		equalOrEmpty(record.ServiceInstanceID, query.ServiceInstanceID) &&
		equalOrEmpty(record.SchemaVersion, query.SchemaVersion) &&
		equalOrEmpty(record.Environment, query.Environment) &&
		equalOrEmpty(record.BuildVersion, query.BuildVersion) &&
		equalOrEmpty(record.Category, query.Category) &&
		equalOrEmpty(record.Level, query.Level) &&
		equalOrEmpty(record.Event, query.Event) &&
		equalOrEmpty(record.TraceID, query.TraceID) &&
		equalOrEmpty(record.RequestID, query.RequestID) &&
		equalOrEmpty(record.SessionID, query.SessionID) &&
		equalOrEmpty(record.RoomID, query.RoomID) &&
		equalOrEmpty(record.MatchID, query.MatchID) &&
		equalOrEmpty(record.PlayerID, query.PlayerID) &&
		equalOrEmpty(record.AccountID, query.AccountID) &&
		equalOrEmpty(record.EventID, query.EventID) &&
		equalOrEmpty(record.DiagnosticReportID, query.DiagnosticReportID) &&
		equalOrEmpty(record.AuditEventID, query.AuditEventID) &&
		equalOrEmpty(record.IdempotencyKey, query.IdempotencyKey) &&
		(query.AuditRequired == nil || record.AuditRequired == *query.AuditRequired)
}

func equalOrEmpty(value, filter string) bool {
	return filter == "" || value == filter
}

func cloneRecords(records []storage.Record) []storage.Record {
	clones := make([]storage.Record, len(records))
	for index, record := range records {
		clones[index] = cloneRecord(record)
	}
	return clones
}

func cloneRecord(record storage.Record) storage.Record {
	if record.Payload != nil {
		record.Payload = append([]byte(nil), record.Payload...)
	}
	return record
}
