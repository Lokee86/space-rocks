package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/serviceidentity"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage"
)

type statusStore struct {
	status storage.Status
	err    error
}

func (s statusStore) Status(context.Context) (storage.Status, error) { return s.status, s.err }

func TestHandlerLiveReadyAndStatus(t *testing.T) {
	state := NewState("i", "v", "test", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	store := statusStore{status: storage.Status{Ready: true, RecordCount: 4, ByteCount: 20}}
	handler := NewHandler(state, store)
	check := func(path string, want int) string {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Code != want {
			t.Fatalf("%s status = %d, want %d", path, recorder.Code, want)
		}
		return recorder.Body.String()
	}
	liveBody := check("/live", http.StatusOK)
	if !strings.Contains(liveBody, `"service":"`+serviceidentity.ServiceName+`"`) || !strings.Contains(liveBody, `"status":"live"`) {
		t.Fatal("live response missing status")
	}
	if check("/ready", http.StatusServiceUnavailable) == "" {
		t.Fatal("empty not-ready response")
	}
	state.MarkReady()
	readyBody := check("/ready", http.StatusOK)
	if !strings.Contains(readyBody, `"service":"`+serviceidentity.ServiceName+`"`) || !strings.Contains(readyBody, `"status":"ready"`) {
		t.Fatal("ready response missing status")
	}
	body := check("/status", http.StatusOK)
	if !strings.Contains(body, `"service":"`+serviceidentity.ServiceName+`"`) || !strings.Contains(body, `"events_accepted":0`) || !strings.Contains(body, `"record_count":4`) {
		t.Fatalf("unexpected status response: %s", body)
	}
}

func TestHandlerHidesStorageErrorsAndGuardsMethods(t *testing.T) {
	state := NewState("i", "v", "test", time.Time{})
	handler := NewHandler(state, statusStore{err: errors.New("secret backend failure")})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/live", nil))
	if recorder.Code != http.StatusMethodNotAllowed || recorder.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("unexpected method response: %d %q", recorder.Code, recorder.Header().Get("Allow"))
	}
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/status", nil))
	if recorder.Code != http.StatusOK || strings.Contains(recorder.Body.String(), "secret backend failure") {
		t.Fatalf("storage error leaked: %s", recorder.Body.String())
	}
}
