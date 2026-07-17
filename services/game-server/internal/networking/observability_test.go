package networking

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/Lokee86/space-rocks/shared/go/servicelog"
)

const (
	testNetworkingTraceID      = "550e8400-e29b-41d4-a716-446655440018"
	testNetworkingQueueTraceID = "550e8400-e29b-41d4-a716-446655440019"
)

func captureNetworkingJSONL(t *testing.T, emit func()) []map[string]any {
	t.Helper()

	if err := logging.CloseFileOutput(); err != nil {
		t.Fatalf("CloseFileOutput() error = %v", err)
	}
	if err := logging.ConfigureRuntime(servicelog.ServiceIdentity{
		Name:        logging.ServiceName,
		Version:     "test-build",
		Environment: "test",
		InstanceID:  "550e8400-e29b-41d4-a716-446655440017",
	}); err != nil {
		t.Fatalf("ConfigureRuntime() error = %v", err)
	}
	path, err := logging.ConfigureFileOutput(t.TempDir(), "networking-observability-test")
	if err != nil {
		t.Fatalf("ConfigureFileOutput() error = %v", err)
	}
	t.Cleanup(func() { _ = logging.CloseFileOutput() })

	emit()
	if err := logging.CloseFileOutput(); err != nil {
		t.Fatalf("CloseFileOutput() error = %v", err)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	content := strings.TrimSpace(string(payload))
	if content == "" {
		return nil
	}

	lines := strings.Split(content, "\n")
	records := make([]map[string]any, 0, len(lines))
	for index, line := range lines {
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("decode JSONL record %d: %v", index, err)
		}
		records = append(records, record)
	}
	return records
}

func TestWebSocketCloseFailuresWriteCanonicalJSONL(t *testing.T) {
	records := captureNetworkingJSONL(t, func() {
		logWebSocketReadClose(errors.New("raw read failure"), testNetworkingTraceID, "session-read-write", "room-read-write", "Player-1")
		logWebSocketWriteClose(errors.New("raw write failure"), testNetworkingTraceID, "session-read-write", "room-read-write", "Player-1")
	})

	if len(records) != 2 {
		t.Fatalf("JSONL records = %d, want 2", len(records))
	}
	wantEvents := []string{
		string(observability.EventNameGameServerReadFailed),
		string(observability.EventNameGameServerWriteFailed),
	}
	for index, record := range records {
		if record["event"] != wantEvents[index] {
			t.Fatalf("record %d event = %v, want %q", index, record["event"], wantEvents[index])
		}
		for key, want := range map[string]string{
			"trace_id":   testNetworkingTraceID,
			"session_id": "session-read-write",
			"room_id":    "room-read-write",
			"player_id":  "Player-1",
		} {
			if record[key] != want {
				t.Fatalf("record %d %s = %v, want %q", index, key, record[key], want)
			}
		}
		fields, ok := record["fields"].(map[string]any)
		if !ok {
			t.Fatalf("record %d fields = %#v, want object", index, record["fields"])
		}
		wantCode := "websocket_read_failed"
		if index == 1 {
			wantCode = "websocket_write_failed"
		}
		if fields["error_code"] != wantCode || fields["failure_mode"] != wantCode {
			t.Fatalf("record %d failure fields = %#v, want %q", index, fields, wantCode)
		}
		serialized, err := json.Marshal(record)
		if err != nil {
			t.Fatalf("marshal record %d: %v", index, err)
		}
		if strings.Contains(string(serialized), "raw read failure") || strings.Contains(string(serialized), "raw write failure") {
			t.Fatalf("record %d leaked raw error text: %s", index, serialized)
		}
	}
}

func TestWebSocketOutboundQueueOverflowWritesOnceCanonicalJSONL(t *testing.T) {
	records := captureNetworkingJSONL(t, func() {
		session := &webSocketSession{
			connectionTraceID: testNetworkingQueueTraceID,
			sessionID:         "session-queue-overflow",
			context: SessionContext{
				RoomID:       "room-queue-overflow",
				GamePlayerID: "Player-2",
			},
			outbound: make(chan []byte, 1),
		}
		session.outbound <- []byte("occupied")
		session.enqueue([]byte("dropped-1"))
		session.enqueue([]byte("dropped-2"))
	})

	if len(records) != 1 {
		t.Fatalf("JSONL records = %d, want one queue-overflow event", len(records))
	}
	record := records[0]
	if record["event"] != string(observability.EventNameGameServerClientDisconnected) {
		t.Fatalf("event = %v, want %q", record["event"], observability.EventNameGameServerClientDisconnected)
	}
	for key, want := range map[string]string{
		"trace_id":   testNetworkingQueueTraceID,
		"session_id": "session-queue-overflow",
		"room_id":    "room-queue-overflow",
		"player_id":  "Player-2",
	} {
		if record[key] != want {
			t.Fatalf("%s = %v, want %q", key, record[key], want)
		}
	}
	fields, ok := record["fields"].(map[string]any)
	if !ok {
		t.Fatalf("fields = %#v, want object", record["fields"])
	}
	for key, want := range map[string]string{
		"reason_code":  "outbound_queue_full",
		"failure_mode": "outbound_queue_full",
	} {
		if fields[key] != want {
			t.Fatalf("fields[%s] = %v, want %q", key, fields[key], want)
		}
	}
}
