package auditprojection

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage"
)

const (
	eventID    = "123e4567-e89b-12d3-a456-426614174000"
	traceID    = "123e4567-e89b-12d3-a456-426614174001"
	instanceID = "123e4567-e89b-12d3-a456-426614174002"
	auditID    = "123e4567-e89b-12d3-a456-426614174003"
)

var sourceTime = time.Unix(1, 2).UTC()
var promotedTime = time.Unix(3, 4).In(time.FixedZone("test", 3600))

type fakeStore struct {
	record Record
	err    error
	saves  int
}

func (s *fakeStore) Save(_ context.Context, record Record) error {
	s.saves++
	s.record = cloneRecord(record)
	return s.err
}
func (s *fakeStore) Get(_ context.Context, id string) (Record, error) {
	if s.record.AuditEventID != id {
		return Record{}, &RecordNotFoundError{AuditEventID: id}
	}
	return cloneRecord(s.record), nil
}

func canonicalPayload() json.RawMessage {
	return json.RawMessage(`{
		"audit_required":true,"audit_type":"account_change",
		"event_id":"123e4567-e89b-12d3-a456-426614174000",
		"trace_id":"123e4567-e89b-12d3-a456-426614174001",
		"service_instance_id":"123e4567-e89b-12d3-a456-426614174002",
		"actor_id":"actor-1","actor_type":"user",
		"target_type":"account","target_id":"target-1","action":"update",
		"reason_code":"requested","case_id":"case-1","transaction_id":"txn-1",
		"result_id":"result-1","match_id":"match-1","account_id":"account-1"
	}`)
}

func sourceRecord(required bool) storage.Record {
	return storage.Record{
		EventID: eventID, TraceID: traceID, ServiceInstanceID: instanceID,
		Service: "game-server", Timestamp: sourceTime, AuditRequired: required,
		Payload: canonicalPayload(),
	}
}

func configuredPromoter(store Store) Promoter {
	return Promoter{
		Store:   store,
		NewUUID: func() (string, error) { return auditID, nil },
		Now:     func() time.Time { return promotedTime },
		Sanitize: func(payload json.RawMessage) (json.RawMessage, error) {
			return clonePayload(payload), nil
		},
	}
}

func TestPromoterNoOpHasNoSideEffects(t *testing.T) {
	store := &fakeStore{}
	calls := 0
	p := configuredPromoter(store)
	p.Sanitize = func(json.RawMessage) (json.RawMessage, error) { calls++; return nil, nil }
	p.ValidatePolicy = func(Record) error { calls++; return nil }
	p.NewUUID = func() (string, error) { calls++; return "", nil }
	p.Now = func() time.Time { calls++; return time.Time{} }
	result, err := p.Promote(context.Background(), storage.Record{Payload: json.RawMessage("bad")})
	if err != nil || result.Promoted || calls != 0 || store.saves != 0 {
		t.Fatalf("unexpected no-op: %#v %v calls=%d", result, err, calls)
	}
}

func TestPromoterSuccessfulPromotionMapsRecord(t *testing.T) {
	store := &fakeStore{}
	result, err := configuredPromoter(store).Promote(context.Background(), sourceRecord(true))
	if err != nil || !result.Promoted || store.saves != 1 {
		t.Fatalf("unexpected promotion: %#v %v", result, err)
	}
	if result.Record.Version != RecordVersion || result.Record.AuditEventID != auditID ||
		result.Record.SourceEventID != eventID || result.Record.TraceID != traceID ||
		result.Record.SourceService != "game-server" || result.Record.SourceServiceInstanceID != instanceID ||
		!result.Record.SourceTimestamp.Equal(sourceTime) || !result.Record.PromotedAt.Equal(promotedTime.UTC()) {
		t.Fatalf("incorrect record metadata: %#v", result.Record)
	}
	checks := map[string]string{
		"actor": result.Record.ActorID, "actor type": result.Record.ActorType,
		"target type": result.Record.TargetType, "target": result.Record.TargetID,
		"action": result.Record.Action, "reason": result.Record.ReasonCode,
		"case": result.Record.CaseID, "transaction": result.Record.TransactionID,
		"result": result.Record.ResultID, "match": result.Record.MatchID, "account": result.Record.AccountID,
	}
	for name, value := range checks {
		if value == "" {
			t.Errorf("missing %s", name)
		}
	}
}

func TestPromoterDefensiveOwnership(t *testing.T) {
	store := &fakeStore{}
	source := sourceRecord(true)
	original := append([]byte(nil), source.Payload...)
	var sanitized, policyPayload []byte
	p := configuredPromoter(store)
	p.Sanitize = func(payload json.RawMessage) (json.RawMessage, error) {
		sanitized = append([]byte(nil), payload...)
		return payload, nil
	}
	p.ValidatePolicy = func(record Record) error {
		policyPayload = append([]byte(nil), record.Payload...)
		record.Payload[0] = 'x'
		return nil
	}
	result, err := p.Promote(context.Background(), source)
	if err != nil {
		t.Fatal(err)
	}
	source.Payload[0] = 'x'
	sanitized[0] = 'x'
	policyPayload[0] = 'x'
	if string(store.record.Payload) != string(canonicalPayload()) || string(result.Record.Payload) != string(canonicalPayload()) {
		t.Fatal("payload ownership was not preserved")
	}
	if string(original) != string(canonicalPayload()) {
		t.Fatal("source payload changed")
	}
}

func TestPromoterSourceValidation(t *testing.T) {
	for name, mutate := range map[string]func(*storage.Record){
		"event":    func(r *storage.Record) { r.EventID = "bad" },
		"trace":    func(r *storage.Record) { r.TraceID = "bad" },
		"instance": func(r *storage.Record) { r.ServiceInstanceID = "bad" },
	} {
		t.Run(name, func(t *testing.T) {
			r := sourceRecord(true)
			mutate(&r)
			_, err := configuredPromoter(&fakeStore{}).Promote(context.Background(), r)
			if !errors.Is(err, ErrInvalidSourceID) {
				t.Fatalf("got %v", err)
			}
		})
	}
}
