package rooms

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/Lokee86/space-rocks/shared/go/servicelog"
	"github.com/google/uuid"
)

func TestRoomLifecycleObservabilitySuccessChain(t *testing.T) {
	path := configureRoomSuccessObservability(t)
	manager := NewRoomManagerWithCleanupDelay(0)
	room, roomErr := manager.CreateStartedSinglePlayerRoom("session-owner")
	if roomErr != nil {
		t.Fatalf("start single-player match: %v", roomErr)
	}
	t.Cleanup(manager.StopAll)

	matchID := room.CurrentMatchID()
	traceID := room.CurrentMatchTraceID()
	if traceID == "" {
		t.Fatal("expected an existing room match trace")
	}
	markLifecycleTickTestGameOver(t, room.GameInstance())
	if !TickRoomGameOverLifecycle(room, func(*Room) {}) {
		t.Fatal("expected game-over lifecycle to advance")
	}
	if !ReportResolvedMatchResultOnce(room, &fakeMatchResultReporter{}) {
		t.Fatal("expected match result report to succeed")
	}

	records, _ := readRoomSuccessObservabilityJSONL(t, path)
	wantEvents := []string{
		string(observability.EventNameGameOverDetected),
		string(observability.EventNameMatchResultReportStarted),
		string(observability.EventNameMatchResultReportSucceeded),
	}
	if len(records) != len(wantEvents) {
		t.Fatalf("JSONL records = %d, want %d: %#v", len(records), len(wantEvents), records)
	}
	for index, record := range records {
		if record["event"] != wantEvents[index] {
			t.Fatalf("record %d event = %#v, want %q", index, record["event"], wantEvents[index])
		}
		if record["trace_id"] != traceID || record["room_id"] != room.ID || record["match_id"] != matchID {
			t.Fatalf("record %d context = %#v, want trace=%q room=%q match=%q", index, record, traceID, room.ID, matchID)
		}
	}

	assertRoomSuccessObservabilityFields(t, records[0], map[string]any{
		"reason_code": "simulation_complete",
	})
	assertRoomSuccessObservabilityFields(t, records[1], map[string]any{
		"reason_code":  "game_over",
		"mode":         "single_player",
		"player_count": float64(1),
	})
	assertRoomSuccessObservabilityFields(t, records[2], map[string]any{
		"reason_code":  "game_over",
		"player_count": float64(1),
	})
}

func TestRoomLifecycleObservabilityFailureChain(t *testing.T) {
	path := configureRoomSuccessObservability(t)
	manager := NewRoomManagerWithCleanupDelay(0)
	room, roomErr := manager.CreateStartedSinglePlayerRoom("session-owner")
	if roomErr != nil {
		t.Fatalf("start single-player match: %v", roomErr)
	}
	t.Cleanup(manager.StopAll)
	matchID := room.CurrentMatchID()
	traceID := room.CurrentMatchTraceID()
	markLifecycleTickTestGameOver(t, room.GameInstance())
	if !TickRoomGameOverLifecycle(room, func(*Room) {}) {
		t.Fatal("expected game-over lifecycle to advance")
	}
	sentinel := errors.New("sentinel raw reporter error")
	if ReportResolvedMatchResultOnce(room, &fakeMatchResultReporter{err: sentinel}) {
		t.Fatal("expected match result report to fail")
	}

	records, raw := readRoomSuccessObservabilityJSONL(t, path)
	wantEvents := []string{
		string(observability.EventNameGameOverDetected),
		string(observability.EventNameMatchResultReportStarted),
		string(observability.EventNameMatchResultReportFailed),
	}
	if len(records) != len(wantEvents) {
		t.Fatalf("JSONL records = %d, want %d: %#v", len(records), len(wantEvents), records)
	}
	for index, record := range records {
		if record["event"] != wantEvents[index] || record["trace_id"] != traceID || record["room_id"] != room.ID || record["match_id"] != matchID {
			t.Fatalf("record %d = %#v, want event/context %q/%q/%q/%q", index, record, wantEvents[index], traceID, room.ID, matchID)
		}
	}
	assertRoomSuccessObservabilityFields(t, records[2], map[string]any{
		"reason_code":  "game_over",
		"failure_mode": "report_failed",
	})
	if strings.Contains(raw, sentinel.Error()) {
		t.Fatalf("sentinel raw error leaked into JSONL: %s", raw)
	}
}

func TestRoomCurrentOrCreateMatchTraceIDIsStable(t *testing.T) {
	room := NewRoom("observability-trace-room", RoomStateLobby, nil)
	first := room.CurrentOrCreateMatchTraceID()
	if err := uuid.Validate(first); err != nil {
		t.Fatalf("first trace ID = %q, want valid UUID: %v", first, err)
	}
	if second := room.CurrentOrCreateMatchTraceID(); second != first {
		t.Fatalf("second trace ID = %q, want %q", second, first)
	}
	if stored := room.CurrentMatchTraceID(); stored != first {
		t.Fatalf("stored trace ID = %q, want %q", stored, first)
	}
}

func configureRoomSuccessObservability(t *testing.T) string {
	t.Helper()
	t.Cleanup(func() { _ = logging.CloseFileOutput() })
	if err := logging.CloseFileOutput(); err != nil {
		t.Fatal(err)
	}
	if err := logging.ConfigureRuntime(servicelog.ServiceIdentity{
		Name: logging.ServiceName, Version: "test-build", Environment: "test",
		InstanceID: "550e8400-e29b-41d4-a716-446655440019",
	}); err != nil {
		t.Fatal(err)
	}
	path, err := logging.ConfigureFileOutput(t.TempDir(), "game-server-room-test")
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func readRoomSuccessObservabilityJSONL(t *testing.T, path string) ([]map[string]any, string) {
	t.Helper()
	if err := logging.CloseFileOutput(); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(payload)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, string(payload)
	}
	records := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("decode JSONL record: %v", err)
		}
		records = append(records, record)
	}
	return records, string(payload)
}

func assertRoomSuccessObservabilityFields(t *testing.T, record map[string]any, want map[string]any) {
	t.Helper()
	fields, ok := record["fields"].(map[string]any)
	if !ok {
		t.Fatalf("record fields = %#v, want nested fields object", record["fields"])
	}
	for key, value := range want {
		if fields[key] != value {
			t.Fatalf("record fields[%q] = %#v, want %#v", key, fields[key], value)
		}
	}
}
