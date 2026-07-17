package rooms

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/Lokee86/space-rocks/shared/go/servicelog"
	"github.com/google/uuid"
)

func TestRoomCleanupObservability(t *testing.T) {
	path := configureCleanupObservability(t)
	manager := NewRoomManagerWithCleanupDelay(0)
	room := NewRoom("observability-cleanup-room", RoomStateLobby, nil)

	manager.mu.Lock()
	manager.rooms[room.ID] = room
	manager.mu.Unlock()
	room.mu.Lock()
	cleanupVersion := room.cleanup.IncrementVersion()
	room.mu.Unlock()

	previousStopRoomForCleanup := stopRoomForCleanup
	stopRoomForCleanup = func(*Room) {}
	t.Cleanup(func() { stopRoomForCleanup = previousStopRoomForCleanup })
	manager.cleanupEmptyRoom(room.ID, cleanupVersion)

	if err := logging.CloseFileOutput(); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(payload)), "\n")
	if len(lines) != 1 {
		t.Fatalf("JSONL records = %d, want 1", len(lines))
	}
	var record map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &record); err != nil {
		t.Fatal(err)
	}
	if record["event"] != string(observability.EventNameRoomCleanedUp) || record["room_id"] != room.ID {
		t.Fatalf("cleanup record = %#v", record)
	}
	traceID, ok := record["trace_id"].(string)
	if !ok || uuid.Validate(traceID) != nil {
		t.Fatalf("trace_id = %#v, want valid UUID", record["trace_id"])
	}
	fields, ok := record["fields"].(map[string]any)
	if !ok {
		t.Fatalf("fields = %#v, want nested fields object", record["fields"])
	}
	if fields["reason_code"] != "empty_room_cleanup" || fields["cleanup_version"] != float64(cleanupVersion) {
		t.Fatalf("cleanup fields = %#v", fields)
	}
}

func configureCleanupObservability(t *testing.T) string {
	t.Helper()
	t.Cleanup(func() { _ = logging.CloseFileOutput() })
	if err := logging.CloseFileOutput(); err != nil {
		t.Fatal(err)
	}
	if err := logging.ConfigureRuntime(servicelog.ServiceIdentity{
		Name: logging.ServiceName, Version: "test-build", Environment: "test",
		InstanceID: "550e8400-e29b-41d4-a716-446655440020",
	}); err != nil {
		t.Fatal(err)
	}
	path, err := logging.ConfigureFileOutput(t.TempDir(), "game-server-cleanup-test")
	if err != nil {
		t.Fatal(err)
	}
	return path
}
