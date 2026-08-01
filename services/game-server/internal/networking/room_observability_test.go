package networking

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/Lokee86/space-rocks/shared/go/servicelog"
)

func TestHandleStartSinglePlayerRequestEmitsRoomCreated(t *testing.T) {
	path := configureRoomCreatedObservability(t)
	manager := NewRoomManagerWithCleanupDelay(0)
	t.Cleanup(manager.StopAll)
	session := &webSocketSession{
		sessionID: "session-room-created",
		rooms:     manager,
		outbound:  make(chan []byte, 16),
	}
	traceID := "550e8400-e29b-41d4-a716-446655440021"
	session.handleStartSinglePlayerRequest("", traceID, "", 0, false, 0, 0)
	roomID := session.sessionContext().RoomID
	if roomID == "" {
		t.Fatal("expected single-player room to be created")
	}

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
	if record["event"] != string(observability.EventNameRoomCreated) || record["trace_id"] != traceID || record["session_id"] != session.sessionID || record["room_id"] != roomID {
		t.Fatalf("room-created record = %#v", record)
	}
	fields, ok := record["fields"].(map[string]any)
	if !ok || fields["reason_code"] != "single_player_created" {
		t.Fatalf("room-created fields = %#v", record["fields"])
	}
}

func configureRoomCreatedObservability(t *testing.T) string {
	t.Helper()
	t.Cleanup(func() { _ = logging.CloseFileOutput() })
	if err := logging.CloseFileOutput(); err != nil {
		t.Fatal(err)
	}
	if err := logging.ConfigureRuntime(servicelog.ServiceIdentity{
		Name: logging.ServiceName, Version: "test-build", Environment: "test",
		InstanceID: "550e8400-e29b-41d4-a716-446655440022",
	}); err != nil {
		t.Fatal(err)
	}
	path, err := logging.ConfigureFileOutput(t.TempDir(), "game-server-room-created-test")
	if err != nil {
		t.Fatal(err)
	}
	return path
}
