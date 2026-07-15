package diagnosticreports

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/diagnostics"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/events"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/operationcontext"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/redaction"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/google/uuid"
)

var (
	ErrInvalid     = errors.New("diagnosticreports: invalid report")
	ErrRejected    = errors.New("diagnosticreports: report rejected")
	ErrInvalidID   = errors.New("diagnosticreports: invalid report id")
	ErrNotFound    = errors.New("diagnosticreports: report not found")
	ErrUnavailable = errors.New("diagnosticreports: service unavailable")
)

type Clock func() time.Time
type UUIDGenerator func() (string, error)

type EventEmitter interface {
	Emit(observability.Request) observability.Result
}

type Service struct {
	store   storage.ReportStore
	limits  diagnostics.SubmissionLimits
	policy  redaction.Policy
	now     Clock
	newUUID UUIDGenerator
	emitter EventEmitter
}

func NewService(store storage.ReportStore, limits diagnostics.SubmissionLimits, policy redaction.Policy, now Clock, newUUID UUIDGenerator, emitter EventEmitter) (*Service, error) {
	if store == nil || now == nil || newUUID == nil || emitter == nil {
		return nil, errors.New("diagnosticreports: dependencies are required")
	}
	if err := limits.Validate(); err != nil {
		return nil, err
	}
	return &Service{store: store, limits: limits, policy: policy, now: now, newUUID: newUUID, emitter: emitter}, nil
}

func eventContext(ctx context.Context, reportID string) observability.Context {
	operation, _ := operationcontext.From(ctx)
	return observability.Context{
		TraceID:            operation.TraceID,
		RequestID:          operation.RequestID,
		Route:              operation.Route,
		DiagnosticReportID: reportID,
	}
}

func storageFailureMode(err error) string {
	switch {
	case errors.Is(err, context.Canceled):
		return "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return "deadline_exceeded"
	default:
		return "unavailable"
	}
}

func (s *Service) emitStorageFailure(ctx context.Context, reportID, operation, stage string, err error) {
	s.emitter.Emit(observability.Request{
		Event:   observability.EventNameAggregatorStorageFailed,
		Context: eventContext(ctx, reportID),
		Fields: observability.Fields{
			"operation":     operation,
			"failure_stage": stage,
			"failure_mode":  storageFailureMode(err),
		},
	})
}

func (s *Service) emitStored(ctx context.Context, report diagnostics.DiagnosticReport, event observability.EventName) {
	s.emitter.Emit(observability.Request{
		Event: event, Context: eventContext(ctx, report.DiagnosticReportID),
		Fields: observability.Fields{"trigger": string(report.Trigger), "accepted_event_count": report.Summary.AcceptedEventCount, "source_service": report.Source.Service},
	})
}
func (s *Service) Create(ctx context.Context, submission diagnostics.DiagnosticSubmission) (diagnostics.DiagnosticReport, error) {
	if err := diagnostics.ValidateSubmissionEnvelope(submission, s.limits); err != nil {
		return diagnostics.DiagnosticReport{}, ErrInvalid
	}
	envelope, err := json.Marshal(submission)
	if err != nil {
		return diagnostics.DiagnosticReport{}, ErrInvalid
	}
	if findings, err := redaction.Inspect(envelope, s.policy); err != nil {
		return diagnostics.DiagnosticReport{}, ErrInvalid
	} else if len(findings) != 0 {
		return diagnostics.DiagnosticReport{}, ErrRejected
	}
	decoded, err := diagnostics.DecodeSubmissionEvents(submission, s.limits)
	if err != nil {
		return diagnostics.DiagnosticReport{}, ErrInvalid
	}
	id, err := s.newUUID()
	if err != nil || uuid.Validate(id) != nil {
		return diagnostics.DiagnosticReport{}, ErrUnavailable
	}
	created := s.now().UTC()
	report := diagnostics.DiagnosticReport{DiagnosticReportID: id, ReportVersion: submission.ReportVersion, Trigger: submission.Trigger, CreatedAt: created, SubmittedAt: submission.SubmittedAt, Source: submission.Source, Correlation: submission.Correlation, UserDescription: submission.UserDescription, Failure: submission.Failure, Events: decoded}
	report.Summary = summarize(decoded)
	if err := diagnostics.ValidateDiagnosticReport(report, s.limits); err != nil {
		return diagnostics.DiagnosticReport{}, ErrInvalid
	}
	s.emitter.Emit(observability.Request{
		Event: observability.EventNameAggregatorEventAccepted, Context: eventContext(ctx, id),
		Fields: observability.Fields{"submitted_event_count": len(submission.Events), "trigger": string(submission.Trigger), "source_service": submission.Source.Service},
	})
	raw, err := json.Marshal(report)
	if err != nil {
		return diagnostics.DiagnosticReport{}, ErrUnavailable
	}
	if err := s.store.Save(ctx, storage.Report{DiagnosticReportID: id, CreatedAt: created, RawJSON: raw}); err != nil {
		s.emitStorageFailure(ctx, id, "save", "report_save", err)
		return diagnostics.DiagnosticReport{}, mapStorageError(err)
	}
	s.emitStored(ctx, report, observability.EventNameDiagnosticReportStored)
	return report, nil
}

