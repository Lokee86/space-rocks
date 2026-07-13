package diagnostics

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

const (
	BundlesPath            = "/v1/diagnostic-bundles"
	BundleItemPathPrefix   = BundlesPath + "/"
	DefaultMaxRequestBytes int64 = 4 << 10
)

type BundleService interface {
	Create(context.Context, string, int) (Bundle, error)
	Get(context.Context, string) (Bundle, error)
}

type bundleHTTPHandler struct {
	service         BundleService
	maxRequestBytes int64
}

func NewHTTPHandler(service BundleService, maxRequestBytes int64) http.Handler {
	if maxRequestBytes <= 0 {
		maxRequestBytes = DefaultMaxRequestBytes
	}
	return &bundleHTTPHandler{service: service, maxRequestBytes: maxRequestBytes}
}

func (h *bundleHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == BundlesPath:
		if r.Method != http.MethodPost {
			h.methodError(w, http.MethodPost)
			return
		}
		h.create(w, r)
	case strings.HasPrefix(r.URL.Path, BundleItemPathPrefix):
		if r.Method != http.MethodGet {
			h.methodError(w, http.MethodGet)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, BundleItemPathPrefix)
		if id == "" || strings.Contains(id, "/") {
			h.error(w, http.StatusBadRequest, "invalid_diagnostic_report_id")
			return
		}
		h.get(w, r, id)
	default:
		h.error(w, http.StatusNotFound, "not_found")
	}
}

type createRequest struct {
	TraceID string `json:"trace_id"`
	Limit   int    `json:"limit,omitempty"`
}

func (h *bundleHTTPHandler) create(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		h.error(w, http.StatusServiceUnavailable, "service_unavailable")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, h.maxRequestBytes+1))
	if err != nil {
		h.error(w, http.StatusBadRequest, "invalid_json")
		return
	}
	if int64(len(body)) > h.maxRequestBytes {
		h.error(w, http.StatusBadRequest, "request_too_large")
		return
	}
	var request createRequest
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if len(body) == 0 || decoder.Decode(&request) != nil {
		h.error(w, http.StatusBadRequest, "invalid_json")
		return
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		h.error(w, http.StatusBadRequest, "invalid_json")
		return
	}
	if request.TraceID == "" {
		h.error(w, http.StatusBadRequest, "invalid_trace_id")
		return
	}
	bundle, err := h.service.Create(r.Context(), request.TraceID, request.Limit)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	h.json(w, http.StatusCreated, bundle)
}

func (h *bundleHTTPHandler) get(w http.ResponseWriter, r *http.Request, id string) {
	if h.service == nil {
		h.error(w, http.StatusServiceUnavailable, "service_unavailable")
		return
	}
	bundle, err := h.service.Get(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	h.json(w, http.StatusOK, bundle)
}

func (h *bundleHTTPHandler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidTraceID):
		h.error(w, http.StatusBadRequest, "invalid_trace_id")
	case errors.Is(err, ErrInvalidReportID), errors.Is(err, ErrInvalidDiagnosticReportID):
		h.error(w, http.StatusBadRequest, "invalid_diagnostic_report_id")
	case errors.Is(err, ErrInvalidLimit):
		h.error(w, http.StatusBadRequest, "invalid_limit")
	case errors.Is(err, ErrNoEvents), errors.Is(err, ErrBundleNotFound):
		h.error(w, http.StatusNotFound, "not_found")
	default:
		h.error(w, http.StatusServiceUnavailable, "service_unavailable")
	}
}

func (h *bundleHTTPHandler) methodError(w http.ResponseWriter, allowed string) {
	w.Header().Set("Allow", allowed)
	h.error(w, http.StatusMethodNotAllowed, "method_not_allowed")
}

func (h *bundleHTTPHandler) error(w http.ResponseWriter, status int, code string) {
	h.json(w, status, map[string]string{"error": code})
}

func (h *bundleHTTPHandler) json(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
