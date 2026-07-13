package query

import (
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
)

const TracesPathPrefix = "/v1/traces/"

func NewTraceHandler(querier EventQuerier) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID, shaped := traceIDFromPath(r.URL.Path)
		if !shaped {
			writeJSON(w, http.StatusNotFound, map[string]string{"code": "not_found"})
			return
		}
		if !validUUID(traceID) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"code": "invalid_trace_id"})
			return
		}
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"code": "method_not_allowed"})
			return
		}
		if len(r.URL.Query()) > 0 {
			for name, values := range r.URL.Query() {
				if name != "limit" || len(values) != 1 {
					writeJSON(w, http.StatusBadRequest, map[string]string{"code": "invalid_filter"})
					return
				}
			}
		}
		filter := Filter{TraceID: traceID}
		if raw, present := r.URL.Query()["limit"]; present {
			limit, err := strconv.Atoi(raw[0])
			if err != nil || limit <= 0 {
				writeJSON(w, http.StatusBadRequest, map[string]string{"code": "invalid_limit"})
				return
			}
			filter.Limit = limit
		}
		result, err := querier.Query(r.Context(), filter)
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": "query_unavailable"})
			return
		}
		writeJSON(w, http.StatusOK, EventsResponse{Events: result.Events, Total: result.Total, Limited: result.Limited})
	})
}

func traceIDFromPath(path string) (string, bool) {
	parts := strings.Split(path, "/")
	if len(parts) != 4 || parts[1] != "v1" || parts[2] != "traces" {
		return "", false
	}
	return parts[3], true
}

func validUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return false
	}
	compact := strings.ReplaceAll(value, "-", "")
	_, err := hex.DecodeString(compact)
	return err == nil
}


