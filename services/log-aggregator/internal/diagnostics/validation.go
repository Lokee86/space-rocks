package diagnostics

import (
	"errors"
	"strings"
	"time"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/events"
)

// ValidationCode identifies a safe, stable diagnostic-submission validation failure.
type ValidationCode string

const (
	ValidationCodeEmbeddedEventsLimit     ValidationCode = "embedded_events_limit_must_be_positive"
	ValidationCodeUserDescriptionLimit    ValidationCode = "user_description_bytes_limit_must_be_positive"
	ValidationCodeFailureMessageLimit     ValidationCode = "failure_message_bytes_limit_must_be_positive"
	ValidationCodeContextStringLimit      ValidationCode = "context_string_bytes_limit_must_be_positive"
	ValidationCodeInvalidReportVersion    ValidationCode = "invalid_report_version"
	ValidationCodeUnsupportedTrigger      ValidationCode = "unsupported_trigger"
	ValidationCodeMissingSubmittedAt      ValidationCode = "missing_submitted_at"
	ValidationCodeMissingSourceField      ValidationCode = "missing_source_field"
	ValidationCodeInvalidCorrelationID    ValidationCode = "invalid_correlation_id"
	ValidationCodeEmbeddedEventsExceeded  ValidationCode = "embedded_events_limit_exceeded"
	ValidationCodeUserDescriptionExceeded ValidationCode = "user_description_limit_exceeded"
	ValidationCodeFailureMessageExceeded  ValidationCode = "failure_message_limit_exceeded"
	ValidationCodeContextStringExceeded   ValidationCode = "context_string_limit_exceeded"
	ValidationCodeMissingFailureMode      ValidationCode = "missing_failure_mode"
	ValidationCodeEmptyEmbeddedEvent      ValidationCode = "empty_embedded_event"
	ValidationCodeInvalidEmbeddedEvent    ValidationCode = "invalid_embedded_event"
	ValidationCodeInvalidReportID         ValidationCode = "invalid_report_id"
	ValidationCodeInvalidReportEvent      ValidationCode = "invalid_report_event"
	ValidationCodeReportInvariant         ValidationCode = "report_invariant_violation"
)

// ValidationError reports a validation code without retaining or exposing submitted values.
type ValidationError struct {
	code ValidationCode
}

func (e *ValidationError) Error() string        { return string(e.code) }
func (e *ValidationError) Code() ValidationCode { return e.code }

// ValidationCodeOf extracts a stable validation code from an error.
func ValidationCodeOf(err error) (ValidationCode, bool) {
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) || validationErr == nil {
		return "", false
	}
	return validationErr.Code(), true
}

// SubmissionLimits bounds the user-controlled portions of a diagnostic submission.
type SubmissionLimits struct {
	MaxEmbeddedEvents       int
	MaxUserDescriptionBytes int
	MaxFailureMessageBytes  int
	MaxContextStringBytes   int
}

func (l SubmissionLimits) Validate() error {
	limits := []struct {
		value int
		code  ValidationCode
	}{
		{l.MaxEmbeddedEvents, ValidationCodeEmbeddedEventsLimit},
		{l.MaxUserDescriptionBytes, ValidationCodeUserDescriptionLimit},
		{l.MaxFailureMessageBytes, ValidationCodeFailureMessageLimit},
		{l.MaxContextStringBytes, ValidationCodeContextStringLimit},
	}
	for _, limit := range limits {
		if limit.value <= 0 {
			return &ValidationError{code: limit.code}
		}
	}
	return nil
}

// ValidateSubmissionEnvelope validates the untrusted diagnostic submission envelope.
func ValidateSubmissionEnvelope(submission DiagnosticSubmission, limits SubmissionLimits) error {
	if err := limits.Validate(); err != nil {
		return err
	}
	if submission.ReportVersion != DiagnosticReportVersion {
		return validationError(ValidationCodeInvalidReportVersion)
	}
	if !submission.Trigger.Valid() {
		return validationError(ValidationCodeUnsupportedTrigger)
	}
	if submission.SubmittedAt.IsZero() {
		return validationError(ValidationCodeMissingSubmittedAt)
	}
	for _, value := range []string{submission.Source.Service, submission.Source.ServiceInstanceID, submission.Source.Environment, submission.Source.BuildVersion, submission.Source.Platform} {
		if strings.TrimSpace(value) == "" {
			return validationError(ValidationCodeMissingSourceField)
		}
		if len([]byte(value)) > limits.MaxContextStringBytes {
			return validationError(ValidationCodeContextStringExceeded)
		}
	}
	if !validUUID(submission.Source.ServiceInstanceID) {
		return validationError(ValidationCodeInvalidCorrelationID)
	}
	correlationIDs := []string{submission.Correlation.TraceID, submission.Correlation.RequestID, submission.Correlation.SessionID, submission.Correlation.RoomID, submission.Correlation.MatchID, submission.Correlation.PlayerID, submission.Correlation.AccountID}
	for _, value := range correlationIDs {
		if value != "" && !validUUID(value) {
			return validationError(ValidationCodeInvalidCorrelationID)
		}
		if len([]byte(value)) > limits.MaxContextStringBytes {
			return validationError(ValidationCodeContextStringExceeded)
		}
	}
	if len(submission.Events) > limits.MaxEmbeddedEvents {
		return validationError(ValidationCodeEmbeddedEventsExceeded)
	}
	if len([]byte(submission.UserDescription)) > limits.MaxUserDescriptionBytes {
		return validationError(ValidationCodeUserDescriptionExceeded)
	}
	if submission.Failure != nil {
		if strings.TrimSpace(submission.Failure.FailureMode) == "" {
			return validationError(ValidationCodeMissingFailureMode)
		}
		if len([]byte(submission.Failure.Message)) > limits.MaxFailureMessageBytes {
			return validationError(ValidationCodeFailureMessageExceeded)
		}
		for _, value := range []string{submission.Failure.FailureMode, submission.Failure.ErrorCode, submission.Failure.Component} {
			if len([]byte(value)) > limits.MaxContextStringBytes {
				return validationError(ValidationCodeContextStringExceeded)
			}
		}
	}
	for _, event := range submission.Events {
		if len(event) == 0 {
			return validationError(ValidationCodeEmptyEmbeddedEvent)
		}
	}
	return nil
}

