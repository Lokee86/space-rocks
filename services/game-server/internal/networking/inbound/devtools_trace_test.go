package inbound

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/devtools"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/packetcodec"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
	"github.com/Lokee86/space-rocks/shared/go/servicelog"
)

func TestDebugCommandIngressPreservesTraceID(t *testing.T) {
	const traceID = "00000000-0000-4000-8000-000000000701"
	var command devtools.DebugCommand
	if err := packetcodec.Decode([]byte(`{"type":"debug_clear_bullets","trace_id":"`+traceID+`"}`), &command); err != nil {
		t.Fatalf("decode debug command: %v", err)
	}

	if command.TraceID != traceID {
		t.Fatalf("trace ID = %q, want %q", command.TraceID, traceID)
	}
}

func TestDebugCommandIngressPreservesTraceIDThroughAuthoritativeHandler(t *testing.T) {
	const traceID = "00000000-0000-4000-8000-000000000715"
	baseDir := filepath.Join(".go-test-tmp")
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := logging.ConfigureRuntime(servicelog.ServiceIdentity{
		Name: logging.ServiceName, Version: "test-build", Environment: "test",
		InstanceID: "550e8400-e29b-41d4-a716-446655440018",
	}); err != nil {
		t.Fatal(err)
	}
	logPath, err := logging.ConfigureFileOutput(baseDir, "devtools-trace-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = logging.CloseFileOutput()
		_ = os.RemoveAll(baseDir)
	})

	gameInstance := game.New()
	control := game.NewControl(gameInstance)
	if !control.EnsurePlayerSession("player-1", physics.Vector2{}) {
		t.Fatal("create player")
	}
	room := rooms.NewRoom("room", rooms.RoomStateInGame, gameInstance)
	session := &contextCountSession{context: SessionContext{
		Room: room, RoomID: room.ID, GamePlayerID: "player-1",
	}}
	msg := []byte(`{"type":"debug_set_score","trace_id":"` + traceID + `","target_scope":"single_player","target_player_id":"player-1","score":17}`)

	if !HandleSimpleDevtoolsPacket(session, "remote", msg, ClientPacketEnvelope{Type: devtools.PacketTypeDebugSetScore}) {
		t.Fatal("expected devtools packet to be handled")
	}
	if score := gameInstance.GameplayPresentationSnapshot("player-1").PlayerSessions["player-1"].Score; score != 17 {
		t.Fatalf("score = %d, want 17", score)
	}
	if err := logging.CloseFileOutput(); err != nil {
		t.Fatal(err)
	}

	payload, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(payload)), "\n") {
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatal(err)
		}
		if record["event"] == "devtools_command_applied" {
			if record["trace_id"] != traceID {
				t.Fatalf("applied trace ID = %v, want %s", record["trace_id"], traceID)
			}
			return
		}
	}
	t.Fatal("authoritative devtools applied event not found")
}
