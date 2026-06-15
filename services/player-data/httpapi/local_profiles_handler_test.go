package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/playerdata"
)

func TestLocalProfilesHandlerReturnsUnavailableWhenLocalProfileStoreMissing(t *testing.T) {
	runtime, err := playerdata.NewConfiguredRuntime(playerdata.RuntimeConfig{})
	if err != nil {
		t.Fatalf("NewConfiguredRuntime returned error: %v", err)
	}

	handler := NewLocalProfilesHandler(runtime)

	tests := []struct {
		name       string
		method     string
		target     string
		body       string
		wantStatus int
	}{
		{
			name:       "list",
			method:     http.MethodGet,
			target:     "/api/player-data/local-profiles",
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "default get",
			method:     http.MethodGet,
			target:     "/api/player-data/local-profiles/default",
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "create",
			method:     http.MethodPost,
			target:     "/api/player-data/local-profiles",
			body:       `{"display_name":"Pilot_1"}`,
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "default put",
			method:     http.MethodPut,
			target:     "/api/player-data/local-profiles/default",
			body:       `{"identity_kind":"guest","local_profile_id":""}`,
			wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.target, nil)
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, tc.target, strings.NewReader(tc.body))
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}

			var payload map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}
			if payload["error"] != "local_profiles_unavailable" {
				t.Fatalf("error = %q, want %q", payload["error"], "local_profiles_unavailable")
			}
		})
	}
}
