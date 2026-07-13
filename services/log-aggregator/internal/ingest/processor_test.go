package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/redaction"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage"
	"github.com/google/uuid"
)

type processorStore struct {
	records []storage.Record
	err     error
}

func (s *processorStore) AppendBatch(_ context.Context, records []storage.Record) error {
	s.records = append([]storage.Record(nil), records...)
	return s.err
}
func (s *processorStore) Query(context.Context, storage.Query) (storage.QueryResult, error) {
	return storage.QueryResult{}, nil
}
func (s *processorStore) Status(context.Context) (storage.Status, error) {
	return storage.Status{}, nil
}
func (s *processorStore) Close() error { return nil }

func processorEvent(id, name string) json.RawMessage {
	return json.RawMessage(`{"event_id":"` + id + `","timestamp":"2026-07-13T12:00:00Z","level":"info","event":"` + name + `","service":"log-aggregator","schema_version":1,"service_instance_id":"550e8400-e29b-41d4-a716-446655440001"}`)
}

func TestProcessorMixedBatchPreservesAcceptedOrder(t *testing.T) {
	store := &processorStore{}
	p := NewProcessor(store, redaction.DefaultPolicy(), func() time.Time { return time.Unix(10, 0) }, func() uuid.UUID { return uuid.MustParse("550e8400-e29b-41d4-a716-446655440099") }, 1000)
	raw := []json.RawMessage{processorEvent("550e8400-e29b-41d4-a716-446655440000", "first_event"), processorEvent("550e8400-e29b-41d4-a716-446655440002", "second_event"), json.RawMessage(`{"password":"secret"}`)}
	result, err := p.IngestBatch(context.Background(), raw)
	if err != nil || result.Accepted != 2 || len(result.Rejected) != 1 {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	if len(store.records) != 2 || store.records[0].Event != "first_event" || store.records[1].Event != "second_event" {
		t.Fatalf("records=%#v", store.records)
	}
	if result.BatchID != "550e8400-e29b-41d4-a716-446655440099" {
		t.Fatalf("batch id=%q", result.BatchID)
	}
	if result.Rejected[0].ReasonCode != "forbidden_credential_data" {
		t.Fatalf("rejection=%#v", result.Rejected[0])
	}
}

func TestProcessorSizeAndStorageFailures(t *testing.T) {
	store := &processorStore{err: errors.New("disk failure")}
	p := NewProcessor(store, redaction.DefaultPolicy(), time.Now, uuid.New, 10)
	result, err := p.IngestBatch(context.Background(), []json.RawMessage{processorEvent("550e8400-e29b-41d4-a716-446655440000", "too_large")})
	if err != nil || result.Accepted != 0 || result.Rejected[0].Code != CodeEventTooLarge {
		t.Fatalf("result=%#v err=%v", result, err)
	}

	p.maxBytes = 1000
	_, err = p.IngestBatch(context.Background(), []json.RawMessage{processorEvent("550e8400-e29b-41d4-a716-446655440000", "stored")})
	if !errors.Is(err, ErrStorageAppend) {
		t.Fatalf("error=%v", err)
	}
}

func TestProcessorNilStoreCancellationAndSafeRejectionID(t *testing.T) {
	event := processorEvent("550e8400-e29b-41d4-a716-446655440000", "valid_event")
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	p := NewProcessor(nil, redaction.DefaultPolicy(), time.Now, uuid.New, 1000)
	if _, err := p.IngestBatch(cancelled, []json.RawMessage{event}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error=%v", err)
	}
	if _, err := p.IngestBatch(context.Background(), []json.RawMessage{event}); !errors.Is(err, ErrStorageAppend) {
		t.Fatalf("nil store error=%v", err)
	}

	store := &processorStore{}
	p = NewProcessor(store, redaction.DefaultPolicy(), time.Now, uuid.New, 1000)
	result, err := p.IngestBatch(context.Background(), []json.RawMessage{
		json.RawMessage(`{"event_id":"not-valid","password":"secret"}`),
		json.RawMessage(`{"event_id":"550e8400-e29b-41d4-a716-446655440002","password":"secret"}`),
	})
	if err != nil || len(result.Rejected) != 2 || result.Rejected[0].EventID != "" || result.Rejected[1].EventID == "" {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}
