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
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/redaction"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage"
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

type Service struct {
	store   storage.ReportStore
	limits  diagnostics.SubmissionLimits
	policy  redaction.Policy
	now     Clock
	newUUID UUIDGenerator
}

func NewService(store storage.ReportStore, limits diagnostics.SubmissionLimits, policy redaction.Policy, now Clock, newUUID UUIDGenerator) (*Service, error) {
	if store == nil || now == nil || newUUID == nil {
		return nil, errors.New("diagnosticreports: dependencies are required")
	}
	if err := limits.Validate(); err != nil {
		return nil, err
	}
	return &Service{store: store, limits: limits, policy: policy, now: now, newUUID: newUUID}, nil
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
	raw, err := json.Marshal(report)
	if err != nil {
		return diagnostics.DiagnosticReport{}, ErrUnavailable
	}
	if err := s.store.Save(ctx, storage.Report{DiagnosticReportID: id, CreatedAt: created, RawJSON: raw}); err != nil {
		return diagnostics.DiagnosticReport{}, mapStorageError(err)
	}
	return report, nil
}

func (s *Service) Get(ctx context.Context, id string) (diagnostics.DiagnosticReport, error) {
	if uuid.Validate(id) != nil {
		return diagnostics.DiagnosticReport{}, ErrInvalidID
	}
	stored, err := s.store.Get(ctx, id)
	if err != nil {
		return diagnostics.DiagnosticReport{}, mapStorageError(err)
	}
	var report diagnostics.DiagnosticReport
	decoder := json.NewDecoder(bytes.NewReader(stored.RawJSON))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&report); err != nil || report.DiagnosticReportID != id {
		return diagnostics.DiagnosticReport{}, ErrUnavailable
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return diagnostics.DiagnosticReport{}, ErrUnavailable
	}
	if err := diagnostics.ValidateDiagnosticReport(report, s.limits); err != nil {
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
