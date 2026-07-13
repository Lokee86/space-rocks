package auditprojection

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage"
)

var (
	ErrStoreRequired           = errors.New("audit: store required")
	ErrSanitizerRequired       = errors.New("audit: sanitizer required")
	ErrInvalidSourceID         = errors.New("audit: invalid source id")
	ErrInvalidPayload          = errors.New("audit: invalid payload")
	ErrProjectionMismatch      = errors.New("audit: projection mismatch")
	ErrAuditTypeRequired       = errors.New("audit: audit type required")
	ErrInvalidAuditType        = errors.New("audit: invalid audit type")
	ErrInvalidSanitizedOutput  = errors.New("audit: invalid sanitized output")
	ErrInvalidGeneratedAuditID = errors.New("audit: invalid generated audit id")
)

type PayloadSanitizer func(json.RawMessage) (json.RawMessage, error)
type UUIDGenerator func() (string, error)
type Clock func() time.Time
type PolicyValidator func(Record) error

type PromotionResult struct {
	Promoted bool
	Record   Record
}

type Promoter struct {
	Store          Store
	NewUUID        UUIDGenerator
	Now            Clock
	Sanitize       PayloadSanitizer
	ValidatePolicy PolicyValidator
}

func (p Promoter) Promote(ctx context.Context, source storage.Record) (PromotionResult, error) {
	if !source.AuditRequired {
		return PromotionResult{Promoted: false}, nil
	}
	if p.Store == nil {
		return PromotionResult{}, ErrStoreRequired
	}
	if p.Sanitize == nil {
		return PromotionResult{}, ErrSanitizerRequired
	}
	if !validUUID(source.EventID) || !validUUID(source.TraceID) || !validUUID(source.ServiceInstanceID) {
		return PromotionResult{}, ErrInvalidSourceID
	}

	fields, err := objectFields(source.Payload)
	if err != nil {
		return PromotionResult{}, ErrInvalidPayload
	}
	auditType, err := validateProjection(fields, source)
	if err != nil {
		return PromotionResult{}, err
	}

	sanitized, err := p.Sanitize(clonePayload(source.Payload))
	if err != nil {
		return PromotionResult{}, err
	}
	sanitizedFields, err := objectFields(sanitized)
	if err != nil {
		return PromotionResult{}, ErrInvalidSanitizedOutput
	}
	auditType, err = validateProjection(sanitizedFields, source)
	if err != nil {
		return PromotionResult{}, ErrInvalidSanitizedOutput
	}

	record := Record{
		Version: RecordVersion, SourceEventID: source.EventID, SourceTimestamp: source.Timestamp,
		TraceID: source.TraceID, AuditType: auditType, SourceService: source.Service,
		SourceServiceInstanceID: source.ServiceInstanceID, Payload: clonePayload(sanitized),
		ActorID: stringField(sanitizedFields, "actor_id"), ActorType: stringField(sanitizedFields, "actor_type"),
		TargetType: stringField(sanitizedFields, "target_type"), TargetID: stringField(sanitizedFields, "target_id"),
		Action: stringField(sanitizedFields, "action"), ReasonCode: stringField(sanitizedFields, "reason_code"),
		CaseID: stringField(sanitizedFields, "case_id"), TransactionID: stringField(sanitizedFields, "transaction_id"),
		ResultID: stringField(sanitizedFields, "result_id"), MatchID: stringField(sanitizedFields, "match_id"),
		AccountID: stringField(sanitizedFields, "account_id"),
	}
	if p.ValidatePolicy != nil {
		if err := p.ValidatePolicy(cloneRecord(record)); err != nil {
			return PromotionResult{}, err
		}
	}
	generator := p.NewUUID
	if generator == nil {
		generator = newUUIDv4
	}
	auditID, err := generator()
	if err != nil {
		return PromotionResult{}, err
	}
	if !validUUID(auditID) {
		return PromotionResult{}, ErrInvalidGeneratedAuditID
	}
	clock := p.Now
	if clock == nil {
		clock = time.Now
	}
	record.AuditEventID = auditID
	record.PromotedAt = clock().UTC()
	if err := p.Store.Save(ctx, cloneRecord(record)); err != nil {
		return PromotionResult{}, err
	}
	return PromotionResult{Promoted: true, Record: cloneRecord(record)}, nil
}

func objectFields(payload json.RawMessage) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if len(payload) == 0 || json.Unmarshal(payload, &fields) != nil || fields == nil {
		return nil, ErrInvalidPayload
	}
	return fields, nil
}

func validateProjection(fields map[string]json.RawMessage, source storage.Record) (string, error) {
	var required bool
	if json.Unmarshal(fields["audit_required"], &required) != nil || !required {
		return "", ErrProjectionMismatch
	}
	var auditType string
	if json.Unmarshal(fields["audit_type"], &auditType) != nil || auditType == "" {
		return "", ErrAuditTypeRequired
	}
	if !snakeCase.MatchString(auditType) {
		return "", ErrInvalidAuditType
	}
	for key, expected := range map[string]string{"event_id": source.EventID, "trace_id": source.TraceID, "service_instance_id": source.ServiceInstanceID} {
		var actual string
		if json.Unmarshal(fields[key], &actual) != nil || actual != expected {
			return "", ErrProjectionMismatch
		}
	}
	return auditType, nil
}

func stringField(fields map[string]json.RawMessage, key string) string {
	var value string
	_ = json.Unmarshal(fields[key], &value)
	return value
}

func clonePayload(payload json.RawMessage) json.RawMessage {
	return append(json.RawMessage(nil), payload...)
}

var snakeCase = regexp.MustCompile(`^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$`)

func validUUID(value string) bool {
	if len(value) != 36 {
		return false
	}
	for i, c := range value {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
			continue
		}
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func newUUIDv4() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = b[6]&15 | 64
	b[8] = b[8]&63 | 128
	s := hex.EncodeToString(b[:])
	return s[:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:], nil
}
