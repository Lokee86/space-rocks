package events

import (
	"regexp"
	"strings"
	"time"
)

const (
	CodeMalformedJSON        = "malformed_json"
	CodeTrailingJSON         = "trailing_json"
	CodeUnknownField         = "unknown_field"
	CodeRequiredFieldMissing = "required_field_missing"
	CodeUnsupportedSchema    = "unsupported_schema_version"
	CodeUnsupportedLevel     = "unsupported_level"
	CodeInvalidEventName     = "invalid_event_name"
	CodeInvalidTimestamp     = "invalid_timestamp"
	CodeInvalidUUID          = "invalid_uuid"
	CodeInvalidDuration      = "invalid_duration"
	CodeInvalidProcessID     = "invalid_process_id"
	CodeInvalidFieldName     = "invalid_field_name"
	CodeAuditTypeRequired    = "audit_type_required"
	CodeInvalidFields        = "invalid_fields"
)

type ValidationError struct{ Code string }

func (e *ValidationError) Error() string { return e.Code }

func validationError(code string) error { return &ValidationError{Code: code} }

var (
	eventNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	uuidPattern      = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
)

// Validate checks the shared observability envelope without including raw
// input values in returned errors.
func Validate(event Event) error {
	for _, value := range []string{event.EventID, event.Timestamp, event.Level, event.Event, event.Service, event.ServiceInstanceID} {
		if strings.TrimSpace(value) == "" {
			return validationError(CodeRequiredFieldMissing)
		}
	}
	if event.SchemaVersion != 1 {
		return validationError(CodeUnsupportedSchema)
	}
	if !isSupportedLevel(event.Level) {
		return validationError(CodeUnsupportedLevel)
	}
	if !eventNamePattern.MatchString(event.Event) {
		return validationError(CodeInvalidEventName)
	}
	if _, err := time.Parse(time.RFC3339, event.Timestamp); err != nil {
		return validationError(CodeInvalidTimestamp)
	}
	for _, value := range []string{
		event.EventID, event.ServiceInstanceID, event.TraceID, event.DiagnosticReportID,
		event.AuditEventID, event.BatchID, event.SourceEventID, event.SourceServiceInstanceID,
	} {
		if value != "" && !uuidPattern.MatchString(value) {
			return validationError(CodeInvalidUUID)
		}
	}
	if event.AuditRequired && strings.TrimSpace(event.AuditType) == "" {
		return validationError(CodeAuditTypeRequired)
	}
	if event.DurationMS != nil && *event.DurationMS < 0 {
		return validationError(CodeInvalidDuration)
	}
	if event.ProcessID != nil && *event.ProcessID < 0 {
		return validationError(CodeInvalidProcessID)
	}
	for _, value := range []string{event.AuditType, event.Action, event.ReasonCode} {
		if value != "" && !eventNamePattern.MatchString(value) {
			return validationError(CodeInvalidFieldName)
		}
	}
	if event.Fields != nil {
		for key := range event.Fields {
			if !eventNamePattern.MatchString(key) {
				return validationError(CodeInvalidFields)
			}
		}
	}
	return nil
}

func isSupportedLevel(level string) bool {
	switch level {
	case "debug", "info", "warn", "error", "critical", "fatal":
		return true
	default:
		return false
	}
}
