package ingest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

const EventsPath = "/v1/events"

type BatchRequest struct {
	Events []json.RawMessage `json:"events"`
}

type EventRejection struct {
	Index   int    `json:"index"`
	EventID string `json:"event_id,omitempty"`
	Code    string `json:"code"`
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

func NewHandler(ingestor BatchIngestor) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"code": "method_not_allowed"})
			return
		}

		var request BatchRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"code": "malformed_json"})
			return
		}
		var extra any
		if err := decoder.Decode(&extra); err == nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"code": "trailing_json"})
			return
		} else if err != io.EOF {
			writeJSON(w, http.StatusBadRequest, map[string]string{"code": "trailing_json"})
			return
		}
		if len(request.Events) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"code": "empty_events"})
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
