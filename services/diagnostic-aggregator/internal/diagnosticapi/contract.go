package diagnosticapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/diagnostics"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

const (
	DiagnosticReportsPath          = "/v1/diagnostic-reports"
	DiagnosticReportItemPathPrefix = DiagnosticReportsPath + "/"
	DiagnosticReportItemRoute      = DiagnosticReportItemPathPrefix + "{diagnostic_report_id}"
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

type DiagnosticReportCreateRequest = diagnostics.DiagnosticSubmission
type DiagnosticReport = diagnostics.DiagnosticReport
type SourceContext = diagnostics.DiagnosticSourceContext
type CorrelationContext = diagnostics.DiagnosticCorrelationContext
type FailureContext = diagnostics.DiagnosticFailureContext
type ReportSummary = diagnostics.DiagnosticReportSummary
type EventTimeRange = diagnostics.TimeRange

type ReportService interface {
	Create(context.Context, DiagnosticReportCreateRequest) (DiagnosticReport, error)
	Get(context.Context, string) (DiagnosticReport, error)
}

type RequestAuthorizer func(*http.Request) bool

type EventEmitter interface {
	Emit(observability.Request) observability.Result
}

type UUIDGenerator func() string

type HandlerConfig struct {
	MaxRequestBytes int64
	Authorize       RequestAuthorizer
	Emitter         EventEmitter
	NewUUID         UUIDGenerator
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}
