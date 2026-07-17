package networking

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
	"github.com/Lokee86/space-rocks/shared/go/servicelog"
	"github.com/google/uuid"
)

func configureOutboundObservabilityLogging(t *testing.T) func() []map[string]any {
	t.Helper()
	if err := logging.CloseFileOutput(); err != nil {
		t.Fatalf("CloseFileOutput() error = %v", err)
	}
	if err := logging.ConfigureRuntime(servicelog.ServiceIdentity{
		Name:        logging.ServiceName,
		Version:     "test-build",
		Environment: "test",
		InstanceID:  "550e8400-e29b-41d4-a716-446655440000",
	}); err != nil {
		t.Fatalf("ConfigureRuntime() error = %v", err)
	}
	logging.Configure("warn")
	logPath, err := logging.ConfigureFileOutput(t.TempDir(), "outbound-observability-test")
	if err != nil {
		t.Fatalf("ConfigureFileOutput() error = %v", err)
	}

	closed := false
	t.Cleanup(func() {
		if !closed {
			if err := logging.CloseFileOutput(); err != nil {
				t.Errorf("CloseFileOutput() cleanup error = %v", err)
			}
		}
		logging.Configure("warn")
	})

	return func() []map[string]any {
		t.Helper()
		if !closed {
			if err := logging.CloseFileOutput(); err != nil {
				t.Fatalf("CloseFileOutput() error = %v", err)
			}
			closed = true
		}
		payload, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", logPath, err)
		}
		payload = bytes.TrimSpace(payload)
		if len(payload) == 0 {
			return nil
		}

		lines := bytes.Split(payload, []byte{'\n'})
		records := make([]map[string]any, 0, len(lines))
		for _, line := range lines {
			var record map[string]any
			if err := json.Unmarshal(line, &record); err != nil {
				t.Fatalf("Unmarshal JSONL record %q: %v", line, err)
			}
			records = append(records, record)
		}
		return records
	}
}

func assertOutboundEventRecord(t *testing.T, records []map[string]any, event string, traceID string, sessionID string, roomID string, playerID string, failureMode string) {
	t.Helper()
	if len(records) != 1 {
		t.Fatalf("JSONL records = %d, want 1: %#v", len(records), records)
	}
	record := records[0]
	for field, want := range map[string]string{
		"event":      event,
		"trace_id":   traceID,
		"session_id": sessionID,
		"room_id":    roomID,
		"player_id":  playerID,
	} {
		if got, ok := record[field].(string); !ok || got != want {
			t.Fatalf("record[%q] = %#v, want %q", field, record[field], want)
		}
	}
	fields, ok := record["fields"].(map[string]any)
	if !ok {
		t.Fatalf("record fields = %#v, want object", record["fields"])
	}
	if got, ok := fields["failure_mode"].(string); !ok || got != failureMode {
		t.Fatalf("fields.failure_mode = %#v, want %q", fields["failure_mode"], failureMode)
	}
}

func TestNewWebSocketSessionOwnsConnectionTraceID(t *testing.T) {
	roomManager := rooms.NewRoomManager()
	t.Cleanup(roomManager.StopAll)

	session := newWebSocketSession(nil, roomManager, nil, nil)
	if session.connectionTraceID == "" {
		t.Fatal("expected session trace ID to be non-empty")
	}
	if _, err := uuid.Parse(session.connectionTraceID); err != nil {
		t.Fatalf("expected session trace ID to be a UUID, got %q: %v", session.connectionTraceID, err)
	}
}

func TestEnqueuePacketEmitsCanonicalEncodeFailure(t *testing.T) {
	readRecords := configureOutboundObservabilityLogging(t)
	roomManager := rooms.NewRoomManager()
	t.Cleanup(roomManager.StopAll)
	traceID := "550e8400-e29b-41d4-a716-446655440001"
	session := &webSocketSession{
		sessionID:         "session-encode-test",
		connectionTraceID: traceID,
		context: SessionContext{
			RoomID:       "room-encode-test",
			GamePlayerID: "player-encode-test",
		},
		rooms: roomManager,
	}

	session.enqueuePacket(map[string]any{"unsupported": func() {}})

	assertOutboundEventRecord(t, readRecords(), "outbound_packet_encode_failed", traceID, "session-encode-test", "room-encode-test", "player-encode-test", "websocket_control_packet_encode_failed")
}

func TestEnqueueOverflowEmitsCanonicalDisconnectEvent(t *testing.T) {
	readRecords := configureOutboundObservabilityLogging(t)
	roomManager := rooms.NewRoomManager()
	t.Cleanup(roomManager.StopAll)
	traceID := "550e8400-e29b-41d4-a716-446655440002"
	outbound := make(chan []byte, 1)
	outbound <- []byte("prefilled")
	session := &webSocketSession{
		sessionID:         "session-overflow-test",
		connectionTraceID: traceID,
		context: SessionContext{
			RoomID:       "room-overflow-test",
			GamePlayerID: "player-overflow-test",
		},
		rooms:    roomManager,
		outbound: outbound,
	}

	session.enqueue([]byte("overflow"))

	assertOutboundEventRecord(t, readRecords(), "game_server_client_disconnected", traceID, "session-overflow-test", "room-overflow-test", "player-overflow-test", "outbound_queue_full")
}
