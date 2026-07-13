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
	return &handler{service: service, config: config}, nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
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
		writeError(w, http.StatusBadRequest, ErrorCodeInvalidReportID)
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

func (h *handler) authorized(w http.ResponseWriter, r *http.Request) bool {
	if !h.config.Authorize(r) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		writeError(w, http.StatusUnauthorized, ErrorCodeUnauthorized)
		return false
	}
	return true
}

func (h *handler) post(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r) {
		return
	}
	if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, ErrorCodeUnsupportedMediaType)
		return
	}
	if r.ContentLength > h.config.MaxRequestBytes {
		writeError(w, http.StatusRequestEntityTooLarge, ErrorCodeRequestTooLarge)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, h.config.MaxRequestBytes+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrorCodeMalformedJSON)
		return
	}
	if int64(len(body)) > h.config.MaxRequestBytes {
		writeError(w, http.StatusRequestEntityTooLarge, ErrorCodeRequestTooLarge)
		return
	}
	var request DiagnosticReportCreateRequest
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, ErrorCodeMalformedJSON)
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		writeError(w, http.StatusBadRequest, ErrorCodeTrailingJSON)
		return
	}
	report, err := h.service.Create(r.Context(), request)
	if err != nil {
		writeServiceError(w, err)
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
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, diagnosticreports.ErrInvalid):
		writeError(w, http.StatusUnprocessableEntity, ErrorCodeInvalidDiagnosticReport)
	case errors.Is(err, diagnosticreports.ErrRejected):
		writeError(w, http.StatusUnprocessableEntity, ErrorCodeDiagnosticReportRejected)
	case errors.Is(err, diagnosticreports.ErrInvalidID):
		writeError(w, http.StatusBadRequest, ErrorCodeInvalidReportID)
	case errors.Is(err, diagnosticreports.ErrNotFound):
		writeError(w, http.StatusNotFound, ErrorCodeDiagnosticReportNotFound)
	case errors.Is(err, diagnosticreports.ErrUnavailable):
		writeError(w, http.StatusServiceUnavailable, ErrorCodeServiceUnavailable)
	case errors.Is(err, ErrInvalidReport):
		writeError(w, http.StatusUnprocessableEntity, ErrorCodeInvalidDiagnosticReport)
	case errors.Is(err, ErrRejectedReport):
		writeError(w, http.StatusUnprocessableEntity, ErrorCodeDiagnosticReportRejected)
	case errors.Is(err, ErrInvalidReportID):
		writeError(w, http.StatusBadRequest, ErrorCodeInvalidReportID)
	case errors.Is(err, ErrReportNotFound):
		writeError(w, http.StatusNotFound, ErrorCodeDiagnosticReportNotFound)
	case errors.Is(err, ErrUnavailable):
		writeError(w, http.StatusServiceUnavailable, ErrorCodeServiceUnavailable)
	default:
		writeError(w, http.StatusServiceUnavailable, ErrorCodeServiceUnavailable)
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
