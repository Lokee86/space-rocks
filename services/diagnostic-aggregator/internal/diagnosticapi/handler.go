package diagnosticapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/diagnosticreports"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/operationcontext"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/google/uuid"
)

type handler struct {
	service ReportService
	config  HandlerConfig
}

func NewHandler(service ReportService, config HandlerConfig) (http.Handler, error) {
	if service == nil {
		return nil, errors.New("diagnosticapi: nil report service")
	}
	if config.MaxRequestBytes <= 0 {
		return nil, errors.New("diagnosticapi: max request bytes must be positive")
	}
	if config.Authorize == nil {
		return nil, errors.New("diagnosticapi: authorizer is required")
	}
	if config.Emitter == nil {
		return nil, errors.New("diagnosticapi: emitter is required")
	}
	if config.NewUUID == nil {
		config.NewUUID = uuid.NewString
	}
	return &handler{service: service, config: config}, nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	operation := operationcontext.Values{TraceID: h.config.NewUUID(), RequestID: h.config.NewUUID(), Route: routeForPath(r.URL.Path)}
	r = r.WithContext(operationcontext.With(r.Context(), operation))
	if r.URL.Path == DiagnosticReportsPath {
		if r.Method != http.MethodPost {
			h.methodError(w, http.MethodPost)
			return
		}
		h.post(w, r)
		return
	}
	if id, ok := reportIDFromPath(r.URL.Path); ok {
		if r.Method != http.MethodGet {
			h.methodError(w, http.MethodGet)
			return
		}
		h.get(w, r, id)
		return
	}
	if strings.HasPrefix(r.URL.Path, DiagnosticReportItemPathPrefix) {
		h.reject(w, r, http.StatusBadRequest, ErrorCodeInvalidReportID, "route_validation", nil)
		return
	}
	writeError(w, http.StatusNotFound, ErrorCodeNotFound)
}

func reportIDFromPath(path string) (string, bool) {
	if !strings.HasPrefix(path, DiagnosticReportItemPathPrefix) {
		return "", false
	}
	id := strings.TrimPrefix(path, DiagnosticReportItemPathPrefix)
	if id == "" || strings.Contains(id, "/") {
		return "", false
	}
	return id, true
}

func routeForPath(path string) string {
	switch {
	case path == DiagnosticReportsPath:
		return DiagnosticReportsPath
	case strings.HasPrefix(path, DiagnosticReportItemPathPrefix):
		return DiagnosticReportItemRoute
	default:
		return "unmatched"
	}
}

func (h *handler) authorized(w http.ResponseWriter, r *http.Request) bool {
	if !h.config.Authorize(r) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		h.reject(w, r, http.StatusUnauthorized, ErrorCodeUnauthorized, "authorization", nil)
		return false
	}
	return true
}

func (h *handler) post(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r) {
		return
	}
	if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
		h.reject(w, r, http.StatusUnsupportedMediaType, ErrorCodeUnsupportedMediaType, "content_type_validation", nil)
		return
	}
	if r.ContentLength > h.config.MaxRequestBytes {
		size := r.ContentLength
		h.reject(w, r, http.StatusRequestEntityTooLarge, ErrorCodeRequestTooLarge, "request_size_validation", &size)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, h.config.MaxRequestBytes+1))
	if err != nil {
		h.reject(w, r, http.StatusBadRequest, ErrorCodeMalformedJSON, "body_read", nil)
		return
	}
	if int64(len(body)) > h.config.MaxRequestBytes {
		size := int64(len(body))
		h.reject(w, r, http.StatusRequestEntityTooLarge, ErrorCodeRequestTooLarge, "body_read", &size)
		return
	}
	var request DiagnosticReportCreateRequest
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		size := int64(len(body))
		h.reject(w, r, http.StatusBadRequest, ErrorCodeMalformedJSON, "report_decode", &size)
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		size := int64(len(body))
		h.reject(w, r, http.StatusBadRequest, ErrorCodeTrailingJSON, "report_decode", &size)
		return
	}
	if request.Correlation.TraceID != "" && uuid.Validate(request.Correlation.TraceID) == nil {
		operation, _ := operationcontext.From(r.Context())
		operation.TraceID = request.Correlation.TraceID
		r = r.WithContext(operationcontext.With(r.Context(), operation))
	}
	report, err := h.service.Create(r.Context(), request)
	if err != nil {
		status, code := serviceError(err)
		switch code {
		case ErrorCodeInvalidDiagnosticReport:
			h.reject(w, r, status, code, "report_validation", nil)
		case ErrorCodeDiagnosticReportRejected:
			h.reject(w, r, status, code, "safety_inspection", nil)
		default:
			writeError(w, status, code)
		}
		return
	}
	w.Header().Set("Location", DiagnosticReportItemPathPrefix+report.DiagnosticReportID)
	writeJSON(w, http.StatusCreated, report)
}

func (h *handler) get(w http.ResponseWriter, r *http.Request, id string) {
	if !h.authorized(w, r) {
		return
	}
	report, err := h.service.Get(r.Context(), id)
	if err != nil {
		status, code := serviceError(err)
		if code == ErrorCodeInvalidReportID {
			h.reject(w, r, status, code, "report_identifier", nil)
		} else {
			writeError(w, status, code)
		}
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func serviceError(err error) (int, string) {
	switch {
	case errors.Is(err, diagnosticreports.ErrInvalid), errors.Is(err, ErrInvalidReport):
		return http.StatusUnprocessableEntity, ErrorCodeInvalidDiagnosticReport
	case errors.Is(err, diagnosticreports.ErrRejected), errors.Is(err, ErrRejectedReport):
		return http.StatusUnprocessableEntity, ErrorCodeDiagnosticReportRejected
	case errors.Is(err, diagnosticreports.ErrInvalidID), errors.Is(err, ErrInvalidReportID):
		return http.StatusBadRequest, ErrorCodeInvalidReportID
	case errors.Is(err, diagnosticreports.ErrNotFound), errors.Is(err, ErrReportNotFound):
		return http.StatusNotFound, ErrorCodeDiagnosticReportNotFound
	default:
		return http.StatusServiceUnavailable, ErrorCodeServiceUnavailable
	}
}

func (h *handler) reject(w http.ResponseWriter, r *http.Request, status int, code, stage string, bodySize *int64) {
	operation, _ := operationcontext.From(r.Context())
	fields := observability.Fields{
		"reason_code":   code,
		"failure_stage": stage,
		"http_method":   safeHTTPMethod(r.Method),
		"status_code":   status,
	}
	if bodySize != nil && *bodySize >= 0 {
		fields["body_size"] = *bodySize
	}
	h.config.Emitter.Emit(observability.Request{
		Event: observability.EventNameDiagnosticReportRejected,
		Context: observability.Context{
			TraceID: operation.TraceID, RequestID: operation.RequestID, Route: operation.Route,
		},
		Fields: fields,
	})
	writeError(w, status, code)
}

func safeHTTPMethod(method string) string {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions:
		return method
	default:
		return "other"
	}
}
func (h *handler) methodError(w http.ResponseWriter, allowed string) {
	w.Header().Set("Allow", allowed)
	writeError(w, http.StatusMethodNotAllowed, ErrorCodeMethodNotAllowed)
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, ErrorResponse{Code: code})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
