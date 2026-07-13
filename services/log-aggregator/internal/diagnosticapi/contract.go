package diagnosticapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

const (
	DiagnosticReportsPath          = "/v1/diagnostic-reports"
	DiagnosticReportItemPathPrefix = DiagnosticReportsPath + "/"
)

const (
	ErrorCodeMethodNotAllowed         = "method_not_allowed"
	ErrorCodeUnauthorized             = "unauthorized"
	ErrorCodeUnsupportedMediaType     = "unsupported_media_type"
	ErrorCodeMalformedJSON            = "malformed_json"
	ErrorCodeTrailingJSON             = "trailing_json"
	ErrorCodeRequestTooLarge          = "request_too_large"
	ErrorCodeInvalidDiagnosticReport  = "invalid_diagnostic_report"
	ErrorCodeDiagnosticReportRejected = "diagnostic_report_rejected"
	ErrorCodeInvalidReportID          = "invalid_diagnostic_report_id"
	ErrorCodeDiagnosticReportNotFound = "diagnostic_report_not_found"
	ErrorCodeNotFound                 = "not_found"
	ErrorCodeServiceUnavailable       = "service_unavailable"
)

var (
	ErrInvalidReport   = errors.New("diagnosticapi: invalid report")
	ErrRejectedReport  = errors.New("diagnosticapi: report rejected")
	ErrInvalidReportID = errors.New("diagnosticapi: invalid report id")
	ErrReportNotFound  = errors.New("diagnosticapi: report not found")
	ErrUnavailable     = errors.New("diagnosticapi: service unavailable")
)

type SourceContext struct {
	Service           string `json:"service"`
	ServiceInstanceID string `json:"service_instance_id"`
	Environment       string `json:"environment"`
	BuildVersion      string `json:"build_version"`
	Platform          string `json:"platform"`
}

type CorrelationContext struct {
	TraceID   string `json:"trace_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	RoomID    string `json:"room_id,omitempty"`
	MatchID   string `json:"match_id,omitempty"`
	PlayerID  string `json:"player_id,omitempty"`
	AccountID string `json:"account_id,omitempty"`
}

type FailureContext struct {
	FailureMode string `json:"failure_mode"`
	ErrorCode   string `json:"error_code,omitempty"`
	Component   string `json:"component,omitempty"`
	Message     string `json:"message,omitempty"`
}

type EventTimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type ReportSummary struct {
	SubmittedEventCount uint64         `json:"submitted_event_count"`
	AcceptedEventCount  uint64         `json:"accepted_event_count"`
	RejectedEventCount  uint64         `json:"rejected_event_count"`
	RedactedFieldCount  uint64         `json:"redacted_field_count"`
	DroppedFieldCount   uint64         `json:"dropped_field_count"`
	Truncated           bool           `json:"truncated"`
	TruncatedEventCount uint64         `json:"truncated_event_count"`
	Services            []string       `json:"services"`
	ServiceInstances    []string       `json:"service_instances"`
	Environments        []string       `json:"environments"`
	Builds              []string       `json:"builds"`
	EventTimeRange      EventTimeRange `json:"event_time_range"`
}

type DiagnosticReportCreateRequest struct {
	ReportVersion   int                `json:"report_version"`
	Trigger         string             `json:"trigger"`
	SubmittedAt     time.Time          `json:"submitted_at"`
	Source          SourceContext      `json:"source"`
	Correlation     CorrelationContext `json:"correlation"`
	UserDescription string             `json:"user_description,omitempty"`
	Failure         *FailureContext    `json:"failure,omitempty"`
	Events          []json.RawMessage  `json:"events"`
}

type DiagnosticReport struct {
	DiagnosticReportID string             `json:"diagnostic_report_id"`
	ReportVersion      int                `json:"report_version"`
	Trigger            string             `json:"trigger"`
	SubmittedAt        time.Time          `json:"submitted_at"`
	CreatedAt          time.Time          `json:"created_at"`
	Source             SourceContext      `json:"source"`
	Correlation        CorrelationContext `json:"correlation"`
	UserDescription    string             `json:"user_description,omitempty"`
	Failure            *FailureContext    `json:"failure,omitempty"`
	Events             []json.RawMessage  `json:"events"`
	Summary            ReportSummary      `json:"summary"`
}

type ReportService interface {
	Create(context.Context, DiagnosticReportCreateRequest) (DiagnosticReport, error)
	Get(context.Context, string) (DiagnosticReport, error)
}

type RequestAuthorizer func(*http.Request) bool

type HandlerConfig struct {
	MaxRequestBytes int64
	Authorize       RequestAuthorizer
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}
