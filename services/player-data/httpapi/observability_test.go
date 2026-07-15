package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/logging"
	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/player-data/protocol"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/Lokee86/space-rocks/shared/go/servicelog"
)

type canonicalLocalProfileStore struct {
	*localProfilesHandlerTestStore
}

func (s *canonicalLocalProfileStore) CreateLocalProfile(string, string, protocol.PlayerDataStats) (playerdata.LocalProfileSummary, error) {
	return playerdata.LocalProfileSummary{}, errors.New("store body must not be logged")
}

func (s *canonicalLocalProfileStore) ListLocalProfiles() ([]playerdata.LocalProfileSummary, error) {
	return nil, playerdata.ErrLocalProfileUnavailable
}

func (s *canonicalLocalProfileStore) DeleteLocalProfile(string) error {
	return errors.New("write body must not be logged")
}

func captureHTTPPlayerDataEvents(t *testing.T, fn func()) ([]map[string]any, observability.Status) {
	t.Helper()
	if err := logging.ConfigureRuntime(servicelog.ServiceIdentity{
		Name: logging.ServiceName, Version: "test-build", Environment: "test",
		InstanceID: "550e8400-e29b-41d4-a716-446655440014",
	}); err != nil {
		t.Fatal(err)
	}
	path, err := logging.ConfigureFileOutput(t.TempDir(), "player-data-http-test")
	if err != nil {
		t.Fatal(err)
	}

	fn()
	status := logging.EventStatus()
	if err := logging.CloseFileOutput(); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var events []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(string(payload)), "\n") {
		if line == "" {
			continue
		}
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatal(err)
		}
		events = append(events, event)
	}
	return events, status
}

func TestLocalProfileCanonicalFailureEventsAreAcceptedAndJSONL(t *testing.T) {
	events, status := captureHTTPPlayerDataEvents(t, func() {
		store := &canonicalLocalProfileStore{localProfilesHandlerTestStore: newLocalProfilesHandlerTestStore()}
		runtime, err := playerdata.NewRuntime(playerdata.Config{Store: store})
		if err != nil {
			t.Fatal(err)
		}
		handler := NewLocalProfilesHandler(runtime)

		create := httptest.NewRequest(http.MethodPost, "/api/player-data/local-profiles", strings.NewReader(`{"display_name":"Pilot"}`))
		createRecorder := httptest.NewRecorder()
		handler.ServeHTTP(createRecorder, create)
		if createRecorder.Code != http.StatusServiceUnavailable {
			t.Fatalf("create status = %d", createRecorder.Code)
		}

		listRecorder := httptest.NewRecorder()
		handler.ServeHTTP(listRecorder, httptest.NewRequest(http.MethodGet, "/api/player-data/local-profiles", nil))
		if listRecorder.Code != http.StatusServiceUnavailable {
			t.Fatalf("list status = %d", listRecorder.Code)
		}

		deleteRecorder := httptest.NewRecorder()
		deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/player-data/local-profiles/local-1", nil)
		deleteRequest.SetPathValue("local_profile_id", "local-1")
		handler.ServeHTTP(deleteRecorder, deleteRequest)
		if deleteRecorder.Code != http.StatusServiceUnavailable {
			t.Fatalf("delete status = %d", deleteRecorder.Code)
		}
	})
	if status.RejectedCount != 0 || status.AcceptedCount != 3 {
		t.Fatalf("emitter status = %+v, want three accepted events and no rejection", status)
	}
	wantEvents := []string{
		string(observability.EventNameLocalProfileCreateFailed),
		string(observability.EventNamePlayerDataReadFailed),
		string(observability.EventNamePlayerDataWriteFailed),
	}
	if len(events) != len(wantEvents) {
		t.Fatalf("events = %#v", events)
	}
	for index, event := range events {
		if event["event"] != wantEvents[index] || event["trace_id"] == "" || event["request_id"] == "" {
			t.Fatalf("event context = %#v", event)
		}
		if strings.Contains(string(mustJSONHTTP(t, event)), "must not be logged") {
			t.Fatalf("raw error leaked into event = %#v", event)
		}
	}
	createFields, ok := events[0]["fields"].(map[string]any)
	if !ok || createFields["failure_mode"] != "store_write" || createFields["local_profile_id"] == "" {
		t.Fatalf("create event fields = %#v", events[0])
	}
}

func mustJSONHTTP(t *testing.T, value any) []byte {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}