func (s *Service) Get(ctx context.Context, id string) (diagnostics.DiagnosticReport, error) {
	if uuid.Validate(id) != nil {
		return diagnostics.DiagnosticReport{}, ErrInvalidID
	}
	stored, err := s.store.Get(ctx, id)
	if err != nil {
		if !errors.Is(err, storage.ErrReportNotFound) {
			s.emitStorageFailure(ctx, id, "load", "report_load", err)
		}
		return diagnostics.DiagnosticReport{}, mapStorageError(err)
	}
	var report diagnostics.DiagnosticReport
	decoder := json.NewDecoder(bytes.NewReader(stored.RawJSON))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&report); err != nil || report.DiagnosticReportID != id {
		s.emitStorageFailure(ctx, id, "decode", "stored_report_decode", errors.New("stored report decode failed"))
		return diagnostics.DiagnosticReport{}, ErrUnavailable
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		s.emitStorageFailure(ctx, id, "decode", "stored_report_trailing_json", errors.New("stored report trailing JSON"))
		return diagnostics.DiagnosticReport{}, ErrUnavailable
	}
	if err := diagnostics.ValidateDiagnosticReport(report, s.limits); err != nil {
		s.emitStorageFailure(ctx, id, "validate", "stored_report_validation", errors.New("stored report validation failed"))
		return diagnostics.DiagnosticReport{}, ErrUnavailable
	}
	return report, nil
}

func mapStorageError(err error) error {
	if errors.Is(err, storage.ErrReportNotFound) {
		return ErrNotFound
	}
	return ErrUnavailable
}

func summarize(items []events.Event) diagnostics.DiagnosticReportSummary {
	sets := [4]map[string]struct{}{make(map[string]struct{}), make(map[string]struct{}), make(map[string]struct{}), make(map[string]struct{})}
	var from, to time.Time
	for _, event := range items {
		values := []string{event.Service, event.ServiceInstanceID, event.Environment, event.BuildVersion}
		for i, value := range values {
			if value != "" {
				sets[i][value] = struct{}{}
			}
		}
		t, _ := time.Parse(time.RFC3339, event.Timestamp)
		if from.IsZero() || t.Before(from) {
			from = t
		}
		if t.After(to) {
			to = t
		}
	}
	return diagnostics.DiagnosticReportSummary{SubmittedEventCount: uint64(len(items)), AcceptedEventCount: uint64(len(items)), Services: sorted(sets[0]), ServiceInstances: sorted(sets[1]), Environments: sorted(sets[2]), Builds: sorted(sets[3]), EventTimeRange: diagnostics.TimeRange{From: from, To: to}}
}
func sorted(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
