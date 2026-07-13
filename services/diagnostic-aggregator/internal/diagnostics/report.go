package diagnostics

import (
	"encoding/json"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/events"
)

// DiagnosticTrigger identifies why a diagnostic submission was created.
type DiagnosticTrigger string

const (
	DiagnosticTriggerManualBugReport       DiagnosticTrigger = "manual_bug_report"
	DiagnosticTriggerDevelopmentSubmission DiagnosticTrigger = "development_submission"
	DiagnosticTriggerCrash                 DiagnosticTrigger = "crash"
	DiagnosticTriggerStartupFailure        DiagnosticTrigger = "startup_failure"
	DiagnosticTriggerUnrecoverableState    DiagnosticTrigger = "unrecoverable_state"
	DiagnosticTriggerRecoveryExhausted     DiagnosticTrigger = "recovery_exhausted"
)

// Valid reports whether the trigger is one of the supported contract values.
func (trigger DiagnosticTrigger) Valid() bool {
	switch trigger {
	case DiagnosticTriggerManualBugReport,
		DiagnosticTriggerDevelopmentSubmission,
		DiagnosticTriggerCrash,
		DiagnosticTriggerStartupFailure,
		DiagnosticTriggerUnrecoverableState,
		DiagnosticTriggerRecoveryExhausted:
		return true
	default:
		return false
	}
}

const DiagnosticReportVersion = 1

type DiagnosticSourceContext struct {
	Service           string `json:"service"`
	ServiceInstanceID string `json:"service_instance_id"`
	Environment       string `json:"environment"`
	BuildVersion      string `json:"build_version"`
	Platform          string `json:"platform"`
}

type DiagnosticCorrelationContext struct {
	TraceID   string `json:"trace_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	RoomID    string `json:"room_id,omitempty"`
	MatchID   string `json:"match_id,omitempty"`
	PlayerID  string `json:"player_id,omitempty"`
	AccountID string `json:"account_id,omitempty"`
}

type DiagnosticFailureContext struct {
	FailureMode string `json:"failure_mode"`
	ErrorCode   string `json:"error_code,omitempty"`
	Component   string `json:"component,omitempty"`
	Message     string `json:"message,omitempty"`
}

// DiagnosticSubmission is the untrusted intake contract. Events remain raw
// until each event is strictly decoded by the validation pipeline.
type DiagnosticSubmission struct {
	ReportVersion   int                          `json:"report_version"`
	Trigger         DiagnosticTrigger            `json:"trigger"`
	SubmittedAt     time.Time                    `json:"submitted_at"`
	Source          DiagnosticSourceContext      `json:"source"`
	Correlation     DiagnosticCorrelationContext `json:"correlation"`
	UserDescription string                       `json:"user_description,omitempty"`
	Failure         *DiagnosticFailureContext    `json:"failure,omitempty"`
	Events          []json.RawMessage            `json:"events"`
}

type DiagnosticReport struct {
	DiagnosticReportID string                       `json:"diagnostic_report_id"`
	ReportVersion      int                          `json:"report_version"`
	Trigger            DiagnosticTrigger            `json:"trigger"`
	CreatedAt          time.Time                    `json:"created_at"`
	SubmittedAt        time.Time                    `json:"submitted_at"`
	Source             DiagnosticSourceContext      `json:"source"`
	Correlation        DiagnosticCorrelationContext `json:"correlation"`
	UserDescription    string                       `json:"user_description,omitempty"`
	Failure            *DiagnosticFailureContext    `json:"failure,omitempty"`
	Events             []events.Event               `json:"events"`
	Summary            DiagnosticReportSummary      `json:"summary"`
}

type DiagnosticReportSummary struct {
	SubmittedEventCount uint64    `json:"submitted_event_count"`
	AcceptedEventCount  uint64    `json:"accepted_event_count"`
	RejectedEventCount  uint64    `json:"rejected_event_count"`
	RedactedFieldCount  uint64    `json:"redacted_field_count"`
	DroppedFieldCount   uint64    `json:"dropped_field_count"`
	Truncated           bool      `json:"truncated"`
	TruncatedEventCount uint64    `json:"truncated_event_count"`
	Services            []string  `json:"services"`
	ServiceInstances    []string  `json:"service_instances"`
	Environments        []string  `json:"environments"`
	Builds              []string  `json:"builds"`
	EventTimeRange      TimeRange `json:"event_time_range"`
}
