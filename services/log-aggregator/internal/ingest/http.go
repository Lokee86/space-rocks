package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

const EventsPath = "/v1/events"

const (
	defaultMaxRequestBytes = 1 << 20
	defaultMaxEvents       = 1000
)

type BatchRequest struct {
	Events []json.RawMessage `json:"events"`
}

type EventRejection struct {
	Index               int    `json:"index"`
	EventID             string `json:"event_id,omitempty"`
	Code                string `json:"code"`
	RuleID              string `json:"rule_id,omitempty"`
	NormalizedFieldPath string `json:"normalized_field_path,omitempty"`
	ReasonCode          string `json:"reason_code,omitempty"`
}

type BatchResult struct {
	BatchID  string
	Accepted int
	Rejected []EventRejection
}
type BatchResponse struct {
	BatchID    string           `json:"batch_id"`
	Accepted   int              `json:"accepted"`
	Rejected   int              `json:"rejected"`
	Rejections []EventRejection `json:"rejections,omitempty"`
}

type BatchIngestor interface {
	IngestBatch(context.Context, []json.RawMessage) (BatchResult, error)
}
type RequestAuthorizer func(*http.Request) bool

type HandlerConfig struct {
	MaxRequestBytes int
	MaxEvents       int
	Authorize       RequestAuthorizer
}

func NewHandler(ingestor BatchIngestor) http.Handler {
	return NewHandlerWithConfig(ingestor, HandlerConfig{MaxRequestBytes: defaultMaxRequestBytes, MaxEvents: defaultMaxEvents})
}

func NewHandlerWithConfig(ingestor BatchIngestor, config HandlerConfig) http.Handler {
	if config.MaxRequestBytes <= 0 {
		config.MaxRequestBytes = defaultMaxRequestBytes
	}
	if config.MaxEvents <= 0 {
		config.MaxEvents = defaultMaxEvents
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"code": "method_not_allowed"})
			return
		}
		if config.Authorize != nil && !config.Authorize(r) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"code": "unauthorized"})
			return
		}
		if r.ContentLength > int64(config.MaxRequestBytes) {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"code": "request_too_large"})
			return
		}

		limited := io.LimitReader(r.Body, int64(config.MaxRequestBytes)+1)
		body, err := io.ReadAll(limited)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"code": "malformed_json"})
			return
		}
		if len(body) > config.MaxRequestBytes {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"code": "request_too_large"})
			return
		}
		var request BatchRequest
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"code": "malformed_json"})
			return
		}
		var extra any
		if err := decoder.Decode(&extra); err == nil || err != io.EOF {
			writeJSON(w, http.StatusBadRequest, map[string]string{"code": "trailing_json"})
			return
		}
		if len(request.Events) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"code": "empty_events"})
			return
		}
		if len(request.Events) > config.MaxEvents {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"code": "batch_too_large"})
			return
		}
		if ingestor == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": "ingestion_unavailable"})
			return
		}
		result, err := ingestor.IngestBatch(r.Context(), request.Events)
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": "ingestion_unavailable"})
			return
		}
		writeJSON(w, http.StatusOK, BatchResponse{BatchID: result.BatchID, Accepted: result.Accepted, Rejected: len(result.Rejected), Rejections: result.Rejected})
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