func validationError(code ValidationCode) error { return &ValidationError{code: code} }

// DecodeSubmissionEvents strictly decodes and validates all embedded canonical events.
func DecodeSubmissionEvents(submission DiagnosticSubmission, limits SubmissionLimits) ([]events.Event, error) {
	if err := ValidateSubmissionEnvelope(submission, limits); err != nil {
		return nil, err
	}
	decoded := make([]events.Event, len(submission.Events))
	for index, raw := range submission.Events {
		event, err := events.Decode(raw)
		if err != nil {
			return nil, validationError(ValidationCodeInvalidEmbeddedEvent)
		}
		decoded[index] = event
	}
	return decoded, nil
}

// ValidateDiagnosticReport validates server-generated report invariants.
func ValidateDiagnosticReport(report DiagnosticReport, limits SubmissionLimits) error {
	if err := limits.Validate(); err != nil {
		return err
	}
	if !validUUID(report.DiagnosticReportID) {
		return validationError(ValidationCodeInvalidReportID)
	}
	if report.ReportVersion != DiagnosticReportVersion || !report.Trigger.Valid() || report.CreatedAt.IsZero() || report.SubmittedAt.IsZero() {
		return validationError(ValidationCodeReportInvariant)
	}
	for _, value := range []string{report.Source.Service, report.Source.ServiceInstanceID, report.Source.Environment, report.Source.BuildVersion, report.Source.Platform} {
		if strings.TrimSpace(value) == "" || len([]byte(value)) > limits.MaxContextStringBytes {
			return validationError(ValidationCodeReportInvariant)
		}
	}
	if !validUUID(report.Source.ServiceInstanceID) {
		return validationError(ValidationCodeReportInvariant)
	}
	for _, value := range []string{report.Correlation.TraceID, report.Correlation.RequestID, report.Correlation.SessionID, report.Correlation.RoomID, report.Correlation.MatchID, report.Correlation.PlayerID, report.Correlation.AccountID} {
		if value != "" && (!validUUID(value) || len([]byte(value)) > limits.MaxContextStringBytes) {
			return validationError(ValidationCodeReportInvariant)
		}
	}
	if len([]byte(report.UserDescription)) > limits.MaxUserDescriptionBytes {
		return validationError(ValidationCodeReportInvariant)
	}
	if report.Failure != nil {
		if strings.TrimSpace(report.Failure.FailureMode) == "" || len([]byte(report.Failure.Message)) > limits.MaxFailureMessageBytes {
			return validationError(ValidationCodeReportInvariant)
		}
		for _, value := range []string{report.Failure.FailureMode, report.Failure.ErrorCode, report.Failure.Component} {
			if len([]byte(value)) > limits.MaxContextStringBytes {
				return validationError(ValidationCodeReportInvariant)
			}
		}
	}
	if len(report.Events) > limits.MaxEmbeddedEvents {
		return validationError(ValidationCodeReportInvariant)
	}
	var earliest, latest time.Time
	for _, event := range report.Events {
		if err := events.Validate(event); err != nil {
			return validationError(ValidationCodeInvalidReportEvent)
		}
		timestamp, err := time.Parse(time.RFC3339, event.Timestamp)
		if err != nil {
			return validationError(ValidationCodeInvalidReportEvent)
		}
		if earliest.IsZero() || timestamp.Before(earliest) {
			earliest = timestamp
		}
		if latest.IsZero() || timestamp.After(latest) {
			latest = timestamp
		}
	}
	if report.Summary.AcceptedEventCount != uint64(len(report.Events)) || report.Summary.SubmittedEventCount != report.Summary.AcceptedEventCount+report.Summary.RejectedEventCount+report.Summary.TruncatedEventCount {
		return validationError(ValidationCodeReportInvariant)
	}
	if (!report.Summary.Truncated && report.Summary.TruncatedEventCount != 0) || (report.Summary.Truncated && report.Summary.TruncatedEventCount == 0) {
		return validationError(ValidationCodeReportInvariant)
	}
	if earliest.IsZero() {
		if !report.Summary.EventTimeRange.From.IsZero() || !report.Summary.EventTimeRange.To.IsZero() {
			return validationError(ValidationCodeReportInvariant)
		}
	} else if !report.Summary.EventTimeRange.From.Equal(earliest) || !report.Summary.EventTimeRange.To.Equal(latest) {
		return validationError(ValidationCodeReportInvariant)
	}
	return nil
}
