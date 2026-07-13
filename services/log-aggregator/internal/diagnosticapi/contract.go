package diagnosticapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/diagnostics"
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

type DiagnosticReportCreateRequest = diagnostics.DiagnosticSubmission
type DiagnosticReport = diagnostics.DiagnosticReport

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
