package auditprojection

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestPromoterRequirements(t *testing.T) {
	p := configuredPromoter(nil)
	if _, err := p.Promote(context.Background(), sourceRecord(true)); !errors.Is(err, ErrStoreRequired) { t.Fatalf("store: %v", err) }
	p.Store = &fakeStore{}
	p.Sanitize = nil
	if _, err := p.Promote(context.Background(), sourceRecord(true)); !errors.Is(err, ErrSanitizerRequired) { t.Fatalf("sanitizer: %v", err) }
}

func TestPromoterOriginalPayloadValidation(t *testing.T) {
	cases := []struct {
		name string
		payload string
		err error
	}{
		{"malformed", `{`, ErrInvalidPayload},
		{"array", `[]`, ErrInvalidPayload},
		{"missing required", `{"audit_type":"account_change"}`, ErrProjectionMismatch},
		{"false required", `{"audit_required":false}`, ErrProjectionMismatch},
		{"missing event", `{"audit_required":true,"audit_type":"account_change"}`, ErrProjectionMismatch},
		{"mismatched event", `{"audit_required":true,"audit_type":"account_change","event_id":"bad","trace_id":"123e4567-e89b-12d3-a456-426614174001","service_instance_id":"123e4567-e89b-12d3-a456-426614174002"}`, ErrProjectionMismatch},
		{"missing trace", `{"audit_required":true,"audit_type":"account_change","event_id":"123e4567-e89b-12d3-a456-426614174000","service_instance_id":"123e4567-e89b-12d3-a456-426614174002"}`, ErrProjectionMismatch},
		{"missing instance", `{"audit_required":true,"audit_type":"account_change","event_id":"123e4567-e89b-12d3-a456-426614174000","trace_id":"123e4567-e89b-12d3-a456-426614174001"}`, ErrProjectionMismatch},
		{"missing type", `{"audit_required":true}`, ErrAuditTypeRequired},
		{"invalid type", `{"audit_required":true,"audit_type":"AccountChange"}`, ErrInvalidAuditType},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := sourceRecord(true); r.Payload = json.RawMessage(tc.payload)
			_, err := configuredPromoter(&fakeStore{}).Promote(context.Background(), r)
			if !errors.Is(err, tc.err) { t.Fatalf("got %v, want %v", err, tc.err) }
		})
	}
}

func TestPromoterSanitizedOutputValidation(t *testing.T) {
	outputs := []struct {
		name string
		output json.RawMessage
	}{
		{"empty", nil}, {"malformed", json.RawMessage(`{`)},
		{"array", json.RawMessage(`[]`)},
		{"removed required", json.RawMessage(`{"audit_required":true,"audit_type":"account_change"}`)},
		{"mismatched type", json.RawMessage(`{"audit_required":true,"audit_type":"Bad","event_id":"123e4567-e89b-12d3-a456-426614174000","trace_id":"123e4567-e89b-12d3-a456-426614174001","service_instance_id":"123e4567-e89b-12d3-a456-426614174002"}`)},
	}
	for _, tc := range outputs {
		t.Run(tc.name, func(t *testing.T) {
			p := configuredPromoter(&fakeStore{})
			p.Sanitize = func(json.RawMessage) (json.RawMessage, error) { return tc.output, nil }
			_, err := p.Promote(context.Background(), sourceRecord(true))
			if !errors.Is(err, ErrInvalidSanitizedOutput) { t.Fatalf("got %v", err) }
		})
	}
	want := errors.New("sanitize")
	p := configuredPromoter(&fakeStore{})
	p.Sanitize = func(json.RawMessage) (json.RawMessage, error) { return nil, want }
	if _, err := p.Promote(context.Background(), sourceRecord(true)); !errors.Is(err, want) { t.Fatalf("sanitizer error: %v", err) }
}

func TestPromoterPolicyAndGenerationFailures(t *testing.T) {
	policyErr := errors.New("reject")
	store := &fakeStore{}
	calls := 0
	p := configuredPromoter(store)
	p.ValidatePolicy = func(Record) error { return policyErr }
	p.NewUUID = func() (string, error) { calls++; return auditID, nil }
	if result, err := p.Promote(context.Background(), sourceRecord(true)); !errors.Is(err, policyErr) || result.Promoted || calls != 0 || store.saves != 0 { t.Fatalf("policy: %#v %v", result, err) }

	generatorErr := errors.New("uuid")
	p = configuredPromoter(&fakeStore{})
	p.NewUUID = func() (string, error) { return "", generatorErr }
	if result, err := p.Promote(context.Background(), sourceRecord(true)); !errors.Is(err, generatorErr) || result.Promoted { t.Fatalf("uuid error: %#v %v", result, err) }

	p = configuredPromoter(&fakeStore{})
	p.NewUUID = func() (string, error) { return "bad", nil }
	if result, err := p.Promote(context.Background(), sourceRecord(true)); !errors.Is(err, ErrInvalidGeneratedAuditID) || result.Promoted { t.Fatalf("uuid validation: %#v %v", result, err) }

	store = &fakeStore{err: errors.New("save")}
	p = configuredPromoter(store)
	if result, err := p.Promote(context.Background(), sourceRecord(true)); !errors.Is(err, store.err) || result.Promoted || store.saves != 1 { t.Fatalf("store error: %#v %v", result, err) }
}
