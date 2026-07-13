package health

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/serviceidentity"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage"
)

type StoreStatusReader interface {
	Status(context.Context) (storage.Status, error)
}

type storageStatus struct {
	Ready       bool   `json:"ready"`
	Degraded    bool   `json:"degraded"`
	RecordCount uint64 `json:"record_count"`
	ByteCount   uint64 `json:"byte_count"`
}

type statusResponse struct {
	Service  string        `json:"service"`
	Snapshot Snapshot      `json:"snapshot"`
	Storage  storageStatus `json:"storage"`
}

type healthResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

func NewHandler(state *State, store StoreStatusReader) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/live", method(http.MethodGet, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{Service: serviceidentity.ServiceName, Status: "live"})
	}))
	mux.HandleFunc("/ready", method(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		if state == nil || store == nil || !state.Snapshot().Ready {
			writeJSON(w, http.StatusServiceUnavailable, healthResponse{Service: serviceidentity.ServiceName, Status: "not_ready"})
			return
		}
		status, err := store.Status(r.Context())
		if err != nil || !status.Ready {
			writeJSON(w, http.StatusServiceUnavailable, healthResponse{Service: serviceidentity.ServiceName, Status: "not_ready"})
			return
		}
		writeJSON(w, http.StatusOK, healthResponse{Service: serviceidentity.ServiceName, Status: "ready"})
	}))
	mux.HandleFunc("/status", method(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		response := statusResponse{Service: serviceidentity.ServiceName}
		if state != nil {
			response.Snapshot = state.Snapshot()
		}
		if store != nil {
			status, err := store.Status(r.Context())
			if err == nil {
				response.Storage = storageStatus{Ready: status.Ready, Degraded: status.Degraded, RecordCount: status.RecordCount, ByteCount: status.ByteCount}
			}
		}
		writeJSON(w, http.StatusOK, response)
	}))
	return mux
}

func method(want string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != want {
			w.Header().Set("Allow", want)
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"code": "method_not_allowed"})
			return
		}
		handler(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
